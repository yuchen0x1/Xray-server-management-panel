# Prometheus monitoring

The panel exposes low-cardinality Prometheus metrics at
`/panel/api/server/metrics`. The endpoint uses the panel's normal session or
Bearer token authentication and returns the current panel, Xray, host resource,
and observatory gauges.

Download the matching alert rules from `/panel/api/server/alertRules`:

```sh
curl --fail --header "Authorization: Bearer $XUI_TOKEN" \
  -o xui-alert-rules.yml \
  https://panel.example.com/panel/api/server/alertRules
```

The rules use only metrics emitted by the endpoint. Tune the CPU and memory
thresholds for the host before loading them into Prometheus.

```yaml
scrape_configs:
  - job_name: xui
    metrics_path: /panel/api/server/metrics
    authorization:
      credentials_file: /etc/prometheus/xui-token
    static_configs:
      - targets: [panel.example.com]
```

Keep the endpoint behind the panel authentication boundary or a trusted reverse
proxy. The metrics intentionally avoid client emails, inbound tags, API tokens,
and other high-cardinality or secret values.
