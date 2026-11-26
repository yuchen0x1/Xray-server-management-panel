package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRemoteIpIgnoresForwardedHeadersFromUntrustedRemote(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "203.0.113.10:12345"
	c.Request.Header.Set("X-Real-IP", "198.51.100.9")
	c.Request.Header.Set("X-Forwarded-For", "198.51.100.8")

	if got := getRemoteIp(c); got != "203.0.113.10" {
		t.Fatalf("remote IP = %q, want request remote address", got)
	}
}

func TestGetRemoteIpHonorsForwardedHeadersFromTrustedLoopbackProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "127.0.0.1:12345"
	c.Request.Header.Set("X-Forwarded-For", "198.51.100.8, 127.0.0.1")

	if got := getRemoteIp(c); got != "198.51.100.8" {
		t.Fatalf("remote IP = %q, want forwarded client IP", got)
	}
}

func TestGetRemoteIpRejectsSpoofedForwardedPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		header string
		want   string
	}{
		{name: "forged prefix", header: "192.0.2.99, 198.51.100.8", want: "198.51.100.8"},
		{name: "multiple trusted proxies", header: "192.0.2.99, 198.51.100.8, 127.0.0.1, ::1", want: "198.51.100.8"},
		{name: "malformed nearest hop", header: "192.0.2.99, unknown", want: "127.0.0.1"},
		{name: "all hops trusted", header: "127.0.0.1, ::1", want: "127.0.0.1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.RemoteAddr = "127.0.0.1:12345"
			c.Request.Header.Set("X-Forwarded-For", tt.header)
			if got := getRemoteIp(c); got != tt.want {
				t.Fatalf("remote IP = %q, want %q", got, tt.want)
			}
		})
	}
}
