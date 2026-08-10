package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDomainValidatorMiddlewareNormalizesHostComparison(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		host string
		want int
	}{
		{name: "case insensitive", host: "PANEL.EXAMPLE.COM", want: http.StatusOK},
		{name: "ipv4 port", host: "192.0.2.10:2053", want: http.StatusOK},
		{name: "ipv6 port", host: "[2001:db8::10]:2053", want: http.StatusOK},
		{name: "wrong host", host: "other.example.com", want: http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(DomainValidatorMiddleware(map[string]string{
				"case insensitive": "panel.example.com",
				"ipv4 port":        "192.0.2.10",
				"ipv6 port":        "2001:db8::10",
				"wrong host":       "panel.example.com",
			}[tc.name]))
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "http://"+tc.host+"/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}
