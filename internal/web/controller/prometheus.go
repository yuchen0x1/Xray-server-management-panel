package controller

import (
	_ "embed"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

//go:embed prometheus-alert-rules.yml
var prometheusAlertRules []byte

func (a *ServerController) prometheusMetrics(c *gin.Context) {
	status := a.serverService.LastStatus()
	metricsState := a.xrayMetricsService.State()
	body := renderPrometheusMetrics(status, metricsState.Enabled, a.xrayMetricsService.ObservatorySnapshot())
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(body))
}

func (a *ServerController) alertRules(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", "attachment; filename=xui-alert-rules.yml")
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", prometheusAlertRules)
}

func renderPrometheusMetrics(status *service.Status, xrayMetricsEnabled bool, observatory []service.ObsTagSnapshot) string {
	var b strings.Builder
	writePrometheusGauge(&b, "xui_panel_up", "Whether the panel metrics endpoint is serving.", 1)
	xrayUp := 0.0
	if status != nil && status.Xray.State == service.Running {
		xrayUp = 1
	}
	writePrometheusGauge(&b, "xui_xray_up", "Whether the managed Xray process is running.", xrayUp)
	metricsUp := 0.0
	if xrayMetricsEnabled {
		metricsUp = 1
	}
	writePrometheusGauge(&b, "xui_xray_metrics_up", "Whether Xray runtime metrics are enabled and reachable.", metricsUp)

	if status == nil {
		writePrometheusGauge(&b, "xui_xray_observatory_targets", "Number of Xray observatory targets.", 0)
		writePrometheusGauge(&b, "xui_xray_observatory_alive", "Number of healthy Xray observatory targets.", 0)
		writePrometheusGauge(&b, "xui_xray_observatory_max_delay_ms", "Maximum Xray observatory delay in milliseconds.", 0)
		return b.String()
	}

	writePrometheusGauge(&b, "xui_system_cpu_percent", "Host CPU usage percentage.", status.Cpu)
	writePrometheusGauge(&b, "xui_system_memory_used_bytes", "Host memory used in bytes.", float64(status.Mem.Current))
	writePrometheusGauge(&b, "xui_system_memory_total_bytes", "Host memory total in bytes.", float64(status.Mem.Total))
	writePrometheusGauge(&b, "xui_system_swap_used_bytes", "Host swap used in bytes.", float64(status.Swap.Current))
	writePrometheusGauge(&b, "xui_system_swap_total_bytes", "Host swap total in bytes.", float64(status.Swap.Total))
	writePrometheusGauge(&b, "xui_system_disk_used_bytes", "Host disk used in bytes.", float64(status.Disk.Current))
	writePrometheusGauge(&b, "xui_system_disk_total_bytes", "Host disk total in bytes.", float64(status.Disk.Total))
	writePrometheusGauge(&b, "xui_system_tcp_connections", "Number of open TCP connections.", float64(status.TcpCount))
	writePrometheusGauge(&b, "xui_system_udp_connections", "Number of open UDP connections.", float64(status.UdpCount))
	loads := [3]float64{}
	copy(loads[:], status.Loads)
	writePrometheusGauge(&b, "xui_system_load1", "Host one minute load average.", loads[0])
	writePrometheusGauge(&b, "xui_system_load5", "Host five minute load average.", loads[1])
	writePrometheusGauge(&b, "xui_system_load15", "Host fifteen minute load average.", loads[2])

	alive := 0
	maxDelay := int64(0)
	for _, snapshot := range observatory {
		if snapshot.Alive {
			alive++
		}
		if snapshot.Delay > maxDelay {
			maxDelay = snapshot.Delay
		}
	}
	writePrometheusGauge(&b, "xui_xray_observatory_targets", "Number of Xray observatory targets.", float64(len(observatory)))
	writePrometheusGauge(&b, "xui_xray_observatory_alive", "Number of healthy Xray observatory targets.", float64(alive))
	writePrometheusGauge(&b, "xui_xray_observatory_max_delay_ms", "Maximum Xray observatory delay in milliseconds.", float64(maxDelay))
	return b.String()
}

func writePrometheusGauge(b *strings.Builder, name, help string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		value = 0
	}
	b.WriteString("# HELP ")
	b.WriteString(name)
	b.WriteByte(' ')
	b.WriteString(help)
	b.WriteByte('\n')
	b.WriteString("# TYPE ")
	b.WriteString(name)
	b.WriteString(" gauge\n")
	b.WriteString(name)
	b.WriteByte(' ')
	b.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	b.WriteByte('\n')
}
