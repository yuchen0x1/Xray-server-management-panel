package sub

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	subLimitMaxMisses  = 30
	subLimitWindow     = 5 * time.Minute
	subLimitCooldown   = 15 * time.Minute
	subLimitMaxRecords = 10000
)

// subLimitRecord tracks the 404 misses of one client IP inside a fixed window.
// A blockedUntil in the future means the IP is currently cooling down after
// too many misses.
type subLimitRecord struct {
	windowStart  time.Time
	misses       int
	blockedUntil time.Time
}

// subLimiter throttles subscription-endpoint enumeration. Every 404 the sub
// server returns (wrong subId guesses, path probes) counts as a miss; an IP
// exceeding maxMisses inside the window is blocked for the cooldown. Valid
// requests answer 200 and never count, so NAT users sharing one IP are only
// affected when someone behind it is actively guessing.
type subLimiter struct {
	mu        sync.Mutex
	now       func() time.Time
	maxMisses int
	window    time.Duration
	cooldown  time.Duration
	records   map[string]*subLimitRecord
}

func newSubLimiter(maxMisses int, window, cooldown time.Duration) *subLimiter {
	return &subLimiter{
		now:       time.Now,
		maxMisses: maxMisses,
		window:    window,
		cooldown:  cooldown,
		records:   make(map[string]*subLimitRecord),
	}
}

func (l *subLimiter) blocked(ip string) (time.Time, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	record, ok := l.records[ip]
	if !ok || record.blockedUntil.IsZero() {
		return time.Time{}, false
	}
	if l.now().Before(record.blockedUntil) {
		return record.blockedUntil, true
	}
	return time.Time{}, false
}

func (l *subLimiter) registerMiss(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	record, ok := l.records[ip]
	if !ok {
		if len(l.records) >= subLimitMaxRecords {
			l.evictExpiredLocked(now)
			if len(l.records) >= subLimitMaxRecords {
				return
			}
		}
		record = &subLimitRecord{windowStart: now}
		l.records[ip] = record
	}
	if !record.blockedUntil.IsZero() && now.Before(record.blockedUntil) {
		return
	}
	record.blockedUntil = time.Time{}
	if now.Sub(record.windowStart) >= l.window {
		record.windowStart = now
		record.misses = 0
	}
	record.misses++
	if record.misses >= l.maxMisses {
		record.blockedUntil = now.Add(l.cooldown)
		record.misses = 0
		record.windowStart = now
	}
}

func (l *subLimiter) evictExpiredLocked(now time.Time) {
	for ip, record := range l.records {
		blockedActive := !record.blockedUntil.IsZero() && now.Before(record.blockedUntil)
		windowActive := now.Sub(record.windowStart) < l.window
		if !blockedActive && !windowActive {
			delete(l.records, ip)
		}
	}
}

// rateLimitMiddleware blocks cooldown IPs up front and feeds every 404 the
// wrapped handlers produce back into the limiter.
func (a *SUBController) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if blockedUntil, blocked := a.limiter.blocked(ip); blocked {
			retryAfter := int(time.Until(blockedUntil).Seconds() + 0.5)
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
		if c.Writer.Status() == http.StatusNotFound {
			a.limiter.registerMiss(ip)
		}
	}
}

// parseTrustedProxyList splits the comma-separated trusted-proxy CIDR setting
// for gin's SetTrustedProxies. An empty result means no proxy is trusted and
// ClientIP falls back to the TCP peer address, so a direct client cannot
// forge its IP through X-Forwarded-For.
func parseTrustedProxyList(value string) []string {
	var list []string
	for part := range strings.SplitSeq(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			list = append(list, part)
		}
	}
	return list
}
