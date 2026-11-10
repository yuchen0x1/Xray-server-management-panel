# 项目审计与本地升级报告

审计日期：2026-11-09（Asia/Shanghai）  
仓库：`C:\Users\lenmo\Desktop\Xray-server-management-panel`  
审计方式：本地只读检查、有限修正、自动化验证、GitHub 仓库元数据与源码研究。未提交、推送、发布或上传任何内容。

## 1. 项目现状

### 项目与环境

| 项目 | 当前情况 |
|---|---|
| 项目名称 / 用途 | 3x-ui 维护分支，管理 Xray-core 代理服务、客户端和订阅 |
| 类型 | Go Web/API 服务 + React SPA；本次没有发现桌面或移动应用子项目 |
| 后端 | Go 1.27.3、Gin、GORM、Xray-core、gRPC、cron |
| 前端 | React 19、Ant Design 6、TypeScript、Vite 8、TanStack Query、Zod |
| 存储 | SQLite 默认，PostgreSQL 可选 |
| 构建 | Go modules；frontend 使用 npm + package-lock；Vite 构建后嵌入 Go `embed.FS` |
| 应用版本 | `internal/config/version` 为 3.8.0；前端 package 版本为 0.4.3 |
| Git | `main`，HEAD `654ee9a2`（2026-11-08）；审计开始时工作区干净 |
| 项目文件 | 基线 1,520 个 Git 跟踪文件；509 个 Go 文件、297 个 Go 测试文件 |
| 依赖基线 | Go 35 个直接 require；npm 21 个生产依赖、27 个开发依赖；lockfile 909 个 package 项 |
| 运行状态 | 审计开始时没有发现本项目监听端口；未启动本项目服务 |

当前目录分工清晰：`internal/database` 管理模型与迁移；`internal/xray` 管理 Xray 生命周期和 gRPC；`internal/mtproto` 管理 MTProto 子进程；`internal/sub` 提供订阅；`internal/web/controller`、`service`、`runtime` 分别承接 HTTP、业务逻辑和本地/远端节点运行时；`frontend/src` 提供管理 SPA；`tools/openapigen` 从 Go 类型生成前端 schema 和类型。请求主要按 controller → service → database/runtime 分层，订阅服务与管理 API 分开。

### 功能清单（按实际代码与项目文档核对）

| 能力 | 状态与实现位置 |
|---|---|
| 入站、客户端 CRUD、批量管理、分组、限额、到期、流量/IP 限制 | 已实现；`internal/web/service/inbound*.go`、`client*.go`、对应 frontend pages |
| Xray 配置生成、校验、热更新、生命周期监控 | 已实现；`internal/xray`、`internal/web/service/xray*.go` |
| 多节点、远端 runtime、同步、mTLS/TLS 模式 | 已实现；`internal/web/runtime`、`internal/web/service/node*.go` |
| Raw/JSON/Clash 订阅、订阅页与模板 | 已实现；`internal/sub` |
| Hosts 覆盖、外部订阅、出站和路由 | 已实现；`internal/web/service/host.go`、`outbound*`、`internal/sub` |
| WARP、Nord、PIA 等集成 | 已实现；`internal/web/service/integration` |
| Telegram bot、邮件/Webhook 通知、系统告警 | 已实现；`service/tgbot`、`service/email`、`eventbus`、`service/webhook` |
| MTProto 多 secret/ad-tag/配额热更新 | 已实现；`internal/mtproto` 与客户端/入站 service |
| API token、2FA、CSRF、会话及权限 | 已实现；`controller`、`middleware`、`service/panel` |
| UI 多语言、主题、图表、Storybook、API 文档 | 已实现；`frontend/src`、13 个后端 locale、`frontend/src/pages/api-docs` |
| SQLite/PostgreSQL 模式与迁移、导入导出、CLI | 已实现；`internal/database`、`internal/web/service/server.go`、`main.go` |

上述状态来自代码、路由和已有测试的存在性核查；没有逐个在真实 Xray 服务器、远端节点、Telegram、邮件服务器或 PostgreSQL 环境上端到端操作，因此不能视作生产环境验收。

