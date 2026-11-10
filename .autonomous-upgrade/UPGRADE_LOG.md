# 升级日志

## 轮次 1 — 安全与数据完整性（2026-11-10 12:25–13:35）

### 1.1 订阅限流器失败放行（high）

`internal/sub/sub_limiter.go`

**问题**：记录表达到 `subLimitMaxRecords`（10000）且无可回收记录时，`registerMiss` 直接 `return`，新来源的 404 完全不被计数。攻击者可在高流量时段把限流器推入失败放行状态，此后从任意新地址无限枚举 subId。IPv6 下更易触发：`c.ClientIP()` 原样作为 key，单个 /64 可提供海量地址。

**修复**：
- 新增 `shedOldestLocked()`，表满时淘汰 `windowStart` 最旧的记录而不是丢弃新来源。
- 新增 `subLimiterKey()`，IPv6 按 /64 归一化、IPv4 按 /32、映射地址先 Unmap。
- `rateLimitMiddleware` 改用归一化 key。

**测试**：新增 4 个用例（表满仍追踪新来源、表满仍能封禁、/64 分组、中间件对同 /64 多地址统一拦截）；原 `TestSubLimiterCapsRecords` 补充"新来源必须被追踪"断言。

### 1.2 Webhook 密钥随跨域重定向外泄（medium）

`internal/web/service/webhook/webhook.go`

**问题**：`http.Client` 未设 `CheckRedirect`，Go 默认跟随最多 10 次重定向。Go 只剥离固定敏感头（Authorization/Cookie 等），自定义 `X-Webhook-Secret` 不在其中，因此 307/308 跳到任意主机时密钥与原请求体（主机名、登录用户名/IP）会被重放。

**修复**：`CheckRedirect` 拒绝跨主机跳转，同主机最多 3 次。

**测试**：`TestDispatchRefusesCrossHostRedirect` 用两个 httptest 服务器验证密钥不会跟随跳转。

### 1.3 Webhook 系统指标告警无阈值（medium）

`internal/web/service/webhook/webhook.go`、`internal/web/web.go`

**问题**：`CheckCpuJob`/`CheckMemJob` 每分钟无条件发布 `cpu.high`/`memory.high`，email 与 tgbot 各自按自己的配置阈值门控，webhook 订阅者没有门控，导致空闲主机每分钟收到一条"CPU 5.2% (threshold: 0%)"。

**修复**：
- `Subscriber` 新增 `metricThresholds`，由 `XUI_WEBHOOK_CPU_THRESHOLD` / `XUI_WEBHOOK_MEMORY_THRESHOLD` 配置（默认 90，越界回退）。
- `wants()` 在投递前按本订阅者阈值门控。
- 摘要改用订阅者实际阈值，不再显示发布者从未填写的 `Threshold`（恒为 0）。

### 1.4 纯 webhook 部署下 CPU/内存告警永不触发（medium）

`internal/web/web.go`

**问题**：`cpuAlarmWanted()`/`memoryAlarmWanted()` 只检查 Telegram 与 SMTP 配置。只配置 `XUI_WEBHOOK_URL` 时采样任务根本不注册，webhook 的 CPU/内存分支是死代码，运维无任何报错。

**修复**：两个方法先检查 `webhookConfigured()`（env-only），并把重复的双通知者判断抽成 `anyNotifierWants()`。

### 1.5 Webhook 无并发上限（high）

**问题**：`HandleEvent` 对每个事件 `go s.dispatch(...)`，无上限。登录失败事件被排除在限流之外且无需认证即可触发，可造成 goroutine 与 socket 堆积。

**修复**：引入 `dispatchSlots` 带缓冲通道（上限 16），满时丢弃并记 debug 日志。

### 轮次 1 验证

| 检查 | 结果 |
|---|---|
| `go build ./...` | 通过 |
| `go test ./internal/sub/`（限流相关 11 用例） | 全部通过 |
| `go test ./internal/web/service/webhook/`（8 用例） | 全部通过 |
| `go test ./internal/web/` | 通过 |

## 轮次 2 — TUIC UDP 中继预算验证（2026-11-10 13:35–14:05）

### 2.1 补 UDP 中继上限回归测试

新增 `internal/tuic/udp_relay_budget_test.go`，3 个用例：

- `TestEnsureRelayRejectsClientPastRelayBudget`：计数器已达上限时必须在开 SOCKS 隧道前拒绝，计数器不变。
- `TestEnsureRelayReleasesBudgetWhenDialFails`：拨号失败必须归还预算，否则重复失败会耗尽配额。
- `TestRelayBudgetReturnsOnCloseAndIsIdempotent`：`onClose` 只执行一次，重复 Close 不重复扣减。

### 2.2 修复 `SocksUDPSession.Close()` 的 nil panic（medium）

`internal/tuic/socks_bridge.go:277`

**发现方式**：新测试直接触发 `panic: runtime error: invalid memory address or nil pointer dereference`。

**问题**：`Close()` 无条件调用 `s.udpConn.Close()` 与 `s.ctrl.Close()`。生产路径目前保证二者非 nil，但作为资源释放方法，零值实例或部分构造实例调用会崩溃。

**修复**：两处加 nil 判断，`ctrl` 为 nil 时返回 nil。

### 2.3 验证

| 检查 | 结果 |
|---|---|
| `go test -run 'TestEnsureRelay\|TestRelayBudget' ./internal/tuic/` | 3/3 通过 |
| `go test -shuffle=on -count=1 ./...` | 全部通过 |
| `golangci-lint run` | 0 issues |
