package mtproto

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestCollectTrafficCountsFirstSampleAndCounterReset(t *testing.T) {
	var up, down atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"users":{"alice":{"connections":1,"bytes_in":%d,"bytes_out":%d}}}`, up.Load(), down.Load())
	}))
	defer srv.Close()
	mgr := &Manager{procs: map[int]*managed{1: {
		proc:    &Process{cmd: &exec.Cmd{Process: &os.Process{}}, done: make(chan struct{})},
		tag:     "inbound-1",
		apiPort: serverPort(t, srv),
		last:    map[string]clientCounters{},
	}}}
	for _, tc := range []struct {
		name     string
		up, down int64
		want     []Traffic
	}{
		{name: "first sample", up: 100, down: 200, want: []Traffic{{Tag: "inbound-1", Email: "alice", Up: 100, Down: 200}}},
		{name: "next delta", up: 130, down: 245, want: []Traffic{{Tag: "inbound-1", Email: "alice", Up: 30, Down: 45}}},
		{name: "no change", up: 130, down: 245},
		{name: "counter reset", up: 12, down: 20, want: []Traffic{{Tag: "inbound-1", Email: "alice", Up: 12, Down: 20}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			up.Store(tc.up)
			down.Store(tc.down)
			got, online := mgr.CollectTraffic()
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("traffic = %+v, want %+v", got, tc.want)
			}
			if !reflect.DeepEqual(online, []string{"alice"}) {
				t.Fatalf("online = %v, want [alice]", online)
			}
		})
	}
}

func TestCollectTrafficDiscardsReplacedProcess(t *testing.T) {
	mgr := &Manager{procs: map[int]*managed{}}
	replacement := &managed{last: map[string]clientCounters{"alice": {up: 10, down: 20}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mgr.mu.Lock()
		mgr.procs[1] = replacement
		mgr.mu.Unlock()
		_, _ = fmt.Fprint(w, `{"users":{"alice":{"connections":1,"bytes_in":100,"bytes_out":200}}}`)
	}))
	defer srv.Close()
	mgr.procs[1] = &managed{
		proc:    &Process{cmd: &exec.Cmd{Process: &os.Process{}}, done: make(chan struct{})},
		tag:     "old-inbound",
		apiPort: serverPort(t, srv),
		last:    map[string]clientCounters{"alice": {up: 30, down: 40}},
	}
	got, online := mgr.CollectTraffic()
	if len(got) != 0 || len(online) != 0 {
		t.Fatalf("stale process returned traffic=%+v online=%v", got, online)
	}
	if got := replacement.last["alice"]; got != (clientCounters{up: 10, down: 20}) {
		t.Fatalf("replacement baseline = %+v, want {up:10 down:20}", got)
	}
}

func TestScrapeStatsRejectsFailedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, `{"users":{"alice":{"bytes_in":100,"bytes_out":200}}}`)
	}))
	defer srv.Close()
	if users, ok := scrapeStats(serverPort(t, srv), ""); ok || users != nil {
		t.Fatalf("scrapeStats = (%v, %v), want (nil, false)", users, ok)
	}
}
