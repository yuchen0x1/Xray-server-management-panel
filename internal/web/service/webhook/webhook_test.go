package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
)

func TestWebhookSubscriberDispatchesPayload(t *testing.T) {
	var received atomic.Int32
	var lastEvent atomic.Value

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		var p Payload
		if err := json.NewDecoder(r.Body).Decode(&p); err == nil {
			lastEvent.Store(p.Event)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sub := NewSubscriber(server.URL, "test-secret")
	sub.client.Transport = http.DefaultTransport

	sub.HandleEvent(eventbus.Event{
		Type:      eventbus.EventOutboundDown,
		Source:    "out-1",
		Timestamp: time.Now(),
	})

	deadline := time.Now().Add(2 * time.Second)
	for received.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if received.Load() == 0 {
		t.Fatal("expected webhook call to server, but got none")
	}
	if got, _ := lastEvent.Load().(string); got != string(eventbus.EventOutboundDown) {
		t.Errorf("expected event %s, got %s", eventbus.EventOutboundDown, got)
	}
}

func TestFormatSummary(t *testing.T) {
	thresholds := metricThresholds{cpuPercent: 90, memoryPercent: 85}
	evCrash := eventbus.Event{
		Type: eventbus.EventXrayCrash,
		Data: "exit 1",
	}
	summary := formatSummary(evCrash, thresholds)
	if summary != "Xray core crashed: exit 1" {
		t.Fatalf("unexpected summary: %s", summary)
	}

	evCPU := eventbus.Event{
		Type: eventbus.EventCPUHigh,
		Data: &eventbus.SystemMetricData{Percent: 95.5},
	}
	if got, want := formatSummary(evCPU, thresholds), "CPU usage high: 95.5% (threshold: 90%)"; got != want {
		t.Fatalf("cpu summary = %q, want %q", got, want)
	}

	evMem := eventbus.Event{
		Type: eventbus.EventMemoryHigh,
		Data: &eventbus.SystemMetricData{Percent: 88.2},
	}
	if got, want := formatSummary(evMem, thresholds), "Memory usage high: 88.2% (threshold: 85%)"; got != want {
		t.Fatalf("memory summary = %q, want %q", got, want)
	}
}

func TestWantsGatesMetricsOnOwnThreshold(t *testing.T) {
	sub := &Subscriber{webhookURL: "http://127.0.0.1:1", metrics: metricThresholds{cpuPercent: 80, memoryPercent: 75}}

	cases := []struct {
		name string
		ev   eventbus.Event
		want bool
	}{
		{name: "cpu below threshold", ev: eventbus.Event{Type: eventbus.EventCPUHigh, Data: &eventbus.SystemMetricData{Percent: 12}}, want: false},
		{name: "cpu at threshold", ev: eventbus.Event{Type: eventbus.EventCPUHigh, Data: &eventbus.SystemMetricData{Percent: 80}}, want: false},
		{name: "cpu above threshold", ev: eventbus.Event{Type: eventbus.EventCPUHigh, Data: &eventbus.SystemMetricData{Percent: 91}}, want: true},
		{name: "memory below threshold", ev: eventbus.Event{Type: eventbus.EventMemoryHigh, Data: &eventbus.SystemMetricData{Percent: 40}}, want: false},
		{name: "memory above threshold", ev: eventbus.Event{Type: eventbus.EventMemoryHigh, Data: &eventbus.SystemMetricData{Percent: 76}}, want: true},
		{name: "unrelated event always passes", ev: eventbus.Event{Type: eventbus.EventXrayCrash, Data: "exit 1"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sub.wants(tc.ev); got != tc.want {
				t.Fatalf("wants = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEnvThresholdFallsBackOnInvalidValues(t *testing.T) {
	cases := []struct {
		raw  string
		want float64
	}{
		{raw: "", want: 90},
		{raw: "not-a-number", want: 90},
		{raw: "0", want: 90},
		{raw: "-5", want: 90},
		{raw: "101", want: 90},
		{raw: "75.5", want: 75.5},
	}
	for _, tc := range cases {
		t.Setenv("XUI_TEST_THRESHOLD", tc.raw)
		if got := envThreshold("XUI_TEST_THRESHOLD", 90); got != tc.want {
			t.Fatalf("envThreshold(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

func TestDispatchRefusesCrossHostRedirect(t *testing.T) {
	var leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Webhook-Secret") != "" {
			leaked.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	sub := NewSubscriber(origin.URL, "super-secret")
	sub.client.Transport = http.DefaultTransport
	sub.dispatch(Payload{Event: "test", Summary: "test"})

	if leaked.Load() != 0 {
		t.Fatal("webhook secret must not follow a redirect to another host")
	}
}
