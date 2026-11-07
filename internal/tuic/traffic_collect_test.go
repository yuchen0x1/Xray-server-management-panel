package tuic

import (
	"testing"
	"time"
)

func waitForClientTraffic(t *testing.T, server *Server, email string, want int64) {
	t.Helper()
	var up, down int64
	deadline := time.Now().Add(2 * time.Second)
	for {
		for _, delta := range server.CollectClientTraffic() {
			if delta.Email == email {
				up += delta.Up
				down += delta.Down
			}
		}
		if up >= want && down >= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("traffic for %s up=%d down=%d, want at least %d each", email, up, down, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
