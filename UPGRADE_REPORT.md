# 项目升级交付报告

执行日期：2026-11-16（Asia/Shanghai）
执行分支：`main`
基线提交：`517f8977`
当前版本：`3.13.0`

## 交付范围

本轮按“先审计、再对标、再实施、再验证”的顺序处理了仓库。完整审计见 [PROJECT_AUDIT.md](PROJECT_AUDIT.md)，17 个开源项目的研究记录见 [COMPETITOR_ANALYSIS.md](COMPETITOR_ANALYSIS.md)，全部 46 项候选升级见 [UPGRADE_ROADMAP.md](UPGRADE_ROADMAP.md)。

研究覆盖直接竞品、协议核心、客户端、网络管理面板、可观测性、更新器、依赖自动化和发布工程。重点参考了 Marzban、Hiddify Manager、Cockpit、Netdata、Prometheus、Grafana、Uptime Kuma、Watchtower、Renovate 和 GoReleaser 等项目的公开仓库、README、目录、许可证、发布记录、工作流和相关议题。没有复制第三方代码。

## 已实施改进

| 路线项 | 改动 | 验收结果 |
|---|---|---|
| EV-001 | 在 `frontend/src/pages/inbounds/InboundsPage.tsx` 接通 `copyLinks`，复用现有链接生成器、WireGuard 分支和剪贴板服务；13 个 locale 补齐 `pages.inbounds.copyShareLinks`。 | 菜单动作进入实际链接生成与复制流程；无链接时仍不展示动作。 |
| EV-002 | 在 `internal/web/job/ldap_sync_job.go` 增加带字段上下文的配置加载；移除 `mustGet*` panic 路径；客户端和标签读取错误会记录并结束当前轮次。 | LDAP 设置或数据库读取异常不会再直接 panic 调度器；正常配置路径保持原有同步流程。 |
| EV-030 | 新增 `frontend/scripts/check-translations.mjs`，以 `en-US.json` 为 key source 检查其他 12 个 locale 的缺失或空值，并接入 npm 与 CI。 | 本轮 13 个 locale 的 key 集合检查通过；以后漏翻译会在 CI 失败。 |
| EV-035 | 为 LDAP 配置装配增加非法端口错误上下文测试，并验证 `Run` 在非法配置下不 panic。 | 该失败路径已通过；数据库读取和 LDAP 连接失败矩阵仍在路线图中。 |
| EV-020 | 为入站列表保存节点筛选和搜索词，并在重新进入页面时恢复；节点删除后自动清理无效筛选。 | 入站列表刷新后保留运维筛选上下文，清空搜索仍可回到默认结果。 |
| EV-009 / EV-014 | 新增受认证保护的低基数 Prometheus 指标端点和可下载 alert rules；规则引用由测试校验。 | `/panel/api/server/metrics` 输出稳定指标，`/panel/api/server/alertRules` 下载匹配规则；monitor token 可用。 |
| EV-012 | 为成功生成的数据库备份增加可选本地快照保留，默认关闭，按数量/天数清理。 | 使用 0600、fsync、原子重命名和前缀隔离；默认不创建目录，失败只告警不影响下载。 |
| EV-016 | 为 raw、JSON、Clash 订阅增加 ETag 和条件请求。 | 内容未变返回 304，内容变化会生成新 ETag，不缓存敏感正文。 |
| EV-019 | 为 lazy 页面增加页面级错误边界。 | 渲染异常时保留侧栏，提供重试和返回首页操作。 |
| EV-029 | 复核已有真实路由与 endpoint registry 契约测试，并登记本轮新增观测端点。 | 检查缺失和多余登记通过；没有重复实现已有能力。 |

本轮新增或完成复核的功能项共涉及前端页面、订阅服务、观测端点、备份服务、LDAP 测试和翻译修改；新增 locale key CI 检查、Prometheus 规则文档和页面错误边界测试。本轮新增四个总控文档和一份监控 runbook，并更新 README 与 CHANGELOG。新增了两个受保护的 GET API，已登记 endpoint registry 并重新生成 OpenAPI；没有新增数据库字段或配置迁移。

## 路线图完成度

路线图共 46 项：

- 已实施：EV-001、EV-002、EV-009、EV-012、EV-014、EV-016、EV-019、EV-020、EV-030。
- 已具备并复核：EV-029（基线已有真实路由契约测试，本轮验证新增端点登记）。
- 部分实施：EV-035。
- 已规划：EV-003 至 EV-008、EV-010 至 EV-011、EV-013、EV-015、EV-017 至 EV-018、EV-021 至 EV-028、EV-031 至 EV-034、EV-036 至 EV-046，共 35 项。
- 下一批 P0/P1：签名更新 manifest、SBOM/provenance、更新前健康门禁、cron 任务状态、导入 dry-run、统一配置应用结果。

未把设计、竞品观察或本地静态检查写成“已完成”功能；每项候选升级都在路线图中保留了模块、证据、实现边界、难度和验收标准。

## 验证结果

