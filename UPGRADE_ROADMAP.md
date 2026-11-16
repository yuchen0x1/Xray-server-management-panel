# 项目升级路线图

编制日期：2026-11-16（Asia/Shanghai）

优先级定义：P0 为安全或核心稳定性风险，P1 为重要功能与运维缺口，P2 为能力、性能和架构增强，P3 为工程化与体验完善。

## 核心功能升级

### EV-001 — 入站复制分享链接动作接通

- 所属模块：前端入站列表。
- 当前问题：菜单显示 `copyLinks`，页面分发未处理，用户点击后得到未完成提示。
- 依据：`frontend/src/pages/inbounds/list/RowActions.tsx` 注册 action，而 `InboundsPage.tsx` 原先没有对应 case。
- 对标参考：v2rayN、v2rayNG 的一键复制/分享流程。
- 参考实现：从当前协议配置生成链接，写入系统剪贴板并反馈结果。
- 改进方案：hydrate 入站后调用现有链接生成器；WireGuard 使用 link fan-out；复用 `copied` 文案。
- 预期效果：菜单动作真实可用，减少复制配置的步骤。
- 文件或模块：`frontend/src/pages/inbounds/InboundsPage.tsx`、13 个 locale。
- 实施难度：低。
- 优先级：P1。
- 验收标准：VMess/VLESS/Trojan/Shadowsocks/Hysteria/WireGuard 入站点击后写入预期文本；失败时不报告成功。
- 状态：已实施。

### EV-002 — LDAP 同步配置读取取消 panic

- 所属模块：LDAP cron job。
- 当前问题：设置读取错误会被 `mustGet*` 转成 panic，单轮同步中止。
- 依据：`internal/web/job/ldap_sync_job.go` 原先的 `mustGetString`、`mustGetInt`、`mustGetBool`。
- 对标参考：Cockpit、Netdata 的任务错误状态和可恢复运维流程。
- 参考实现：把配置错误作为结构化失败记录，保留调度器和其他任务运行。
- 改进方案：新增 `loadConfig`，逐项返回带字段上下文的 error；不忽略客户端读取错误。
- 预期效果：数据库或配置暂时异常时不再产生 panic，日志能指出具体字段。
- 文件或模块：`internal/web/job/ldap_sync_job.go`。
- 实施难度：中。
- 优先级：P0。
- 验收标准：任一 LDAP 设置读取失败时当前 Run 返回且无 panic；正常配置保持原同步行为。
- 状态：已实施；Go 1.27.4、CGO/SQLite 与 `internal/web/job` 测试已通过。

### EV-003 — 入站导入 dry-run

- 所属模块：入站导入 API/UI。
- 当前问题：导入会直接写库并触发运行时应用，错误发现得晚。
- 依据：`/panel/api/inbounds/import` 当前直接执行导入。
- 对标参考：Marzban 的批量用户 API 和 Nginx Proxy Manager 的向导校验。
- 参考实现：先解析、校验、列出冲突和即将变更的字段，再确认提交。
- 改进方案：增加只读 `validate` 模式和差异摘要，不写 DB、不调用 Xray。
- 预期效果：降低批量误导入和端口冲突风险。
- 文件或模块：`internal/web/controller/inbound.go`、`internal/web/service/inbound.go`、前端导入 modal、OpenAPI registry。
- 实施难度：中。
- 优先级：P1。
- 验收标准：dry-run 返回所有错误/冲突；确认后同一 payload 才能写入。

### EV-004 — 配置应用结果统一化

- 所属模块：XrayService/runtime。
- 当前问题：热应用、重启、远程节点等待和失败状态分散在不同响应格式。
- 依据：`service/xray.go`、`runtime/remote.go`、WebSocket 通知使用不同字段。
- 对标参考：Cockpit 的服务操作状态和 Uptime Kuma 的探针结果。
- 参考实现：稳定的 operation id、状态、开始/结束时间、错误和节点信息。
- 改进方案：定义 `ApplyResult` DTO，controller、runtime、WebSocket 共用。
- 预期效果：前端能准确显示“已应用、待同步、失败、重试中”。
- 文件或模块：`internal/web/entity`、`service/xray.go`、`runtime`、前端通知。
- 实施难度：高。
- 优先级：P1。
- 验收标准：本地热应用、强制重启、远程离线和重连四种路径都返回可追踪状态。

### EV-005 — 选中入站批量启停

- 所属模块：入站列表与 InboundService。
- 当前问题：目前可批量删除，启停仍需逐行操作。
- 依据：`InboundList` 已有选择状态和批量删除入口。
- 对标参考：Marzban、Hiddify Manager 的批量用户/节点操作。
- 参考实现：选择集合后统一提交，返回成功、跳过、失败明细。
- 改进方案：新增 bulk enable endpoint，逐项复用 `SetInboundEnable` 并聚合结果。
- 预期效果：降低维护多个节点和端口时的操作成本。
- 文件或模块：`controller/inbound.go`、`service/inbound.go`、`frontend/src/pages/inbounds`、OpenAPI。
- 实施难度：中。
- 优先级：P1。
- 验收标准：混合本地/远程、在线/离线入站时，结果逐项可解释且不产生部分静默失败。

