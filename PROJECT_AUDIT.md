# 项目现状审计

审计日期：2026-11-16（Asia/Shanghai）

审计基线：`main`，HEAD `517f8977`，版本文件为 `3.12.1`。

## 审计范围与证据

本次审计结合了仓库源代码、Git 历史、构建脚本、依赖清单、CI 工作流、架构文档、前端源码、数据库模型、运行脚本和现有测试。没有把 README 的宣传文字当作实现证据。

| 检查项 | 结果 | 证据 |
|---|---|---|
| 分支与工作树 | PASS | `main` 与 `origin/main` 对齐；审计开始时工作树干净 |
| 版本 | PASS | `internal/config/version` 为 `3.12.1`；最近正式标签为 `v3.12.1` |
| 前端类型检查 | PASS | `cd frontend && npm run typecheck` |
| 前端 lint | PASS | `cd frontend && npm run lint` |
| 前端测试 | PASS | 87 个测试文件、978 个测试通过 |
| Go 测试 | PASS | Go 1.27.4、CGO/SQLite、`go test -shuffle=on -count=1 ./...` 全部通过；依赖使用 `goproxy.cn` |
| Go race | PASS | `go test -race -shuffle=on -count=1 ./...` 全部通过 |
| Go lint/build | PASS | golangci-lint 2.15.0 为 0 issues；`go build ./...` 通过 |
| Go 架构 | PASS | Gin、GORM、SQLite/PostgreSQL、Xray gRPC、cron、WebSocket、事件总线、多节点 runtime |
| 前端架构 | PASS | React 19、Ant Design 6、Vite 8、TypeScript、TanStack Query、Zod、Storybook |
| 发布与更新 | PARTIAL | Linux 服务更新脚本具备校验、路径检查和回滚；当前文档明确记录没有二进制签名门禁 |

## 项目定位

本项目是一个自托管的 Xray 代理服务器管理面板。它把入站、客户端、订阅、节点、Xray 配置、流量、证书、通知和备份集中到一个 Web 控制台，并通过受管的 Xray-core 与 MTProto 子进程把数据库状态应用到实际代理服务。

主要用户是需要管理单台或多台 Xray 节点的个人运营者、小型服务商、实验环境维护者和需要订阅链接管理的网络运维人员。核心场景是：创建入站和客户端、生成分享链接、限制流量与有效期、观察运行状态、把配置同步到远程节点、备份数据并升级面板。

## 真实技术架构

```text
React SPA / login / subscription viewer
        │ Fetch + WebSocket
        ▼
Gin panel server
        │ controllers → services → GORM
        ├── SQLite or PostgreSQL
        ├── Local runtime → Xray gRPC / managed Xray process
        ├── Remote runtime → node HTTPS API with verify / skip / pin / mTLS
        ├── MTProto manager → one mtg-multi process per MTProto inbound
        ├── Subscription server → raw / JSON / Clash output
        └── cron jobs + event bus + traffic writer + tunnel monitor
```

前端构建产物写入 `internal/web/dist/`，再由 Go 的 `embed.FS` 嵌入二进制。后端是状态源，入站和客户端数据先进入数据库，再由 runtime 把变更应用到本机或远程节点。Xray 配置变更优先使用 hot diff，不能热更新时才重启进程。

## 功能模块清单

### 已实现能力

- 入站 CRUD、端口冲突检测、fallback、流量与有效期管理。
- VMess、VLESS、Trojan、Shadowsocks、Hysteria、TUIC、WireGuard、MTProto、Mixed、HTTP、Tunnel 等协议表单与链接生成。
- 客户端 CRUD、批量添加、批量关联、批量解绑、分组、IP 限制、流量重置、到期和用量状态。
- 订阅 raw、JSON、Clash/Mihomo 输出，Host 覆盖和外部订阅聚合。
- 本地 Xray 管理、gRPC API、hot diff、进程监督、崩溃事件和重启策略。
- 多节点 runtime，支持远程 HTTPS、证书校验模式、证书 pin 和 mTLS。
- MTProto 多 secret、广告标签、配额/过期控制和管理 API 热应用。
- 备份、恢复、SQLite/PostgreSQL 迁移、Telegram 备份、日志和证书管理。
- Telegram、SMTP、Webhook 事件通知，CPU、内存、节点和出站健康事件。
- LDAP 同步、WARP/Nord 出站、路由、DNS、balancer、observatory 和 Xray 原始配置编辑。
- 13 个后端 locale、React 前端 i18n、OpenAPI/Zod 生成和 Storybook 组件工作台。
- locale key 检查已以 `en-US.json` 为 source，覆盖 13 个语言包、缺失/额外 key 和占位符一致性，并接入 CI。
- 受认证保护的 Prometheus 指标、告警规则下载、订阅 ETag/304、可选备份保留和页面级错误边界已加入升级实现。
- CI 中的 Go 测试、race、fuzz smoke、golangci、前端测试、构建、Storybook 和 npm audit。

### 代码中确认的缺陷与已修复项

#### 入站复制分享链接动作未接通

`frontend/src/pages/inbounds/list/RowActions.tsx` 已把 `copyLinks` 放入菜单，但 `InboundsPage.tsx` 原先没有处理这个 action，点击后落到默认分支并显示未完成提示。该缺陷已修复：现在会先 hydrate 入站，按协议生成分享链接或 WireGuard 链接，并调用剪贴板 API；复制成功后使用现有 `copied` 提示。

同一功能使用了 `pages.inbounds.copyShareLinks`，但该 key 原先不在 13 个 locale 文件中。已为所有 locale 补齐翻译。