## 2. 清理与依赖记录

本次未删除文件、依赖、数据库或用户数据。仓库没有可安全判定为垃圾且不影响项目/发布流程的已跟踪文件，因此不为减少文件数而删除内容。

| 记录 | 结果 |
|---|---|
| 跟踪文件数 | 基线 1,520；最终增 1 份报告，未删除文件 |
| Go 直接依赖 | 35 → 35；`go mod tidy -diff` 无差异 |
| npm 直接依赖 | 生产 21 + 开发 27 → 未变化 |
| 删除文件/依赖 | 无 |
| 保留项 | `windows_files/SSL/Win64OpenSSL_Light-3_6_0.exe` 为跟踪的安装资源；忽略规则命中的 `x-ui-local.exe`、`system_metrics.gob` 是本地构建/运行数据，不作删除 |
| 生成产物 | Vite `internal/web/dist`、Storybook `frontend/storybook-static` 为忽略目录，未纳入 Git |

`npm audit --registry=https://registry.npmjs.org` 报告 6 项：3 low、3 moderate、没有 high/critical。问题链涉及 Storybook 的 `glob`（GHSA-3g3m-2fjr-6pw4）和 Swagger UI 的 `deepmerge`（GHSA-gcqx-w48m-w2h4）。npm 给出的自动修复方案涉及主版本降级（Storybook 10 → 8、Swagger UI 5 → 4），未采用；建议先验证兼容版本或使用安全的传递依赖覆盖，并运行 Storybook/API 文档回归。

`govulncheck -show verbose ./...` 没有报告可达代码漏洞，但报告 3 项模块级公告：`golang.org/x/net@v0.61.0` 的 GO-2026-6866、GO-2026-6860（修复版本 v0.62.0）；`golang.org/x/crypto@v0.58.0` 的 GO-2026-5932（扫描数据库列出修复版本 `v0.58.1-0.20261009181718-2f1f834330b2`）。包级分析均未发现调用路径；仍应安排定向版本升级及测试，不能据此宣称依赖完全无风险。

Windows 环境初始 `CGO_ENABLED=0`，导致 SQLite 测试无法运行。本机没有 C 编译器。为完成本地验证，使用 winget 将 LLVM-MinGW 安装到 `D:\DevTools\llvm-mingw`，只在 Go 测试/构建进程的 PATH 中临时加入编译器。安装器新增的用户 PATH 项随后已精确移除；工具链保留在 D 盘，未修改项目或系统级配置。

端口记录：审计开始时没有发现本项目的 2053/5173/6006 监听。最终复查时 `localhost:5173` 被 PID 54760 占用；进程命令行指向 `C:\Users\lenmo\Desktop\Doomsday Survival Guide\apps\web`，属于其他项目。该进程未由本任务启动，也未被停止。审计没有启动服务或占用端口。

## 3. GitHub 竞品研究

以下数据由 GitHub CLI 仓库元数据和公开 README 获取，时间为 2026-11-09；Star 数会变化。研究重点是产品设计和功能，不复制代码。

