// Package middleware provides HTTP middleware functions for the 3x-ui web panel,
// including domain validation utilities.
package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DomainValidatorMiddleware returns a Gin middleware that validates the request domain.
// It extracts the host from the request, strips any port number, and compares it
// against the configured domain. Requests from unauthorized domains are rejected
// with HTTP 403 Forbidden status.
func DomainValidatorMiddleware(domain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := requestHost(c.Request.Host)

		if !strings.EqualFold(host, strings.TrimSpace(domain)) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func requestHost(hostport string) string {
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	return strings.Trim(hostport, "[]")
}
