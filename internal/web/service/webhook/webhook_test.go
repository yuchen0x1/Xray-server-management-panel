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
	var lastEvent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		var p Payload
		if err := json.NewDecoder(r.Body).Decode(&p); err == nil {
			lastEvent = p.Event
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

	time.Sleep(100 * time.Millisecond)

	if received.Load() == 0 {
		t.Errorf("expected webhook call to server, but got none")
	}
	if lastEvent != string(eventbus.EventOutboundDown) {
		t.Errorf("expected event %s, got %s", eventbus.EventOutboundDown, lastEvent)
	}
}

func TestFormatSummary(t *testing.T) {
	evCrash := eventbus.Event{
		Type: eventbus.EventXrayCrash,
		Data: "exit 1",
	}
	summary := formatSummary(evCrash)
	if summary != "Xray core crashed: exit 1" {
		t.Fatalf("unexpected summary: %s", summary)
	}
}