### EV-006 — 选中入站批量重置流量

- 所属模块：入站流量服务。
- 当前问题：只有全部重置和逐入站重置，无法对筛选后的集合执行操作。
- 依据：现有 `resetAllTraffics` 和单入站 reset API。
- 对标参考：Netdata/Uptime Kuma 的批量维护动作反馈。
- 参考实现：批量动作有确认、进度和跳过原因。
- 改进方案：复用单项事务，聚合已重置与失败列表。
- 预期效果：减少误操作“全部重置”的风险。
- 文件或模块：`service/inbound_traffic.go`、`controller/inbound.go`、入站列表。
- 实施难度：中。
- 优先级：P1。
- 验收标准：只影响选中 ID，远程节点失败能单独显示。

### EV-007 — 客户端批量任务进度与取消

- 所属模块：客户端批量服务/UI。
- 当前问题：大量客户端操作在一个请求中完成，长请求和失败重试不透明。
- 依据：`client_bulk.go` 规模较大，前端已有多个 bulk modal。
- 对标参考：Watchtower 的分阶段更新和 Cockpit 的任务反馈。
- 参考实现：任务状态、阶段、计数、错误和取消标记。
- 改进方案：先做进程内 operation registry，再决定是否持久化。
- 预期效果：大批量动作更可控，避免重复提交。
- 文件或模块：`internal/web/service/client_bulk.go`、WebSocket、前端 bulk modal。
- 实施难度：高。
- 优先级：P1。
- 验收标准：1000 个客户端批量变更可显示进度，取消后不再启动新项且已完成项可审计。

### EV-008 — 多节点同步差异预览

- 所属模块：runtime/node sync。
- 当前问题：dirty 状态存在，但用户难以知道具体差异和下一次同步会做什么。
- 依据：`Node.ConfigDirty`、`inbound_node.go` 的 reconcile 路径。
- 对标参考：Hiddify 的多服务器入口与 Nginx Proxy Manager 的资源变更摘要。
- 参考实现：以资源、字段、目标节点和冲突原因呈现差异。
- 改进方案：增加只读 diff endpoint，先渲染 JSON path 级差异。
- 预期效果：降低远程节点误同步风险。
- 文件或模块：`service/inbound_node.go`、`controller/node.go`、节点页面。
- 实施难度：高。
- 优先级：P1。
- 验收标准：离线编辑、重连 reconcile、冲突和成功同步均有可读 diff。

## 新增实用功能

### EV-009 — Prometheus 兼容指标端点

- 所属模块：服务器观测。
- 当前问题：已有 Xray metrics 和事件，但没有标准 scrape 端点。
- 依据：`xray_metrics.go`、`metric_history.go`、事件总线已存在。
- 对标参考：Prometheus 的 exposition 与 Netdata 的健康指标。
- 参考实现：稳定的 `xui_*` 指标名、labels 和 HELP/TYPE。
- 改进方案：只读 `/panel/api/server/metrics`，可选 token scope，禁止高基数 email label。
- 预期效果：接入现有监控平台，不新增数据库。
- 文件或模块：`controller/server.go`、`service/xray_metrics.go`、endpoint registry。
- 实施难度：中。
- 优先级：P1。
- 验收标准：curl 可抓取，面板/Xray/主机/observatory 指标稳定输出且不含敏感字段；节点、订阅和 cron 任务维度另由 EV-010、EV-027 继续扩展。
- 状态：已实施；新增受认证保护的低基数 Prometheus 端点，覆盖面板、Xray、主机资源和 observatory 状态。

### EV-010 — 面板/Xray/订阅黑盒健康探针

- 所属模块：健康检查与 UI。
- 当前问题：节点 heartbeat 与 tunnel monitor 分开，缺少面向操作者的统一健康页。
- 依据：`internal/tunnelmonitor`、`node_heartbeat_job.go`、订阅服务器已有独立进程。
- 对标参考：Uptime Kuma 的 probe、Netdata 的 health 状态。
- 参考实现：探针结果有 latency、状态、最近错误、连续失败数和恢复时间。
- 改进方案：复用现有事件 bus，先实现本机面板、订阅、Xray API 和节点探针。
- 预期效果：定位“面板可达但 Xray 不可用”等半故障状态。
- 文件或模块：`internal/web/job`、`eventbus`、首页健康卡片。
- 实施难度：高。
- 优先级：P1。
- 验收标准：停止 Xray、断开订阅端口、远程节点离线时 UI 与事件状态一致。

### EV-011 — 管理员审计日志

- 所属模块：认证、controller、数据库。
- 当前问题：重要操作主要依赖普通日志，无法按用户、资源和结果检索。
- 依据：登录、备份、恢复、token 和节点操作散落在不同服务。
- 对标参考：Cockpit 的系统操作上下文、Nginx Proxy Manager 的管理资源记录。
- 参考实现：操作人、动作、资源、结果、来源 IP、request id 和脱敏摘要。
- 改进方案：新增 append-only `audit_events` 表和查询 API，敏感值只存类型摘要。
- 预期效果：故障排查和安全追责更可靠。
- 文件或模块：`database/model`、migration、middleware、核心 services。
- 实施难度：高。
- 优先级：P1。
- 验收标准：登录、恢复、token 创建、节点变更和 Xray 重启都有记录且 secret 不落盘。

