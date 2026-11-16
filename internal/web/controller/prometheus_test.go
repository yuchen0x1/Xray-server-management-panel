package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

func TestRenderPrometheusMetrics(t *testing.T) {
	status := &service.Status{
		Cpu:      82.5,
		Loads:    []float64{1.2, 0.8, 0.4},
		TcpCount: 7,
		UdpCount: 3,
	}
	status.Xray.State = service.Running
	status.Mem.Current = 900
	status.Mem.Total = 1000

	got := renderPrometheusMetrics(status, true, []service.ObsTagSnapshot{
		{Alive: true, Delay: 40},
		{Alive: false, Delay: 80},
	})

	for _, want := range []string{
		"xui_panel_up 1",
		"xui_xray_up 1",
		"xui_xray_metrics_up 1",
		"xui_system_cpu_percent 82.5",
		"xui_system_memory_used_bytes 900",
		"xui_system_memory_total_bytes 1000",
		"xui_xray_observatory_targets 2",
		"xui_xray_observatory_alive 1",
		"xui_xray_observatory_max_delay_ms 80",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("metrics missing %q in %s", want, got)
		}
	}
}

func TestPrometheusEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := &ServerController{}
	engine := gin.New()
	engine.GET("/metrics", controller.prometheusMetrics)
	engine.GET("/alertRules", controller.alertRules)

	metricsResponse := httptest.NewRecorder()
	engine.ServeHTTP(metricsResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metricsResponse.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", metricsResponse.Code)
	}
	if got := metricsResponse.Header().Get("Content-Type"); got != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("metrics content type = %q", got)
	}
	if !strings.Contains(metricsResponse.Body.String(), "xui_panel_up 1") {
		t.Fatal("metrics response did not include panel gauge")
	}

	rulesResponse := httptest.NewRecorder()
	engine.ServeHTTP(rulesResponse, httptest.NewRequest(http.MethodGet, "/alertRules", nil))
	if rulesResponse.Code != http.StatusOK {
		t.Fatalf("alert rules status = %d, want 200", rulesResponse.Code)
	}
	if got := rulesResponse.Header().Get("Content-Disposition"); got != "attachment; filename=xui-alert-rules.yml" {
		t.Fatalf("alert rules disposition = %q", got)
	}
	if !strings.Contains(rulesResponse.Body.String(), "XUIXrayDown") {
		t.Fatal("alert rules response did not include Xray rule")
	}

	rules := string(prometheusAlertRules)
	for _, metric := range []string{
		"xui_panel_up",
		"xui_xray_up",
		"xui_system_cpu_percent",
		"xui_system_memory_used_bytes",
		"xui_system_memory_total_bytes",
		"xui_xray_observatory_targets",
		"xui_xray_observatory_alive",
	} {
		if !strings.Contains(rules, metric) {
			t.Fatalf("alert rules missing emitted metric %q", metric)
		}
	}
}