#### LDAP 同步配置读取可能导致 cron 任务 panic

`internal/web/job/ldap_sync_job.go` 原先通过 `mustGetString`、`mustGetInt` 和 `mustGetBool` 把设置读取错误转换为 panic。cron 的 Recover 链可以保持调度器存活，但本轮同步会直接中断，并且配置或数据库暂时不可用时日志不够明确。现在改为返回带字段上下文的错误并记录 warning；同时不再忽略读取单个入站客户端列表的错误，也修复了空 flag 字段时提前求值旧字段的风险。

## 当前优势

1. 后端业务逻辑集中在 service，controller 较薄，runtime 抽象覆盖本地与远程节点。
2. Xray 变更使用 hot diff，减少了不必要的连接中断。
3. 数据库迁移、PostgreSQL、WAL 快照、备份恢复和持久化流量写入已经达到较高成熟度。
4. 协议覆盖面、订阅输出、Host 覆盖、MTProto 多 secret 和多节点能力形成了明显产品差异。
5. 前端有 Zod source of truth、OpenAPI 生成、严格 TypeScript、黄金链接夹具和 Storybook。
6. CI 已包含 race、fuzz smoke、依赖审计、locale key 检查和 Storybook 编译检查，不是只有单元测试。
7. 会话、CSRF、请求体大小、CSP、登录限流、API token scope、SSRF 防护和 mTLS 已有实际实现。

## 当前缺陷、技术债务和风险

### P0/P1 风险

- 更新文档明确说明当前只使用 SHA-256，没有 Authenticode、Cosign、Sigstore 或其他签名门禁。SHA-256 可检测传输损坏，但不能证明发布者身份。
- Go 版本要求为 1.27.4；若开发机只安装 1.26.x，自动工具链下载失败时会阻断本地后端验证。当前审计环境已安装 Go 1.27.4 和 CGO 编译器，并完成后端测试。
- 远程节点、Xray、MTProto、订阅服务和 cron 任务的运行状态仍分散在多个页面和日志中；本轮已提供基础 Prometheus 指标，但统一健康摘要和任务级指标仍缺失。
- 备份恢复有安全提示和互斥，但还没有在 CI 中定期执行跨版本恢复演练。
- 更新回滚依赖服务启动判断，Linux 外的安装更新、签名验证和真实旧版本升级仍未完成端到端证明。

### P2/P3 技术债务

- 入站列表和客户端列表的批量运维能力仍有扩展空间，例如批量启停、批量重置选中项和可恢复任务进度。
- 多节点同步有 dirty 状态，但用户界面尚未形成完整的变更预览、冲突解释和重试历史。
- WebSocket 事件、cron 运行、Xray 重启和流量写入缺少统一的结构化指标维度。
- API 文档由手维护 endpoint registry 拼装，新增路由容易漏登记。
- API 文档由手维护 endpoint registry 拼装，新增路由容易漏登记；locale key 已有 CI 检查，但变量语义仍需人工校对。
- 本轮新增观测路由已纳入现有真实路由契约测试，并重新生成 OpenAPI；后续仍应保持 registry 与 Gin 路由同步。
- 前端部分页面仍然较大，移动端列表和大量客户端数据的渲染可继续做虚拟化和渐进加载。

## 安全评估摘要

当前已观察到的防护包括：CSP nonce、HSTS 条件启用、X-Frame-Options、Referrer-Policy、CSRF、登录失败限流、请求体限制、API token hash、远程节点 TLS 模式、mTLS、出站 URL SSRF 检查、归档路径检查和敏感设置脱敏。

本次没有发现足以证明存在立即远程利用的漏洞。仍需优先补强更新签名、备份恢复演练、依赖供应链证明、结构化审计日志和管理员操作可追踪性。

## 运行与构建基线

- 前端类型检查、lint 和测试已通过。
- 前端测试输出包含 jsdom 对 `getComputedStyle` pseudo-elements 的提示，但没有失败；应在后续升级 Vitest/Storybook 配置时单独处理噪声。
- Go 全量测试、race、lint 和构建已在 Go 1.27.4、CGO/SQLite 环境通过；第一次执行曾因工具链下载和 C 编译器缺失受阻，补齐环境后复验通过。
- 临时数据库运行 smoke 已确认 `/panel/` 和根路径返回 200，进程可停止；源码检出缺少 `bin` 运行时资产，Xray 子进程未能启动，因此这不是生产端到端证明。
- 本地尚未完成 Linux systemd 服务替换、签名验证、真实旧版本升级、跨节点故障恢复和物理客户端连接验收。

## 成熟度判断

项目属于功能覆盖较广、工程化基础较强的自托管管理面板，当前可称为“具备生产使用基础但仍有发布供应链和运维可观测性缺口”的成熟阶段。不能仅凭现有本地前端测试或已发布版本就宣称所有平台、更新、签名和跨节点流程都已完成生产级端到端验证。

## 后续升级方向

优先顺序是：

1. 发布供应链：签名 manifest、SBOM、provenance、健康门禁和回滚证据。
2. 稳定性：任务错误分类、LDAP/节点/Xray 重试与状态、备份恢复演练。
3. 可观测性：扩展 Prometheus 指标、统一健康摘要、审计日志和事件历史。
4. 运维体验：入站/客户端批量操作、同步差异预览、导入 dry-run、移动端渐进加载。
5. 工程化：locale key CI、API registry 校验、依赖自动升级和跨平台 smoke 测试。