### EV-012 — 备份保留策略

- 所属模块：备份与 Telegram/文件存储。
- 当前问题：有手动和 Telegram 备份，但缺少按数量/天数清理的策略。
- 依据：`BackupModal`、Telegram backup 文档和 server service。
- 对标参考：Marzban 的备份服务、Hiddify 的自动备份。
- 参考实现：保留数量、最大年龄、失败不删除旧备份。
- 改进方案：先支持本地 backup 目录和 Telegram 任务的保留元数据。
- 预期效果：减少磁盘无界增长，保留可恢复窗口。
- 文件或模块：`service/server.go`、`job`、设置页和文档。
- 实施难度：中。
- 优先级：P1。
- 验收标准：清理只删除超出策略且已成功写入的新备份，恢复中的文件不被删除。
- 状态：已实施；可选环境变量启用本地快照保留，默认关闭，使用 0600 临时文件、fsync、原子重命名和前缀安全清理。

### EV-013 — 恢复前 schema/版本预检

- 所属模块：数据库 restore。
- 当前问题：文档提醒同主版本恢复，但上传阶段缺少结构化兼容报告。
- 依据：`docs/content/docs/en/operations/backup-restore.mdx` 的跨版本限制说明。
- 对标参考：PostgreSQL migration 工具和 Cockpit 的操作前检查。
- 参考实现：读取备份版本、schema、数据库类型和需要的迁移列表。
- 改进方案：恢复前返回兼容级别、风险项和确认文本。
- 预期效果：降低把旧库直接覆盖新库的风险。
- 文件或模块：`database/dump_*`、`service/server.go`、BackupModal。
- 实施难度：高。
- 优先级：P1。
- 验收标准：同版本可直接恢复；跨 major 版本必须显示风险并阻止无确认覆盖。

### EV-014 — 告警规则导出

- 所属模块：指标与通知。
- 当前问题：已有 CPU、内存、节点和出站事件，但外部监控没有规则样例。
- 依据：`eventbus/events.go`、CPU/memory jobs、SMTP/Telegram subscribers。
- 对标参考：Prometheus/Grafana 的 rule examples。
- 参考实现：提供 Prometheus alert rules 和 webhook payload 示例。
- 改进方案：把阈值、事件名和建议动作写入 docs 与下载文件。
- 预期效果：用户能快速接入现有监控系统。
- 文件或模块：`docs/`、`internal/web/controller/server.go`。
- 实施难度：低。
- 优先级：P2。
- 验收标准：导出的规则引用实际指标名，不能出现未实现指标。
- 状态：已实施；新增可下载的内置 Prometheus 规则文件，并用测试校验规则引用已输出指标。

### EV-015 — 客户端二维码批量导出

- 所属模块：客户端/订阅。
- 当前问题：单客户端有二维码，批量运维要逐个打开详情。
- 依据：`ClientQrModal`、`ClientInfoModal` 已有单项生成能力。
- 对标参考：wg-easy 的 peer 配置与二维码下载。
- 参考实现：按客户端生成 PNG/ZIP 或打印页，并显示 remark/expiry。
- 改进方案：复用现有 link generator，后台限制单批数量并给出空结果。
- 预期效果：现场部署和移动端导入更快。
- 文件或模块：`frontend/src/pages/clients`、必要的 server zip endpoint。
- 实施难度：中。
- 优先级：P2。
- 验收标准：混合协议批量导出不泄露其他客户端 secret，文件名稳定且可追踪。

### EV-016 — 订阅 ETag/条件请求

- 所属模块：`internal/sub`。
- 当前问题：客户端重复拉取订阅会重复生成完整内容。
- 依据：`sub/controller.go`、`service.go` 和现有限流器。
- 对标参考：v2rayN/v2rayNG 的订阅更新、Prometheus 的缓存语义。
- 参考实现：按订阅内容和设置生成 ETag，支持 `If-None-Match` 返回 304。
- 改进方案：不缓存敏感正文，只缓存 hash 和生成时间；修改客户端/入站后失效。
- 预期效果：减少 CPU、带宽和移动端流量。
- 文件或模块：`internal/sub/controller.go`、`service.go`、测试。
- 实施难度：中。
- 优先级：P2。
- 验收标准：内容未变时 304；客户端或 Host 变更后 ETag 改变。
- 状态：已实施；raw、JSON、Clash 响应支持 SHA-256 ETag、弱匹配、多值匹配和 304，正文不做内存缓存。

## UI/UX 体验升级

### EV-017 — 首次使用 onboarding

- 所属模块：首页、登录后流程。
- 当前问题：新管理员需要自行理解端口、证书、入站、客户端和订阅之间的关系。
- 依据：当前首页功能丰富但缺少强引导流程。
- 对标参考：Hiddify Manager 的安装/首次配置路径、Cockpit 的模块入口。
- 参考实现：步骤卡、完成度、跳过和稍后继续。
- 改进方案：只基于真实状态判断，不创建虚假配置；提供证书、入站、客户端、订阅和备份五步。
- 预期效果：降低首次可用时间。
- 文件或模块：`frontend/src/pages/index`、设置页、i18n。
- 实施难度：中。
- 优先级：P1。
- 验收标准：空库可完成流程；已有配置不会被 onboarding 覆盖；可重新打开。

