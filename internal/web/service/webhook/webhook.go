package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/util/netsafe"
)

type Payload struct {
	Event     string    `json:"event"`
	Source    string    `json:"source,omitempty"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
	Summary   string    `json:"summary"`
	Data      any       `json:"data,omitempty"`
}

type Subscriber struct {
	webhookURL string
	secret     string
	client     *http.Client
	limiter    *eventbus.RateLimiter
}

func NewSubscriber(webhookURL, secret string) *Subscriber {
	transport := &http.Transport{
		DialContext:         netsafe.SSRFGuardedDialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &Subscriber{
		webhookURL: strings.TrimSpace(webhookURL),
		secret:     strings.TrimSpace(secret),
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
		limiter: eventbus.NewRateLimiter(1 * time.Minute),
	}
}

func (s *Subscriber) HandleEvent(e eventbus.Event) {
	if s.webhookURL == "" {
		return
	}

	if e.Type != eventbus.EventLoginAttempt {
		if !s.limiter.Allow(e.Type, e.Source) {
			return
		}
	}

	summary := formatSummary(e)
	host, _ := os.Hostname()

	payload := Payload{
		Event:     string(e.Type),
		Source:    e.Source,
		Host:      host,
		Timestamp: e.Timestamp,
		Summary:   summary,
		Data:      e.Data,
	}

	go s.dispatch(payload)
}

func formatSummary(e eventbus.Event) string {
	switch e.Type {
	case eventbus.EventOutboundDown:
		return fmt.Sprintf("Outbound [%s] is DOWN", e.Source)
	case eventbus.EventOutboundUp:
		return fmt.Sprintf("Outbound [%s] is UP", e.Source)
	case eventbus.EventXrayCrash:
		return fmt.Sprintf("Xray core crashed: %v", e.Data)
	case eventbus.EventNodeDown:
		return fmt.Sprintf("Node [%s] is DOWN", e.Source)
	case eventbus.EventNodeUp:
		return fmt.Sprintf("Node [%s] is UP", e.Source)
	case eventbus.EventCPUHigh:
		if data, ok := e.Data.(*eventbus.SystemMetricData); ok {
			return fmt.Sprintf("CPU usage high: %.1f%% (threshold: %d%%)", data.Percent, data.Threshold)
		}
		return "CPU usage high"
	case eventbus.EventMemoryHigh:
		if data, ok := e.Data.(*eventbus.SystemMetricData); ok {
			return fmt.Sprintf("Memory usage high: %.1f%% (threshold: %d%%)", data.Percent, data.Threshold)
		}
		return "Memory usage high"
	case eventbus.EventLoginAttempt:
		if data, ok := e.Data.(*eventbus.LoginEventData); ok {
			return fmt.Sprintf("Login %s for user %s from IP %s", data.Status, data.Username, data.IP)
		}
		return "Login attempt"
	default:
		return fmt.Sprintf("Event: %s (source: %s)", e.Type, e.Source)
	}
}

func (s *Subscriber) dispatch(payload Payload) {
	var bodyBytes []byte
	urlLower := strings.ToLower(s.webhookURL)

	if strings.Contains(urlLower, "dingtalk.com") {
		dingMsg := map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": fmt.Sprintf("[3x-ui] %s", payload.Summary),
				"text":  fmt.Sprintf("### 3x-ui 告警提醒\n- **主机**: %s\n- **事件**: %s\n- **摘要**: %s\n- **时间**: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(dingMsg)
	} else if strings.Contains(urlLower, "weixin.qq.com") {
		wechatMsg := map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": fmt.Sprintf("### 3x-ui 告警提醒\n> 主机: <font color=\"comment\">%s</font>\n> 事件: <font color=\"warning\">%s</font>\n> 摘要: %s\n> 时间: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(wechatMsg)
	} else if strings.Contains(urlLower, "feishu.cn") || strings.Contains(urlLower, "larksuite.com") {
		feishuMsg := map[string]any{
			"msg_type": "text",
			"content": map[string]string{
				"text": fmt.Sprintf("[3x-ui 告警]\n主机: %s\n事件: %s\n摘要: %s\n时间: %s", payload.Host, payload.Event, payload.Summary, payload.Timestamp.Format("2006-01-02 15:04:05")),
			},
		}
		bodyBytes, _ = json.Marshal(feishuMsg)
	} else if strings.Contains(urlLower, "discord.com") {
		discordMsg := map[string]any{
			"content": fmt.Sprintf("**[3x-ui Alert]** %s\nHost: `%s` | Event: `%s` | Time: `%s`", payload.Summary, payload.Host, payload.Event, payload.Timestamp.Format("2006-01-02 15:04:05")),
		}
		bodyBytes, _ = json.Marshal(discordMsg)
	} else {
		bodyBytes, _ = json.Marshal(payload)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		logger.Warning("webhook: create request failed:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "3x-ui-Webhook/1.0")
	if s.secret != "" {
		req.Header.Set("X-Webhook-Secret", s.secret)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		logger.Warning("webhook: dispatch failed:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warning(fmt.Sprintf("webhook: dispatch returned non-2xx status: %d", resp.StatusCode))
	}
}