| 检查 | 结果 | 说明 |
|---|---|---|
| `cd frontend && npm run typecheck` | PASS | TypeScript 严格检查通过。 |
| `cd frontend && npm run lint` | PASS | 前端 lint 通过。 |
| `cd frontend && npm run test` | PASS | 87 个测试文件、978 个测试通过；jsdom 输出了既有的 `getComputedStyle` 提示，没有失败。 |
| `cd frontend && npm run build` | PASS | Vite 生产构建通过；OpenAPI 生成结果为 172 条路径、176 个 operation。 |
| 翻译 JSON 解析 | PASS | 13 个 locale 全部可解析，复制分享链接 key 已补齐。 |
| `npm run check:translations` | PASS | 13 个 locale、1,949 个 key 路径，检查缺失 key、额外 key 和占位符不匹配。 |
| `gofmt -d internal/web/job/ldap_sync_job.go` | PASS | 修改后的 Go 文件格式无差异。 |
| `git diff --check` | PASS | 无空白错误。 |
| `go test -shuffle=on -count=1 ./...` | PASS | Go 1.27.4、CGO/SQLite 环境下全量通过；依赖使用 `GOPROXY=https://goproxy.cn,direct`。 |
| `go test -race -shuffle=on -count=1 ./...` | PASS | 全量 race 检查通过。 |
| `go build ./...` | PASS | Go 后端构建通过。 |
| `golangci-lint run` | PASS | golangci-lint 2.15.0，0 issues。 |
| `npm run build-storybook` | PASS | Storybook 10.6.6 静态构建通过；Vite 仅输出既有大 chunk 提示。 |
| `npm audit --omit=dev --audit-level=high` | PASS | 官方 npm registry 返回 3 个 low severity 依赖问题，没有 high/critical；修复需要升级 `swagger-ui-react` 并接受 breaking change，因此本轮未强制改依赖。 |
| `make verify` | NOT RUN | 当前 PowerShell 环境没有 `make` 命令；gen 已执行且可重复，Zod 生成物干净，OpenAPI 的工作树 diff 仅包含本轮新增端点；lint、typecheck、test、build 和 Storybook 已分别通过。 |
| 临时数据库本地运行 smoke | PARTIAL PASS | `go run .` 在临时数据库、临时端口 `29877` 上启动；`/panel/` 和 `/` 返回 200，随后可停止。源码检出缺少运行时 `bin` 资产，Xray 子进程无法启动，因此不宣称 Xray 或生产端到端通过。 |
| `go test -run 'TestRouteRegistryContract|TestPrometheus' -count=1 ./internal/web ./internal/web/controller` | PASS | 新增观测端点已进入路由契约和控制器测试。 |
| `go test -race -shuffle=on -count=1 -run 'TestPersistLocalBackup|TestBackupRetentionConfig' ./internal/web/service` | PASS | 本地备份保留策略的数量、年龄、权限、并发和路径隔离测试通过。 |
| `go test -race -count=1 ./internal/sub/...` | PASS | 订阅 ETag 条件请求回归通过。 |
| `npm run gen` | PASS | OpenAPI 重新生成，当前 172 条路径、176 个 operation；新增非 JSON response content type 与两个观测端点。 |

## 二次审计

实施后重新检查了 action 分发、locale key、生成文件、占位提示、格式和工作树：

- `copyLinks` 已有明确分支和依赖，不再落入未完成提示。
- 13 个语言包均包含复制分享链接文案。
- 新增 metrics/alertRules 已登记到 endpoint registry，生成的 OpenAPI 与当前定义一致。
- `frontend/src/generated` 未发生非预期变化；`frontend/public/openapi.json` 的变化仅来自本轮新增端点和文本响应类型。
- 未新增数据库模型或迁移需求；新增的两个 API 已登记并生成 OpenAPI。
- 仓库仍处于 `main`，没有提交、推送、发布或部署操作。
- OpenAPI/Zod 生成已执行；生成过程可重复，当前 OpenAPI 变更均对应本轮新增端点和响应类型。
- 临时运行验证确认面板 HTTP 可达并能干净停止；Xray 子进程因源码检出缺少 `bin` 运行时资产而未启动。
- 备份保留策略默认关闭；只有显式设置 `XUI_BACKUP_RETENTION_COUNT` 或 `XUI_BACKUP_RETENTION_DAYS` 时才会写入 `${XUI_DB_FOLDER}/backups`。

## 尚未宣称的能力

本轮没有在 Linux/systemd、真实远程节点、浏览器登录态、生产数据库、签名发布资产或真实升级回滚环境中执行端到端验证。因此报告不把本地构建通过写成生产发布通过，也不把竞品功能写成本项目已经具备的功能。

Go 1.27.4、CGO/SQLite 和备用依赖代理已经补齐，LDAP job 与完整后端回归均已通过。EV-035 仍保留后续更细粒度的 LDAP 配置错误注入测试任务。

## 建议执行顺序

1. 在提供 `make` 命令的 Linux/CI 环境补跑完整 `make verify`。
2. 实施 EV-023 至 EV-026，建立更新签名、SBOM、健康门禁和可追踪更新状态。
3. 实施 EV-010、EV-027、EV-032，统一指标、探针和定时任务状态。
4. 实施 EV-003、EV-004、EV-007、EV-008，降低批量操作和多节点同步的误操作成本。
5. 在每批改动后重复前端检查、Go 检查、生成文件检查和发布 smoke matrix。
