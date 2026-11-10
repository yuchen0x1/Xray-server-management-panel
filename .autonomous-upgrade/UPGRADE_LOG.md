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

---

## 轮次 3 — API 令牌作用域与过期（2026-11-10 14:00–22:20）

### 3.1 令牌等同管理员且永久有效（high）

`internal/database/model/model.go`、`internal/web/service/panel/api_token.go`、
`internal/web/controller/api.go`、`internal/web/controller/setting.go`

**问题**：`ApiToken` 只有 name/token/enabled/createdAt。任何令牌——给监控脚本的、
给某个节点用的——都拥有完整管理员权限，而且永不过期。一个被泄露的低价值令牌
等于交出整个面板。

**修复**：

- 模型新增 `Scope`（默认 `admin`）与 `ExpiresAt`（0 = 永不过期），`Token` 列加索引。
- `MatchToken` 按哈希走索引单条查询后校验 scope 与过期；未知 scope 直接判为不可用，
  不回退为管理员。`Match` 保留为布尔包装，调用方无需改变。
- 新增 `enforceTokenScope` 中间件，按 Gin 路由模式白名单放行：
  - `monitor`：`/server/status`、`cpuHistory`、`history`、xray 指标与 observatory、
    `getXrayVersion`、`getPanelUpdateInfo`、`nodes/history`；
  - `node-sync`：`runtime/remote.go` 实际使用的全部集群 RPC；
  - `node-admin`：node-sync + `/server/updatePanel`。
- 会话登录用户与 admin 令牌行为完全不变（不带 `api_token_scope` 即跳过校验）。
- `DeleteExpectedScope` / `SetEnabledExpectedScope` 提供可选的 TOCTOU 保护；缺省时
  回退到按 ID 操作，保证旧调用方与旧前端不被破坏。

**关键判断**：node-sync 白名单是"漏一项就静默 403 破坏集群同步"的高风险面，因此
不是照抄上游列表，而是从本仓库 `runtime/remote.go` 与 `node.go` 的真实调用点反推，
并用 `api_scope_test.go` 的 `nodeSyncRequiredRoutes` 表把两者钉死。

### 3.2 旧库升级后令牌全部失效的风险（high，迁移）

`internal/database/db.go`

**问题**：`Scope` 的 `default:admin` 只对 INSERT 生效。若某驱动的 `ALTER TABLE`
没把默认值回填到既有行，这些行 scope 为空，而 `MatchToken` 拒绝未知 scope——
升级即等于吊销所有已签发令牌。

**修复**：新增 `backfillApiTokenScope()`，在 `initModels` 中把 `scope` 为 NULL 或空
字符串的行显式更新为 `admin`，并打印回填数量。

### 3.3 `-getApiToken` 无限累积管理员令牌（medium）

`main.go`

**问题**：已有令牌时，CLI 每次调用都新建一个 `cli-fallback-<unix>`，名字唯一所以
永远删不掉、也列不全，实际是一堆无法审计的管理员凭据。

**修复**：新增 `ApiTokenService.RecreateByName`，按名字在事务内替换为唯一一条；
保留原令牌的过期时间（已过期则拒绝复活），保留原 scope（未知 scope 拒绝 rotate）。
CLI 改为固定名 `cli-fallback` 的轮换。

### 3.4 前端与文档

- `SecurityTab.tsx`：新增 scope 选择器、过期时间选择器、每行 scope 标签；
  删除/启停带上 `expectedScope`；保留创建时间显示。
- 13 个 locale 各补 9 个键（scope 名称、含义、过期时间、永不过期）。
- `endpoints.ts` 更新请求体与响应示例；`npm run gen` 重新生成
  `frontend/src/generated/*` 与 `frontend/public/openapi.json`。

### 3.5 验证

| 检查 | 结果 |
|---|---|
| `go test ./internal/web/service/panel/ ./internal/database/` | 通过 |
| `go test ./internal/web/controller/` | 通过 |
| `go test -shuffle=on -count=1 ./...` | 全部通过 |
| `golangci-lint run` | 0 issues |
| `npm run typecheck` / `npm run lint` | 通过 |
| `npm run test` | 84 文件 / 971 测试通过 |
| `npm run build` | 成功 |

**测试过程中发现并修正的两处自身错误**：
1. 种子辅助函数用 `Enabled: false` 直接 Create，命中 GORM `default:true` 省略零值的
   行为，行被写成启用态——改为插入后显式 Update，并在注释里记下这个陷阱。
2. 新增行渲染时挤掉了"创建时间"，被既有测试 `api-token-date.test.tsx` 抓住；
   改为创建时间与过期时间同时显示。

---

## 发布 — v3.12.0（2026-11-10）

- 版本文件 `internal/config/version` → `3.12.0`。
- 跳过 `v3.9.0`–`v3.11.0`：这三个号属于上游 3x-ui，本仓库标签命名空间与之重叠，
  复用会让 `git fetch --tags upstream` 因同名标签拒绝覆盖。
- 把 `TestGetBaseVersionIsUtf8` 从硬编码版本号改为校验"非空 + 仅数字与点"，
  保留"嵌入文件损坏能被发现"的原意，同时不再需要每次发版改测试。
- `main` fast-forward 合入，推 tag `v3.12.0`。
- 工作流默认 `prerelease: true`，而面板更新器读 `/releases/latest`（GitHub 会排除
  预发布），等于发不出去。已改为正式发布并设置为 latest。
- 产物：linux amd64/arm64/armv7/armv6/armv5/386/s390x + windows-amd64，共 8 个。