### EV-018 — 统一任务抽屉

- 所属模块：全局通知与 WebSocket。
- 当前问题：长操作结果分散在 toast、日志和页面局部状态。
- 依据：备份、更新、节点同步和批量操作各自维护状态。
- 对标参考：Watchtower/Cockpit 的任务状态、Uptime Kuma 的通知历史。
- 参考实现：全局抽屉按 pending/success/failed 展示最近任务。
- 改进方案：建立统一 operation envelope，先覆盖更新、恢复、同步和批量动作。
- 预期效果：操作结果可回看，不会因 toast 消失而丢信息。
- 文件或模块：`frontend/src/components/feedback`、WebSocket entity、services。
- 实施难度：高。
- 优先级：P1。
- 验收标准：刷新页面后已完成任务仍能显示摘要；敏感字段被脱敏。

### EV-019 — 页面级错误边界

- 所属模块：React routes/layout。
- 当前问题：单个复杂表单或第三方 Swagger/Storybook 组件异常可能影响整页。
- 依据：`routes.tsx`、`ApiDocsPage` 和大量 lazy route。
- 对标参考：Cockpit 的模块隔离和成熟 React 管理台。
- 参考实现：route error boundary 提供重试、返回和错误 id。
- 改进方案：在 PanelLayout 与高风险页面边界加可翻译 fallback。
- 预期效果：局部故障不再白屏整个控制台。
- 文件或模块：`frontend/src/routes.tsx`、`PanelLayout.tsx`、i18n。
- 实施难度：中。
- 优先级：P1。
- 验收标准：模拟渲染异常时保留侧栏和重新加载入口，控制台错误不含 secret。
- 状态：已实施；lazy 页面增加页面级 ErrorBoundary，保留侧栏并提供重试和返回首页动作。

### EV-020 — 筛选条件持久化

- 所属模块：入站/客户端/节点列表。
- 当前问题：刷新或返回页面后搜索、节点和状态筛选丢失。
- 依据：`InboundList` 的 session-only filter 和 `ClientsPage` filter drawer。
- 对标参考：v2rayN 的订阅/节点筛选体验、Grafana 的 dashboard filter。
- 参考实现：按页面和用户浏览器保存安全的 UI 偏好，不保存 secret。
- 改进方案：URL query 优先，localStorage 作为回退；无效 ID 自动清理。
- 预期效果：重复运维操作更快，可分享筛选 URL。
- 文件或模块：各列表页、`utils/base-path.ts`。
- 实施难度：低。
- 优先级：P2。
- 验收标准：返回页面恢复筛选；换 base path 或节点删除后不会产生死筛选。
- 状态：已实施；入站页现在保存节点筛选和搜索词，并在节点删除后自动回到全量筛选。

### EV-021 — 移动端渐进加载和虚拟化

- 所属模块：客户端/入站列表。
- 当前问题：大量客户端同时渲染会增加移动端内存和交互延迟。
- 依据：`ClientsPage` 和 `InboundList` 的大列表渲染。
- 对标参考：v2rayNG 移动端列表、Grafana 大数据表格。
- 参考实现：服务端分页、窗口化渲染、保留选择状态。
- 改进方案：先为客户端表启用已有分页/keepPreviousData，再评估虚拟列表依赖。
- 预期效果：大规模账户列表滚动更稳定。
- 文件或模块：`frontend/src/pages/clients`、`useClients.ts`。
- 实施难度：中。
- 优先级：P2。
- 验收标准：5000 客户端时首屏和滚动无明显卡顿，选中项跨页保留。

### EV-022 — 键盘与读屏回归矩阵

- 所属模块：前端可访问性。
- 当前问题：已有 a11y helper，但复杂 Dropdown、Table、Modal 组合仍缺流程级验证。
- 依据：`utils/a11y.ts`、已有 screen-reader 修复历史和 Storybook a11y addon。
- 对标参考：Cockpit 的键盘运维界面和 Storybook addon-a11y。
- 参考实现：以登录、创建入站、复制链接、恢复备份为关键流程。
- 改进方案：增加可访问名称、焦点回收和键盘测试，不引入视觉重写。
- 预期效果：减少键盘用户和读屏用户的阻塞。
- 文件或模块：关键页面、`frontend/src/test`、Storybook stories。
- 实施难度：中。
- 优先级：P2。
- 验收标准：关键流程可只用键盘完成，Modal 关闭后焦点回到触发控件。

## 技术架构与发布工程

### EV-023 — 签名更新 manifest

- 所属模块：release/update.sh。
- 当前问题：当前只有 SHA-256，不能验证发布者身份。
- 依据：`docs/AUTO_UPDATE.md` 明确记录签名验证缺失。
- 对标参考：Netdata signed Manifest、Prometheus/Grafana signed release、GoReleaser provenance。
- 参考实现：manifest 包含版本、提交、平台、归档 hash、签名和公钥指纹。
- 改进方案：使用 Sigstore/Cosign 或项目维护的离线签名密钥；保留镜像回退但签名内容一致。
- 预期效果：防止镜像或传输层替换合法 hash 文件。
- 文件或模块：`.github/workflows/release.yml`、`update.sh`、`docs/AUTO_UPDATE.md`。
- 实施难度：高。
- 优先级：P0。
- 验收标准：签名错误、版本不匹配、平台不匹配都在替换前失败；旧版本保持原位。