| 仓库 | Stars / 最近推送 | 许可证 | 值得参考的实际设计 |
|---|---:|---|---|
| [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) | 49,586 / 2026-11-06 | GPL-3.0 | 本项目直接上游；多协议、客户端限额、订阅、多节点、REST API、国际化、发布校验和文档 |
| [Gozargah/Marzban](https://github.com/Gozargah/Marzban) | 7,514 / 2026-11-02 | AGPL-3.0 | 多管理员/节点、面向 API 的管理、用户多协议订阅、清晰的文档入口和备份/机器人能力 |
| [hiddify/Hiddify-Manager](https://github.com/hiddify/Hiddify-Manager) | 9,511 / 2026-11-06 | GPL-3.0 | 用户自助配置页、自动备份、自动更新、多管理员、以安装和连接流程为中心的体验 |

比较结论：当前项目在 Xray 协议广度、多节点同步、订阅输出、单二进制部署、强类型 API/schema 生成及前端测试方面已有很完整的实现。Marzban 值得长期关注其文档/API 使用体验和管理员工作流；Hiddify 值得关注其用户自助流程、安装/备份/更新体验。上述功能类别在当前项目里大多已有对应基础，不应简单移植竞品实现。可以进一步核对用户自助页面、备份恢复的操作引导和部署升级的可观测性；插件系统、多核心统一抽象等会显著扩大范围，暂不建议仅为追赶竞品而加入。

**维护状态提示：** 本地 `main` 与已缓存的 `upstream/main` 提交图统计为本地领先 3,298、落后 316。这个数字来自当前本地 remote-tracking refs，不保证 refs 是最新远端状态；它表明双方差异很大，后续上游同步应以明确的基线、功能对照和逐段审查完成，不宜直接合并或重置。

## 4. 发现的问题与差距

### 修复项

1. `frontend/vite.config.js` 多处使用 `__dirname`，`frontend/vitest.config.ts` 还保留兼容分支。在 Node 24/Vite 8 的原有构建中产生 `configLoader: 'native'` 不兼容警告。已统一改为 `import.meta.dirname`，避免依赖旧 CommonJS 全局变量。

### 仍需跟进

1. npm 审计存在 6 项低/中危传递依赖公告；当前没有采用会造成主版本降级的自动修复。
2. Go 审计存在 3 项模块级公告，虽然未发现可达调用路径，仍需针对 `x/net` 与 `x/crypto` 验证更新。
3. Storybook 输出包含约 1.33 MB 的单个 iframe bundle，构建成功但高于通用 500 KB 提示阈值。该目录为组件工作台，不是用户生产面板；可以考虑拆分 Storybook 预览代码。
4. 本机进程没有启动项目服务，本次没有做登录、表单提交、代理进程启动、远端节点、实际浏览器交互或性能压测。
5. Windows Go 测试需要 CGO。D 盘本地工具链已能支持本次测试；新终端运行时需显式配置 PATH/CC 或自行安装 C 编译器。

### 本次未发现的事项

- `go mod tidy -diff` 未发现 Go 模块声明与源码不一致。
- 路由和 API 文档生成器按现有配置成功；OpenAPI 生成 170 paths、174 operations。
- 英文 locale 的键检查未发现 13 个 locale 间的顶层键缺失；该检查不代替翻译质量审阅。
- TODO/FIXME 搜索仅发现少数有上下文的待办（如 Telegram bot 重启按钮），没有据此删除实现或资源。
- 未发现可安全认定的无用依赖；未做全量依赖升级。

## 5. 功能、架构与 UI/UX 升级

- 新增功能：无。审计未发现值得在没有端到端验收条件时直接加入的核心能力。
- 修复：Vite/Vitest 配置改用标准 ESM 路径解析，覆盖前端测试、开发和构建配置。
- 架构：未进行大范围重构。项目已有 controller/service/runtime/database 分层和 Go → Zod/OpenAPI 的类型生成约束，贸然调整会带来较大回归面。
- UI/UX：未改变业务页面或视觉体系。本次没有浏览器交互或用户研究证据支持整体重设计。
- 依赖：没有增加、删除或升级应用依赖。
- 清理：没有删除文件或用户数据。

## 6. 测试与质量验证

| 检查 | 结果 |
|---|---|
| `go test -shuffle=on -count=1 ./...`（Windows、CGO=1、LLVM-MinGW） | 通过 |
| 首次 Go 测试（CGO=0） | 失败；SQLite driver 被编译为 stub。修正本进程环境后全套通过 |
| `go vet ./...` | 通过 |
| `go build ./...`（CGO=1） | 通过 |
| `npm run typecheck` | 通过 |
| `npm run lint` | 通过 |
| `npm run test` | 通过：84 test files、971 tests |
| `npm run build` | 通过；生成 170 paths、174 operations 的 OpenAPI。最大应用 vendor chunk 约 1.35 MB |
| `npm run build-storybook` | 通过；报告 iframe bundle 约 1.33 MB |
| `npm run gen` | 通过；生成 28 schemas；生成文件无差异 |
| `go mod tidy -diff` | 无差异 |
| `govulncheck -show verbose ./...` | 无可达代码漏洞；3 个未达模块公告，详见依赖记录 |
| `npm audit` | 6 项 low/moderate，详见依赖记录 |
| `golangci-lint run` | 通过；0 issues |
| `make verify` | 未运行；未找到常规 `make` 命令。其主要检查已分别运行，完整 Makefile 聚合门禁未执行 |
| 真实服务/浏览器端到端 | 未运行；没有启动后台服务或占用端口 |

Storybook/Vitest 输出还包含 jsdom 对 `getComputedStyle()` pseudo-elements 的提示和 Storybook 配置 `test.include` 的弃用提示；测试退出码为 0，84/971 通过。

## 7. 文件修改记录

| 文件 | 变更 |
|---|---|
| `frontend/vite.config.js` | 新增 `projectDir = import.meta.dirname` 并替换全部 `__dirname` 路径 |
| `frontend/vitest.config.ts` | 移除 `fileURLToPath`/旧兼容判断，使用 `import.meta.dirname` |
| `PROJECT_AUDIT_REPORT.md` | 新增本报告 |

未修改 Go 业务代码、依赖清单、锁文件、数据库、运行配置或生成文件。没有文件删除、移动、提交或推送。构建生成目录受 `.gitignore` 忽略。

## 8. 最终评价与建议

当前项目本身已具备成熟的多协议代理管理、订阅和多节点基础，层次清楚且自动化测试覆盖广。此次升级范围刻意限于可复现、低风险的 Node/Vite 配置兼容问题；没有为了满足“功能新增数量”而合并未经验证功能。

建议后续依次处理：

1. 对 `x/net`、`x/crypto` 做定向升级和 Go 全量测试，评估公告修复版本。
2. 为 npm 的 `glob`、`deepmerge` 选择兼容的修复版或安全 override，再运行 Storybook、API 文档和前端测试；不要采用审计器给出的主版本降级方案。
3. 以当前 Git 提交为审查基线，单独规划与上游的差异同步，避免一次性合并 316 个远端提交。
4. 在隔离测试数据库、项目专属临时端口和测试 Xray 二进制下补做真实登录、关键客户端流程、订阅和多节点端到端验收。
5. 将 Storybook 大包拆分作为独立性能改进，并用体积对比验证收益。

Git 最终状态应只有以上两处配置修改及本报告；审计过程未执行 `git push`、提交、release、远端 tag、部署或上传。

## 9. 续审与依赖安全增量（2026-11-09）

本节覆盖上次审计后实际执行的升级和复测；上文记录的基线扫描结果仍保留作为历史基线。

### 实际变更

| 文件 | 变更 | 原因与影响 |
|---|---|---|
| `frontend/package.json` | 对 `@joshwooding/vite-plugin-react-docgen-typescript` 定向覆盖 `glob` 到 `^14.0.2` | 只影响 Storybook 的 React docgen 构建依赖，不改变生产运行时依赖；插件使用 `glob` API。 |
| `frontend/package-lock.json` | 锁定 `glob@14.0.4` 及对应 `path-scurry@3`；移除 3 个原有 extraneous 锁记录 | `npm install` 重算锁文件；未移除 manifest 中依赖。 |
| `go.mod`、`go.sum` | `golang.org/x/net` 从 `v0.61.0` 升到 `v0.62.0`；`golang.org/x/crypto` 从 `v0.58.0` 升到 `v0.58.1-0.20261009181718-2f1f834330b2` | 分别对齐公告给出的修复版与预发布修复版。 |

未修改 UI 页面、业务 API、数据库结构或用户数据；未添加新依赖、未删除应用依赖，也没有做竞品代码移植。之前留存的 Vite/Vitest ESM 路径修正和本报告均保留。

### 安全扫描复核

- 使用官方 npm registry 运行 `npm audit`。此前 6 项（3 low、3 moderate）降为 3 项 low、0 moderate、0 high、0 critical。`glob` 公告链已通过 `14.0.4` 消除；剩余 3 项是 Swagger UI → swagger-client → `deepmerge@4.3.1` 的同一低危公告链。npm 提出的自动修复会降级 `swagger-ui-react` 到 4.1.3，存在主版本破坏风险，故保留 Swagger UI 5 并将该项记为未解决风险。
- Go 将 `x/net` 升到 `v0.62.0` 后重新执行 `govulncheck ./...`：`No vulnerabilities found`。
- Go `x/crypto/openpgp` 公告给出的修复版本是预发布版本；本轮已定向升级到该版本。当前版本仍包含历史 OpenPGP 包，但项目代码没有调用该包；最终复扫没有报告公告。
- `npm audit` 仍以退出码 1 报告上述 low 项，这是剩余漏洞告警，不代表安装、构建或测试失败。

### 本轮验证结果

| 检查 | 结果 |
|---|---|
| `npm run typecheck` | 通过 |
| `npm run lint` | 通过 |
| `npm test -- --reporter=dot` | 通过：84 个测试文件、971 项测试 |
| `npm run build` | 通过：OpenAPI 170 paths / 174 operations，Vite 生产构建成功 |
| `npm run build-storybook` | 通过；仍有约 1.33 MB iframe chunk 提示，暂未优化 |
| `npm ls glob deepmerge --all` | 通过依赖树检查：Storybook 使用 `glob@14.0.4`，Swagger 仍使用 `deepmerge@4.3.1` |
| `go test -shuffle=on -count=1 ./...` | 对最终 Go 依赖版本通过，全仓 Go 包测试成功 |
| `govulncheck ./...` | 对最终 Go 依赖版本通过：`No vulnerabilities found` |
| `go vet ./...` | 通过 |
| `go build ./...` | 通过 |
| `golangci-lint run` | 通过：0 issues |
| `npm audit --registry=https://registry.npmjs.org` | 未通过安全清零标准：剩 3 个 low 告警，具体原因见上节 |
| `git diff --check` | 通过；仅有 Git 提示换行符可能规范化的警告 |

此次没有启动后台服务或浏览器端到端测试，也没有开展基准性能测试；不能据此判断登录流程、真实节点/Xray 行为或不同屏幕布局已完成验收。

### 有依据的阶段性质量评估

这些分数是对代码与自动化验证现状的审查判断，不是用户体验调查或性能测量结果；UI 与兼容性因本轮未做浏览器验收而降低置信度。

| 维度 | 分数 | 依据 |
|---|---:|---|
| UI 设计 | 78 | React / Ant Design 组件体系完整；本轮未做全页面视觉验收。 |
| UX 体验 | 78 | 搜索、批量管理、国际化等既有能力较丰富；没有真实用户测试证据。 |
| 功能完整度 | 87 | 入站、客户端、订阅、多节点、通知及 API 管理覆盖面广；端到端场景未复测。 |
| 架构合理性 | 86 | controller/service/runtime/database 分层明确，自动生成 API schema；没有大规模架构评估。 |
| 代码质量 | 85 | Go 全量测试、前端 971 测试、lint/typecheck 通过；有遗留低危依赖告警。 |
| 性能 | 76 | 构建使用页面/模块拆分，但仍有大型 vendor 与 Storybook iframe chunk；没有运行时基准。 |
| 安全性 | 82 | Go 漏洞扫描无结果，glob 修复；npm 仍有 3 个 low 告警，未做渗透测试。 |
| 兼容性 | 78 | Go/前端构建覆盖 Windows 开发环境；浏览器矩阵、移动端和生产 Linux 未实测。 |
| 可维护性 | 84 | 生成流程、测试、分层较完整；旧依赖与 Storybook chunk 仍有改进空间。 |
| 可扩展性 | 85 | 多节点、协议、存储和集成模块已存在；没有验证新增插件等能力。 |

下一步优先级：找到不降级 Swagger UI 的 `deepmerge` 安全修复路径；另行量化并拆分大型 Storybook / vendor chunk；在隔离数据库和专用端口下补充真实登录、客户端管理、订阅、多节点的浏览器端到端验收。
