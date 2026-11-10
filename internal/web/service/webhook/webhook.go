package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/util/netsafe"
)

type Payload struct {
	Event     string    `json:"event"`
	Source    string    `json:"source,omitempty"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
	Summary   string    `json:"summary"`
	Data      any       `json:"data,omitempty"`
}

const (
	// defaultCPUThreshold and defaultMemoryThreshold match the panel's own
	// settings defaults, so a webhook-only deployment alerts at the same level
	// the UI would suggest rather than every minute on an idle host.
	defaultCPUThreshold    = 90.0
	defaultMemoryThreshold = 90.0
	// maxConcurrentDispatches bounds in-flight outbound requests. Every event
	// used to spawn an unbounded goroutine holding a socket; login attempts in
	// particular reach this without any authentication.
	maxConcurrentDispatches = 16
)

func envThreshold(name string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || value <= 0 || value > 100 {
		logger.Warning(fmt.Sprintf("webhook: ignoring invalid %s=%q", name, raw))
		return fallback
	}
	return value
}

type Subscriber struct {
	webhookURL    string
	secret        string
	client        *http.Client
	limiter       *eventbus.RateLimiter
	metrics       metricThresholds
	dispatchSlots chan struct{}
}

// metricThresholds is this subscriber's own alerting threshold for the sampled
// system metrics. The cpu.high and memory.high events carry a raw percentage on
// every tick, so each notifier decides for itself whether the reading is worth
// reporting; sharing another notifier's threshold would either spam the webhook
// on a healthy host or drop real alerts.
type metricThresholds struct {
	cpuPercent    float64
	memoryPercent float64
}

func NewSubscriber(webhookURL, secret string) *Subscriber {
	transport := &http.Transport{
		DialContext:         netsafe.SSRFGuardedDialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &Subscriber{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     strings.TrimSpace(secret),
		metrics: metricThresholds{
			cpuPercent:    envThreshold("XUI_WEBHOOK_CPU_THRESHOLD", defaultCPUThreshold),
			memoryPercent: envThreshold("XUI_WEBHOOK_MEMORY_THRESHOLD", defaultMemoryThreshold),
		},
		dispatchSlots: make(chan struct{}, maxConcurrentDispatches),
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 0 && !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
					return fmt.Errorf("webhook: refusing redirect to another host: %s", req.URL.Host)
				}
				if len(via) >= 3 {
					return fmt.Errorf("webhook: stopped after %d redirects", len(via))
				}
				return nil
			},
		},
		limiter: eventbus.NewRateLimiter(1 * time.Minute),
	}
}

func (s *Subscriber) HandleEvent(e eventbus.Event) {
	if s.webhookURL == "" {
		return
	}

	if !s.wants(e) {
		return
	}

	if e.Type != eventbus.EventLoginAttempt {
		if !s.limiter.Allow(e.Type, e.Source) {
			return
		}
	}

	summary := formatSummary(e, s.metrics)
	host, _ := os.Hostname()

	payload := Payload{
		Event:     string(e.Type),
		Source:    e.Source,
		Host:      host,
		Timestamp: e.Timestamp,
		Summary:   summary,
		Data:      e.Data,
	}

	select {
	case s.dispatchSlots <- struct{}{}:
		go func() {
			defer func() { <-s.dispatchSlots }()
			s.dispatch(payload)
		}()
	default:
		logger.Debug("webhook: dropping event, too many deliveries in flight:", e.Type)
	}
}

// wants decides whether this subscriber should report the event at all. The
// sampled metric events fire on every tick regardless of load, so they are
// gated here against this subscriber's own configured threshold.
func (s *Subscriber) wants(e eventbus.Event) bool {
	data, ok := e.Data.(*eventbus.SystemMetricData)
	if !ok {
		return true
	}
	switch e.Type {
	case eventbus.EventCPUHigh:
		return data.Percent > s.metrics.cpuPercent
	case eventbus.EventMemoryHigh:
		return data.Percent > s.metrics.memoryPercent
	default:
		return true
	}
}

func formatSummary(e eventbus.Event, thresholds metricThresholds) string {
	switch e.Type {
	case eventbus.EventOutboundDown:
		return fmt.Sprintf("Outbound [%s] is DOWN", e.Source)
	case eventbus.EventOutboundUp:
		return fmt.Sprintf("Outbound [%s] is UP", e.Source)
	case eventbus.EventXrayCrash:
		return fmt.Sprintf("Xray core crashed: %v", e.Data)
	case eventbus.EventNodeDown:
		return fmt.Sprintf("Node [%s] is DOWN", e.Source)
	case eventbus.EventNodeUp:
		return fmt.Sprintf("Node [%s] is UP", e.Source)
	case eventbus.EventCPUHigh:
		if data, ok := e.Data.(*eventbus.SystemMetricData); ok {
			return fmt.Sprintf("CPU usage high: %.1f%% (threshold: %.0f%%)", data.Percent, thresholds.cpuPercent)
		}
		return "CPU usage high"
	case eventbus.EventMemoryHigh:
		if data, ok := e.Data.(*eventbus.SystemMetricData); ok {
			return fmt.Sprintf("Memory usage high: %.1f%% (threshold: %.0f%%)", data.Percent, thresholds.memoryPercent)
		}
		return "Memory usage high"
	case eventbus.EventLoginAttempt:
		if data, ok := e.Data.(*eventbus.LoginEventData); ok {
			return fmt.Sprintf("Login %s for user %s from IP %s", data.Status, data.Username, data.IP)
		}
		return "Login attempt"
	default:
		return fmt.Sprintf("Event: %s (source: %s)", e.Type, e.Source)
	}
}

func (s *Subscriber) dispatch(payload Payload) {
	var bodyBytes []byte
	urlLower := strings.ToLower(s.webhookURL)

	if strings.Contains(urlLower, "dingtalk.com") {
		dingMsg := map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": fmt.Sprintf("[3x-ui] %s", payload.Summary),
				"text":  fmt.Sprintf("### 3x-ui 告警提醒\n- **主机**: %s\n- **事件**: %s\n- **摘要**: %s\n- **时间**: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(dingMsg)
	} else if strings.Contains(urlLower, "weixin.qq.com") {
		wechatMsg := map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": fmt.Sprintf("### 3x-ui 告警提醒\n> 主机: <font color=\"comment\">%s</font>\n> 事件: <font color=\"warning\">%s</font>\n> 摘要: %s\n> 时间: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(wechatMsg)
	} else if strings.Contains(urlLower, "feishu.cn") || strings.Contains(urlLower, "larksuite.com") {
		feishuMsg := map[string]any{
			"msg_type": "text",
			"content": map[string]string{
				"text": fmt.Sprintf("[3x-ui 告警]\n主机: %s\n事件: %s\n摘要: %s\n时间: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(feishuMsg)
	} else if strings.Contains(urlLower, "discord.com") {
		discordMsg := map[string]any{
			"content": fmt.Sprintf("**[3x-ui Alert]** %s\nHost: `%s` | Event: `%s` | Time: `%s`", payload.Summary, payload.Host, payload.Event, payload.Timestamp.Format("2006-01-02 15:04:05")),
		}
		bodyBytes, _ = json.Marshal(discordMsg)
	} else {
		bodyBytes, _ = json.Marshal(payload)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		logger.Warning("webhook: create request failed:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "3x-ui-Webhook/1.0")
	if s.secret != "" {
		req.Header.Set("X-Webhook-Secret", s.secret)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Warning("webhook: dispatch failed:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warning(fmt.Sprintf("webhook: dispatch returned non-2xx status: %d", resp.StatusCode))
	}
}