### EV-024 — SBOM 与 provenance

- 所属模块：release workflow。
- 当前问题：发布归档缺少软件成分和构建来源证明。
- 依据：当前 release 工作流主要生成归档和 SHA256SUMS。
- 对标参考：Grafana SBOM、GoReleaser SBOM/provenance。
- 参考实现：每个归档附 SPDX/CycloneDX SBOM 和构建提交/工具链信息。
- 改进方案：在 GitHub Actions 生成并上传签名材料，正文链接到资产。
- 预期效果：漏洞响应和供应链审计更快。
- 文件或模块：`.github/workflows/release.yml`、release docs。
- 实施难度：中。
- 优先级：P1。
- 验收标准：Linux/Windows 产物各有 SBOM；内容中的版本、提交和产物 hash 一致。

### EV-025 — 更新前健康门禁

- 所属模块：Linux updater。
- 当前问题：已有启动失败回滚，但健康判断还可扩展到 HTTP、数据库和 Xray 子进程。
- 依据：`update.sh` 的替换/回滚流程和 `AUTO_UPDATE.md`。
- 对标参考：Watchtower health check、Uptime Kuma probes。
- 参考实现：替换后等待服务、健康 API、数据库打开和 Xray 状态均通过。
- 改进方案：将检查分级，超时则恢复并保留失败诊断。
- 预期效果：减少“进程启动但功能不可用”的假成功。
- 文件或模块：`update.sh`、`internal/web/controller/server.go`、systemd service。
- 实施难度：高。
- 优先级：P0。
- 验收标准：模拟 HTTP 200 但 Xray 未启动时触发回滚；健康成功才写 success 状态。

### EV-026 — 更新状态 API schema

- 所属模块：PanelUpdateModal/backend update status。
- 当前问题：状态文件和前端展示字段由约定驱动，扩展错误分类不够稳定。
- 依据：`update-status.json` 文档和 `PanelUpdateModal.tsx`。
- 对标参考：Watchtower JSON logs、统一 operation envelope。
- 参考实现：schema 版本、phase、code、message、startedAt、finishedAt、rollback。
- 改进方案：后端解析时校验未知字段并保持向后兼容。
- 预期效果：更新失败更容易诊断，前端不依赖字符串猜测。
- 文件或模块：`service/server.go`、`entity`、前端版本 modal、Zod schema。
- 实施难度：中。
- 优先级：P1。
- 验收标准：下载、校验、安装、健康、回滚各阶段都能显示明确 code。

### EV-027 — cron 任务运行状态模型

- 所属模块：jobs/observability。
- 当前问题：cron Recover 能保住调度器，但失败次数和最近运行结果不集中。
- 依据：`web.go` 的 cron chain 与多个 job 文件。
- 对标参考：Netdata health state、Uptime Kuma monitor history。
- 参考实现：任务名、上次开始/结束、耗时、连续失败和最近错误。
- 改进方案：统一 job wrapper，保留 SkipIfStillRunning 行为。
- 预期效果：能快速识别交通、节点、LDAP、备份等任务是否失活。
- 文件或模块：`internal/web/job`、`web.go`、metrics endpoint。
- 实施难度：高。
- 优先级：P1。
- 验收标准：任务 panic、跳过、成功和失败都有计数，不能暴露密码。

### EV-028 — 事件 schema 版本化

- 所属模块：eventbus/WebSocket/notifiers。
- 当前问题：事件名有定义，但 payload 演进缺少显式版本。
- 依据：`internal/eventbus/events.go`、WebSocket notifier 和 SMTP/Telegram consumers。
- 对标参考：Prometheus labels 与成熟 webhook payload 版本。
- 参考实现：`eventVersion`、稳定字段、扩展字段和兼容策略。
- 改进方案：先给现有事件 envelope 加版本，不改事件语义。
- 预期效果：通知和第三方集成升级更安全。
- 文件或模块：`internal/eventbus`、`websocket`、通知服务。
- 实施难度：中。
- 优先级：P2。
- 验收标准：旧消费者忽略新字段，新消费者能按版本处理 payload。

### EV-029 — API route registry 自动校验

- 所属模块：controller/OpenAPI。
- 当前问题：新增 Gin route 若忘记 endpoint registry 不会报错。
- 依据：AGENTS.md 明确说明 registry 是手维护且遗漏会静默消失。
- 对标参考：Go generate/OpenAPI generator 和 Cockpit 的测试门禁。
- 参考实现：测试扫描 controller route 与 endpoints.ts 的路径/方法集合。
- 改进方案：先实现只读差异测试，处理动态路径和内部非文档路由白名单。
- 预期效果：避免 API 已上线但文档缺失。
- 文件或模块：`internal/web/routes_contract_test.go`、`frontend/src/pages/api-docs/endpoints.ts`。
- 实施难度：中。
- 优先级：P1。
- 验收标准：新增公开 GET/POST/PUT/DELETE 路由时测试必须明确登记或列入白名单。
- 状态：已具备并复核；现有 `routes_contract_test.go` 已构造真实路由，检查 endpoints registry 的缺失和多余登记；本轮新增 metrics/alertRules 已登记并通过测试。

