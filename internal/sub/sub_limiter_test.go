package sub

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestSubLimiter(t *testing.T) (*subLimiter, *time.Time) {
	t.Helper()
	current := time.Now()
	nowRef := &current
	limiter := newSubLimiter(3, 5*time.Minute, 15*time.Minute)
	limiter.now = func() time.Time { return *nowRef }
	return limiter, nowRef
}

func TestSubLimiterBlocksAfterMaxMisses(t *testing.T) {
	limiter, _ := newTestSubLimiter(t)

	for i := 0; i < 2; i++ {
		limiter.registerMiss("1.2.3.4")
	}
	if _, blocked := limiter.blocked("1.2.3.4"); blocked {
		t.Fatal("IP must not be blocked below the miss threshold")
	}

	limiter.registerMiss("1.2.3.4")
	blockedUntil, blocked := limiter.blocked("1.2.3.4")
	if !blocked {
		t.Fatal("IP must be blocked after reaching the miss threshold")
	}
	if blockedUntil.IsZero() {
		t.Fatal("blockedUntil must be set for a blocked IP")
	}
}

func TestSubLimiterCooldownExpires(t *testing.T) {
	limiter, nowRef := newTestSubLimiter(t)

	for i := 0; i < 3; i++ {
		limiter.registerMiss("1.2.3.4")
	}
	if _, blocked := limiter.blocked("1.2.3.4"); !blocked {
		t.Fatal("IP must be blocked after reaching the miss threshold")
	}

	*nowRef = nowRef.Add(16 * time.Minute)
	if _, blocked := limiter.blocked("1.2.3.4"); blocked {
		t.Fatal("IP must be unblocked after the cooldown elapses")
	}
}

func TestSubLimiterWindowResetsMisses(t *testing.T) {
	limiter, nowRef := newTestSubLimiter(t)

	limiter.registerMiss("1.2.3.4")
	limiter.registerMiss("1.2.3.4")
	*nowRef = nowRef.Add(6 * time.Minute)
	limiter.registerMiss("1.2.3.4")

	if _, blocked := limiter.blocked("1.2.3.4"); blocked {
		t.Fatal("misses from a previous window must not count toward the current one")
	}
}

func TestSubLimiterTracksIPsIndependently(t *testing.T) {
	limiter, _ := newTestSubLimiter(t)

	for i := 0; i < 5; i++ {
		limiter.registerMiss("1.2.3.4")
	}
	if _, blocked := limiter.blocked("5.6.7.8"); blocked {
		t.Fatal("blocking one IP must not affect another")
	}
}

func TestSubLimiterCapsRecords(t *testing.T) {
	limiter := newSubLimiter(3, 5*time.Minute, 15*time.Minute)

	for i := 0; i < subLimitMaxRecords; i++ {
		limiter.registerMiss("10.0.0." + strconv.Itoa(i%256) + ":" + strconv.Itoa(i))
	}
	limiter.registerMiss("999.999.999.999")

	if len(limiter.records) > subLimitMaxRecords {
		t.Fatalf("record map must stay capped at %d, got %d", subLimitMaxRecords, len(limiter.records))
	}
}

func TestRateLimitMiddlewareReturns429WhenBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, _ := newTestSubLimiter(t)
	controller := &SUBController{limiter: limiter}
	router := gin.New()
	router.Use(controller.rateLimitMiddleware())
	router.GET("/sub/:subid", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sub/wrong-id", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 during warm-up, got %d", w.Code)
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sub/wrong-id", nil))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after the miss threshold, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("expected a Retry-After header on the 429 response")
	}
}

func TestRateLimitMiddlewareIgnoresSuccessfulRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter, _ := newTestSubLimiter(t)
	controller := &SUBController{limiter: limiter}
	router := gin.New()
	router.Use(controller.rateLimitMiddleware())
	router.GET("/sub/:subid", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 50; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sub/valid-id", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for valid requests, got %d", w.Code)
		}
	}
	if len(limiter.records) != 0 {
		t.Fatalf("successful requests must not create limiter records, got %d", len(limiter.records))
	}
}

func TestParseTrustedProxyList(t *testing.T) {
	got := parseTrustedProxyList("127.0.0.1/32, ::1/128 ,10.0.0.0/8,,")
	want := []string{"127.0.0.1/32", "::1/128", "10.0.0.0/8"}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d: expected %s, got %s", i, want[i], got[i])
		}
	}

	if got := parseTrustedProxyList(""); len(got) != 0 {
		t.Fatalf("expected empty list for empty setting, got %v", got)
	}
}
