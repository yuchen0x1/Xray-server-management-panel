package tuic

import (
	"strings"
	"testing"
)

type sendStreamCloser interface {
	Close() error
}

func closeSendStream(t *testing.T, stream sendStreamCloser) {
	t.Helper()
	err := stream.Close()
	if err == nil || strings.Contains(err.Error(), "close called for canceled stream") {
		return
	}
	t.Fatal(err)
}