### EV-030 — locale key 完整性 CI

- 所属模块：后端/前端 i18n。
- 当前问题：缺 key 会 fallback 或直接显示原始 key，构建不失败。
- 依据：AGENTS.md 的 i18n 硬规则和本轮 `copyShareLinks` 缺失事实。
- 对标参考：大型管理台的 locale schema 校验。
- 参考实现：以 en-US 为 schema，检查其他 12 个 JSON 的 key 集合和变量占位符。
- 改进方案：新增 Go 或 Node 检查，允许明确的 legacy fallback 白名单。
- 预期效果：新功能不会只在英文或部分语言可见。
- 文件或模块：`tools/`、CI、translation JSON。
- 实施难度：低。
- 优先级：P1。
- 验收标准：缺 key、额外错误嵌套和 `{variable}` 不匹配时 CI 失败。
- 状态：已实施，`frontend/scripts/check-translations.mjs` 已接入 npm script 与 CI。

## 性能与稳定性

### EV-031 — 客户端与流量查询索引审计

- 所属模块：数据库模型/迁移。
- 当前问题：多节点、客户端 IP、流量和分页查询字段较多，索引需求随规模增长。
- 依据：`model` 中的 client/node traffic 表和 `client_paging.go`。
- 对标参考：Prometheus label cardinality 和成熟面板的分页索引策略。
- 参考实现：按真实 WHERE/ORDER 组合建立最小复合索引。
- 改进方案：用 SQLite EXPLAIN 与 PostgreSQL explain fixture 验证后再迁移。
- 预期效果：大规模客户和节点列表延迟更稳定。
- 文件或模块：`internal/database/db.go`、model、scale tests。
- 实施难度：中。
- 优先级：P2。
- 验收标准：常用列表查询计划使用目标索引，迁移可重复且不锁死旧库。

### EV-032 — Traffic writer 可观测性

- 所属模块：流量写入。
- 当前问题：批量持久化降低了 DB 压力，但积压、丢弃和 flush 延迟不明显。
- 依据：`service/traffic_writer.go` 和 Xray traffic job。
- 对标参考：Netdata agent queue/health 指标。
- 参考实现：队列深度、flush 次数、耗时、失败和重试计数。
- 改进方案：增加低基数 metrics 与日志采样。
- 预期效果：能判断流量延迟是 Xray、内存队列还是数据库导致。
- 文件或模块：`traffic_writer.go`、metrics endpoint、tests。
- 实施难度：中。
- 优先级：P2。
- 验收标准：压力测试下计数与实际 flush 一致，日志不含客户端 secret。

### EV-033 — 入站 hydrate 并发与去重

- 所属模块：前端入站列表。
- 当前问题：导出全部订阅和多个行操作可能重复 hydrate 同一入站。
- 依据：`InboundsPage.tsx` 的 `hydrateInbound` 调用路径。
- 对标参考：TanStack Query 的 request deduplication、v2rayN 的批量更新反馈。
- 参考实现：按 inbound id 缓存进行中的 Promise，失败可重试。
- 改进方案：在 `useInbounds` 内部提供 keyed request cache，限制并发数量。
- 预期效果：减少重复请求和大列表导出峰值。
- 文件或模块：`frontend/src/pages/inbounds/useInbounds.ts`、测试。
- 实施难度：中。
- 优先级：P2。
- 验收标准：同一 ID 并发调用只产生一次 HTTP 请求，失败不会永久缓存。

### EV-034 — 订阅输出缓存失效策略

- 所属模块：`internal/sub`。
- 当前问题：ETag 解决重复传输后，仍需避免每次请求重复解析大配置。
- 依据：`sub/service.go` 规模较大，配置变更路径分散。
- 对标参考：Netdata 缓存与 Prometheus scrape 低延迟设计。
- 参考实现：按 node/inbound/client/settings generation 缓存渲染结果。
- 改进方案：先做进程内 bounded cache，所有写路径递增 generation。
- 预期效果：大订阅低延迟，修改后不返回旧数据。
- 文件或模块：`internal/sub/service.go`、写路径 service、tests。
- 实施难度：高。
- 优先级：P2。
- 验收标准：相同 generation 命中缓存，任一相关资源变更立即失效。

### EV-035 — LDAP 配置与失败路径测试

- 所属模块：LDAP job tests。
- 当前问题：当前测试覆盖 client 创建，但未覆盖设置读取失败和 Run 不 panic。
- 依据：`ldap_sync_job_test.go` 现有测试范围。
- 对标参考：Cockpit 强制测试门禁、Go stdlib table-driven 约定。
- 参考实现：可注入设置读取器或抽出纯配置装配函数。
- 改进方案：测试空 flag fallback、读取错误、客户端列表错误和正常配置。
- 预期效果：防止 panic 回归。
- 文件或模块：`internal/web/job/ldap_sync_job_test.go`、`ldap_sync_job.go`。
- 实施难度：中。
- 优先级：P0。
- 验收标准：错误路径精确断言日志/错误，不触发 panic；Go 工具链恢复后纳入 CI。
- 状态：部分实施，已覆盖非法端口配置的错误上下文和 `Run` 无 panic；其余数据库/连接失败矩阵继续保留。

### EV-036 — 节点探测重试与退避

- 所属模块：node heartbeat/runtime。
- 当前问题：节点短暂抖动时状态变化和重试策略需要统一观察。
- 依据：`node_heartbeat_job.go`、`NodeService.Probe`、TLS client。
- 对标参考：Uptime Kuma 的连续失败/恢复和 Watchtower 的退避原则。
- 参考实现：区分连接、认证、证书和业务错误；指数退避且有上限。
- 改进方案：保留当前 cadence，增加单节点 next retry 和 reason。
- 预期效果：减少故障节点造成的请求风暴和误报。
- 文件或模块：`job/node_heartbeat_job.go`、`service/node.go`、runtime。
- 实施难度：中。
- 优先级：P1。
- 验收标准：证书错误不无限重试；网络恢复后按退避上限收敛并发布 node.up。

### EV-037 — Xray 重启预算与抑制

- 所属模块：Xray process/service。
- 当前问题：连续崩溃或健康监控触发重启时，需要防止重启循环。
- 依据：`xray/process.go`、`service/xray.go`、tunnel monitor。
- 对标参考：systemd restart limits、Netdata recovery policy。
- 参考实现：时间窗内计数、冷却期、最后错误和人工确认。
- 改进方案：不改变默认恢复能力，达到预算后进入 degraded 状态并告警。
- 预期效果：避免 CPU/日志风暴，保留现场信息。
- 文件或模块：`internal/xray/process.go`、`service/xray.go`、eventbus。
- 实施难度：高。
- 优先级：P1。
- 验收标准：模拟快速崩溃时不超过预算；人工重试可清除冷却期。

### EV-038 — MTProto orphan 诊断

- 所属模块：MTProto manager。
- 当前问题：已有 orphan cleanup，但管理员需要知道孤儿进程来源和清理结果。
- 依据：`internal/mtproto/orphans_*`、`manager.go`。
- 对标参考：Watchtower 容器生命周期日志、Cockpit 服务列表。
- 参考实现：记录 inbound id、PID、启动时间、清理原因和结果。
- 改进方案：将诊断接入 job metrics 和受限日志，不记录 secret。
- 预期效果：MTProto 子进程问题更容易排查。
- 文件或模块：`internal/mtproto`、`job/mtproto_job.go`、metrics。
- 实施难度：中。
- 优先级：P2。
- 验收标准：正常退出不误报；孤儿清理失败有可行动错误。

## 安全性升级

### EV-039 — secret/credential 日志扫描

- 所属模块：日志与 CI。
- 当前问题：已有多处脱敏，但随着新功能增加，难以保证所有错误路径不打印 token/password。
- 依据：controller/service/logger 分布广，节点 token、LDAP password、Telegram token 均敏感。
- 对标参考：Watchtower token-file 与安全日志修复、Grafana CVE release notes。
- 参考实现：敏感字段名规则扫描、运行时 redactor 和测试 fixture。
- 改进方案：CI 扫描新增日志文本，禁止直接格式化 secret 字段；对已知安全日志加测试。
- 预期效果：降低凭据进入日志和工单的概率。
- 文件或模块：`internal/logger`、CI、security tests。
- 实施难度：中。
- 优先级：P0。
- 验收标准：构造登录、节点、LDAP、更新失败日志，输出不含原始 secret。

### EV-040 — CSP 动态资源审计

- 所属模块：安全 headers/embedded frontend。
- 当前问题：CSP 已有 nonce 和 `unsafe-inline` style，但第三方 Swagger/Storybook/未来插件资源需持续审计。
- 依据：`middleware/security.go` 当前 CSP。
- 对标参考：Cockpit 的安全策略和成熟 Web 管理台。
- 参考实现：资源来源清单、nonce 传播和 CSP report-only 预览。
- 改进方案：先将允许来源写成测试契约，再逐步收紧 style/script。
- 预期效果：减少 XSS 影响面而不破坏现有 SPA。
- 文件或模块：`middleware/security.go`、dist tests、frontend API docs。
- 实施难度：中。
- 优先级：P1。
- 验收标准：主 SPA、登录、订阅、Swagger 页面在严格策略下正常加载。

### EV-041 — API token 轮换与撤销 UX

- 所属模块：SecurityTab/API token service。
- 当前问题：已有 scope/expiry/hash，但大量 token 管理仍依赖手动删除和重新创建。
- 依据：`service/panel/api_token.go`、`SecurityTab.tsx`。
- 对标参考：Marzban 管理员 API、Hiddify 角色化管理。
- 参考实现：轮换生成、旧 token grace window、最后使用时间和撤销确认。
- 改进方案：服务端只存 hash，前端一次性展示新 token；轮换不回显旧值。
- 预期效果：降低长期 token 暴露风险。
- 文件或模块：API token model/service/controller/frontend/OpenAPI。
- 实施难度：中。
- 优先级：P1。
- 验收标准：旧 token 在 grace 后失效，新 token 只展示一次，审计记录不含 token。

## 自动化、依赖和文档

### EV-042 — 依赖更新自动化分组

- 所属模块：GitHub Actions/dependency policy。
- 当前问题：依赖多且更新频繁，Go、npm、Xray、mtg 的更新风险不同。
- 依据：`go.mod`、`package.json`、release workflows。
- 对标参考：Renovate 的分组、迁移和自动测试。
- 参考实现：按安全补丁、前端、核心协议、构建工具分组，全部跑 verify。
- 改进方案：使用 Dependabot/Renovate 风格配置，遵守 main-only 分支政策。
- 预期效果：减少遗漏高危更新和一次性大 diff。
- 文件或模块：`.github/dependabot.yml` 或 Renovate 配置、CI。
- 实施难度：低。
- 优先级：P2。
- 验收标准：自动更新不创建不受控发布分支，合并前通过现有 gate。

### EV-043 — 发布 smoke matrix

- 所属模块：release/CI。
- 当前问题：CI 编译和单测不能证明 Linux 服务安装、Windows 启动、SQLite 初始化和 dist 嵌入均可用。
- 依据：`release.yml`、`smoke.yml` 和本地 Go 工具链限制。
- 对标参考：GoReleaser 跨平台产物矩阵、Cockpit merge tests。
- 参考实现：每个平台启动临时实例，调用 `/panel/api/server/status` 和登录/CSRF 流程。
- 改进方案：先 Linux amd64/arm64 与 Windows amd64，隔离真实凭据和生产资源。
- 预期效果：发布前捕获打包路径和启动回归。
- 文件或模块：`.github/workflows/smoke.yml`、scripts、docs。
- 实施难度：高。
- 优先级：P1。
- 验收标准：每个目标平台产物启动、数据库迁移、健康 API 和停止流程通过。

### EV-044 — 自动恢复演练

- 所属模块：CI/database/update。
- 当前问题：备份和回滚代码存在，但没有周期性演练证据。
- 依据：`docs/AUTO_UPDATE.md` 和 backup-restore 文档的限制说明。
- 对标参考：Netdata/Grafana 发行资产证明、成熟运维系统的 restore drill。
- 参考实现：临时 SQLite/PostgreSQL，导入 fixture，升级 schema，再恢复和比对关键记录。
- 改进方案：每周或手动 workflow，失败只阻止验证，不触碰生产。
- 预期效果：把“有备份代码”转成“恢复路径有证据”。
- 文件或模块：`.github/workflows`、database tests、fixtures。
- 实施难度：高。
- 优先级：P1。
- 验收标准：关键模型、客户端 secret hash、设置和订阅数据恢复后保持一致。

### EV-045 — 运维 runbook 与故障树

- 所属模块：docs。
- 当前问题：文档已有安装、备份、排障和多节点页面，但没有按症状统一索引所有新任务状态。
- 依据：`docs/architecture.md`、`docs/content/docs/en/help/troubleshooting.mdx`。
- 对标参考：Cockpit、Netdata 的故障上下文和 Uptime Kuma 的探针解释。
- 参考实现：症状 → 检查命令 → 日志字段 → 恢复动作 → 回滚点。
- 改进方案：补充 updater、LDAP、node sync、MTProto、订阅 304 和 metrics 章节。
- 预期效果：减少依赖开发者记忆的排障成本。
- 文件或模块：`docs/content/docs/{en,zh,ru,fa}`、`docs/architecture.md`。
- 实施难度：中。
- 优先级：P2。
- 验收标准：每个 P0/P1 路线项有真实命令、限制和“未验证”说明。

### EV-046 — 插件与扩展边界文档

- 所属模块：架构文档。
- 当前问题：协议、通知、runtime、订阅和 job 都有扩展点，但边界主要依赖代码阅读。
- 依据：`docs/architecture.md` 的 layering rules 和 service 目录结构。
- 对标参考：Cockpit 模块化、Prometheus exporter 边界、GoReleaser 声明式配置。
- 参考实现：明确 controller/service/runtime/util/eventbus 的依赖方向和可扩展接口。
- 改进方案：为新增协议、通知器、指标 exporter、job 写最小模板和测试要求。
- 预期效果：后续自主升级和社区贡献更少破坏边界。
- 文件或模块：`docs/architecture.md`、`CONTRIBUTING.md`、examples。
- 实施难度：低。
- 优先级：P3。
- 验收标准：新增模块能找到唯一归属、依赖方向和所需测试，不需猜目录。

## 剩余路线执行顺序

1. P0：EV-023、EV-025、EV-035、EV-039。
2. P1 稳定性和可观测性：EV-004、EV-005、EV-006、EV-008、EV-010、EV-011、EV-013、EV-018、EV-026、EV-027、EV-036、EV-037、EV-041、EV-043、EV-044。
3. P1 用户体验：EV-003、EV-017。
4. P2/P3：EV-007、EV-015、EV-021、EV-022、EV-024、EV-028、EV-031、EV-032、EV-033、EV-034、EV-038、EV-040、EV-042、EV-045、EV-046。

升级点总数：46。
