# Xray 服务器管理面板（3X-UI 二次维护版）

> Go + React 打造的一体化 Xray-core Web 管理后台：入站、客户端、订阅、节点、路由、通知、监控，一个页面全部搞定。

## 一、项目简介

本仓库托管于 Gitee，是在上游开源项目 3X-UI（及其更早的前身 x-ui）基础上继续维护的二次修改分支。仓库内的安装脚本、更新脚本和文档链接已统一指向当前仓库。

- 仓库地址：<https://gitee.com/yuchen0x1/Xray-server-management-panel>
- 克隆命令：`git clone https://gitee.com/yuchen0x1/Xray-server-management-panel.git`
- 安装脚本只拉取仓库已发布的 Release 与 raw 资源。如需验证本地未发布的改动，请使用 Docker 本地构建或源码构建。

请仅在你拥有授权的服务器、网络和业务场景中使用本项目。

## 二、目录

1. 项目简介
2. 目录
3. 功能亮点
4. 技术栈
5. 快速部署
6. 面板使用
7. 命令行工具
8. API 接口
9. 订阅服务
10. 环境变量
11. 数据库管理
12. 本地开发与构建
13. 目录结构
14. 常见问题
15. 安全提醒
16. 交流与反馈
17. 参与贡献
18. 更新日志说明
19. 开源协议

## 三、功能亮点

### 协议与传输

- 入站协议：VLESS、VMess、Trojan、Shadowsocks、WireGuard、Hysteria、HTTP、Mixed、Tunnel、TUN、MTProto。
- 传输与安全：TCP Raw、mKCP、WebSocket、gRPC、HTTPUpgrade、XHTTP、TLS、REALITY、Fallback、Header、Mux 等 Xray 常用配置。
- Xray 全局配置：出站、路由规则、Balancer、Warp / Nord 相关配置、配置测试，以及 Xray 安装 / 重启 / 停止。

### 客户端与订阅

- 客户端全生命周期：流量配额、到期时间、启用状态、批量创建、批量调整、批量启停、批量删除、分组、在线状态、IP 记录、分享链接和二维码。
- 订阅输出：Raw、Xray JSON、Clash / Mihomo，自定义订阅页面和模板，按 User-Agent 自动返回匹配格式。

### 多节点与运维

- 节点协同：远程节点列表、健康状态、证书、节点升级、节点侧入站与订阅链接处理。
- 运维监控：系统状态、CPU / 内存 / 网络历史、Xray observatory 指标、日志、数据库备份恢复、证书工具、Fail2ban、Telegram 与邮件通知。

### 开发者友好

- 面板内置 Swagger UI 并生成 OpenAPI JSON，支持登录会话或 Bearer Token 调用 `/panel/api/*`。
- 数据库同时支持 SQLite 与 PostgreSQL：默认 SQLite，生产或多实例场景可切换 PostgreSQL。
- 前端多语言、浅色 / 深色 / 极暗主题，附带 Storybook 与 Vitest 测试。

## 四、技术栈

- **后端**：Go 1.26.6、Gin、GORM、SQLite、PostgreSQL、go-i18n、gopsutil、gorilla/websocket
- **代理核心**：Xray-core，含 MTProto sidecar 支持
- **前端**：React 19、TypeScript 6、Vite 8、Ant Design 6、React Router、TanStack Query、Zod、i18next
- **API / Schema**：由 Go 源码生成 OpenAPI、TypeScript 类型和 Zod Schema
- **文档站**：Next.js 16、Fumadocs、pnpm
- **部署形态**：Docker、Docker Compose、systemd / OpenRC、cloud-init

## 五、快速部署

以下四种方式任选其一。部署完成后参考「面板使用」进入后台。

### 5.1 一键脚本安装

环境要求：

- Linux 服务器，root 权限。
- 支持的发行版：Debian / Ubuntu / Armbian、Fedora / RHEL / Alma / Rocky / Oracle / Amazon Linux、CentOS、Arch / Manjaro / Parch、openSUSE、Alpine。
- 需要可用的 `curl`、`tar`、系统服务管理器，以及能访问 Gitee 仓库 Release 与 raw 资源的网络环境。

执行安装：

```bash
bash <(curl -Ls https://gitee.com/yuchen0x1/Xray-server-management-panel/raw/main/install.sh)
```

脚本会创建面板服务，生成随机用户名 / 密码 / 端口 / Web base path / API token，并把安装结果写入 `/etc/x-ui/install-result.env`（文件权限 `600`）。常见字段包括：

```text
XUI_USERNAME=...
XUI_PASSWORD=...
XUI_PANEL_PORT=...
XUI_WEB_BASE_PATH=...
XUI_ACCESS_URL=...
XUI_API_TOKEN=...
XUI_DB_TYPE=...
```

注意：安装脚本会从当前 Gitee 仓库的 Release 和 raw 资源获取文件。若要验证本地未发布改动，优先使用 Docker 本地构建或源码构建。

### 5.2 Docker Compose 本地构建

环境要求：Docker 与 Docker Compose v2，默认监听 `2053` 端口。

```bash
git clone https://gitee.com/yuchen0x1/Xray-server-management-panel.git
cd Xray-server-management-panel
docker compose up -d --build
```

默认使用 SQLite。数据目录映射到当前目录的 `db/`，证书目录映射到 `cert/`，acme.sh 状态映射到 `acme/`。

如果只是在已经克隆好的仓库中启动：

```bash
docker compose up -d --build
```

若启用 Fail2ban IP 限制，请在 Compose 中保留 `NET_ADMIN` 和 `NET_RAW` capability。

### 5.3 Docker + PostgreSQL

Compose 文件内提供了 `postgres` profile，但启动 profile 只会创建 PostgreSQL 容器；面板是否使用 PostgreSQL 由 `XUI_DB_TYPE` 和 `XUI_DB_DSN` 决定。

1. 在 `docker-compose.yml` 中启用这两项环境变量：

```yaml
XUI_DB_TYPE: "postgres"
XUI_DB_DSN: "postgres://xui:xui@postgres:5432/xui?sslmode=disable"
```

2. 启动服务：

```bash
docker compose --profile postgres up -d --build
```

如果修改 `XUI_PORT`，也要同步修改 Compose 的端口映射，例如 `8080:8080`。

### 5.4 无人值守安装

```bash
XUI_NONINTERACTIVE=1 bash <(curl -Ls https://gitee.com/yuchen0x1/Xray-server-management-panel/raw/main/install.sh)
```

非交互模式适合 cloud-init 或自动化部署。安装完成后同样读取 `/etc/x-ui/install-result.env` 获取访问地址和凭据。

## 六、面板使用

1. 打开安装结果中的 `XUI_ACCESS_URL`，或访问 `http://<server-ip>:<port>/<web-base-path>`。
2. 使用安装时生成的用户名和密码登录。
3. 在面板中管理入站、客户端、分组、节点、Host 覆盖、路由、出站、订阅和系统设置。
4. 需要命令行维护时运行 `x-ui` 菜单或相关子命令。
5. 需要自动化调用时，在设置中获取 API token，并使用 Bearer Token 调用 `/panel/api/*`。

面板主要页面：

- 登录页：`/`（或配置后的 Web base path 根路径）
- 入站：`/panel/inbounds`
- 客户端：`/panel/clients`
- 客户端分组：`/panel/groups`
- 节点：`/panel/nodes`
- Host 覆盖：`/panel/hosts`
- 设置：`/panel/settings`
- Xray / 出站 / 路由：`/panel/xray`、`/panel/outbound`、`/panel/routing`
- API 文档：`/panel/api-docs`

## 七、命令行工具

Linux 脚本安装后，`/usr/bin/x-ui` 提供交互式菜单和常用服务子命令：

```bash
x-ui
x-ui start
x-ui stop
x-ui restart
x-ui restart-xray
x-ui status
x-ui settings
x-ui enable
x-ui disable
x-ui log
x-ui banlog
x-ui update
x-ui update-dev
x-ui legacy
x-ui install
x-ui uninstall
x-ui setup-fail2ban
x-ui update-all-geofiles
x-ui migrateDB
x-ui pgclient
```

Go 二进制本身也提供若干子命令，适合源码构建或容器内维护：

```bash
./x-ui run
./x-ui migrate
./x-ui migrate-db -dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
./x-ui setting -show
./x-ui setting -reset
./x-ui setting -getApiToken
./x-ui cert
```

数据库迁移、凭据重置、端口、Web base path、证书、Telegram 和 API token 等设置，可通过 `./x-ui setting` 或 `x-ui` 菜单完成。

## 八、API 接口

面板 API 默认挂载在 `/panel/api/*`，支持两种鉴权方式：浏览器登录会话，或请求头携带 `Authorization: Bearer <token>`。

常用入口：

- `/panel/api/openapi.json`：生成的 OpenAPI 3 JSON
- `/panel/api-docs`：面板内 Swagger UI
- `/panel/api/inbounds/*`：入站和入站客户端 API
- `/panel/api/clients/*`：全局客户端、分组、流量、在线状态、分享链接 API
- `/panel/api/server/*`：系统状态、日志、证书、更新、Xray 操作 API
- `/panel/api/nodes/*`：多节点 API
- `/panel/api/hosts/*`：Host 覆盖 API
- `/panel/api/setting/*`：面板设置 API
- `/panel/api/xray/*`：Xray 配置、出站、路由、配置测试 API

脚本或第三方客户端建议使用 API token。请避免在日志、Issue 或截图中暴露 token、订阅链接、私钥和密码。

## 九、订阅服务

订阅服务由面板设置控制，可提供标准 Raw、Xray JSON 和 Clash / Mihomo 输出。代码默认路径：

- Raw 订阅：`/sub/:subid`
- Xray JSON：`/json/:subid`
- Clash / Mihomo：`/clash/:subid`

订阅服务支持 HEAD 请求、流量 / 到期响应头、HTML 信息页、自定义模板目录、User-Agent 自动识别，以及外部链接合并。实际监听端口、路径开关、自动识别规则和模板目录以面板 Settings -> Subscription 中的配置为准。

## 十、环境变量

默认配置来自数据库与运行环境。常用环境变量如下，完整说明请参阅 `docs/content/docs/en/reference/env-vars.mdx`。

| 变量 | 说明 |
| --- | --- |
| `XUI_LOG_LEVEL` | 日志级别 |
| `XUI_DEBUG` | 调试模式；开发时可从磁盘读取前端 dist |
| `XUI_SKIP_HSTS` | 跳过 HSTS 响应头 |
| `XUI_SECURE_COOKIE` | 强制安全 Cookie 行为 |
| `XUI_PORT` | 面板监听端口 |
| `XUI_BIN_FOLDER` | Xray / sidecar 二进制目录 |
| `XUI_DB_FOLDER` | SQLite 数据库目录 |
| `XUI_DB_TYPE` | 数据库类型；空值 / 未设置为 SQLite，`postgres` 为 PostgreSQL |
| `XUI_DB_DSN` | PostgreSQL DSN |
| `XUI_LOG_FOLDER` | 日志目录 |
| `XUI_INIT_WEB_BASE_PATH` | 首次初始化 Web base path |
| `XUI_ENABLE_FAIL2BAN` | 是否启用 Fail2ban 相关配置 |
| `XUI_NONINTERACTIVE` | 安装脚本非交互模式 |
| `XUI_GOGC` | Go GC 调优 |
| `XUI_MEMORY_RELEASE_INTERVAL` | 定期释放内存的间隔，单位分钟 |
| `XUI_MEMORY_LIMIT` | Go 内存软限制，单位 MiB |
| `GOMEMLIMIT` | Go 原生内存限制表达式，优先级高于 `XUI_MEMORY_LIMIT` |
| `XUI_PPROF` | 在本地地址暴露 pprof |
| `XUI_SOURCEMAP` | 前端构建时生成 sourcemap |
| `XRAY_VMESS_AEAD_FORCED` | Xray VMess AEAD 兼容设置 |

本地开发可参考 `.env.example`，Docker 部署可参考 `docker-compose.yml` 中的注释。

## 十一、数据库管理

默认数据库为 SQLite：

- Linux 脚本安装默认路径：`/etc/x-ui/x-ui.db`。
- Docker 默认挂载：`./db/` -> `/etc/x-ui/`。
- Windows 和本地开发路径由配置与运行目录决定。

切换到 PostgreSQL：

```bash
export XUI_DB_TYPE=postgres
export XUI_DB_DSN="postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
```

迁移示例：

```bash
x-ui migrate-db -dsn "postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable"
```

如需导出或恢复，请查看 `main.go` 中 `migrate-db` 子命令的 `-dump`、`-restore`、`-out` 参数，以及 `docs/content/docs/en/reference/database.mdx`。

## 十二、本地开发与构建

开发环境要求：

- Go `1.26.6`。
- Node.js `>=24.0.0`，npm `>=10.0.0`，建议使用仓库 `.nvmrc`。
- 运行 SQLite 相关 Go 测试或 race detector 时，需要可用的 C 编译器。
- 文档站使用 pnpm，`docs/package.json` 声明 `pnpm@11.15.1`。

后端默认监听 `2053`。前端 Vite 开发服务器固定使用 `5173`，并把面板 API、登录、WebSocket 等请求代理到 `http://localhost:2053`。

```bash
# 终端 1：启动 Go 后端
go run main.go

# 终端 2：启动前端开发服务器
cd frontend
npm install
npm run dev
```

前端构建会生成三个入口：`index.html`、`login.html`、`subpage.html`，输出到 `internal/web/dist/`。生产运行时 Go 会通过 `embed.FS` 嵌入并提供这些文件。

### 构建与检查

仓库根目录提供 Makefile，`make verify` 对齐本地 CI 风格检查：代码生成、Go / 前端 lint、TypeScript typecheck、Go / 前端测试、完整构建和 Storybook 构建。

```bash
make help
make gen
make lint
make typecheck
make test
make build
make build-storybook
make verify
```

也可以分别运行：

```bash
make test-go
make test-fe
make race
make vulncheck
```

前端单独命令：

```bash
cd frontend
npm run gen
npm run lint
npm run typecheck
npm test
npm run build
npm run storybook
```

文档站命令：

```bash
cd docs
pnpm install
pnpm dev
pnpm build
pnpm typecheck
pnpm lint
pnpm test
```

## 十三、目录结构

| 路径 | 说明 |
| --- | --- |
| `main.go` | 程序入口、Web 服务启动、数据库迁移、设置和证书子命令 |
| `internal/config/` | 版本、路径、环境变量和运行配置 |
| `internal/database/` | GORM 模型、迁移、SQLite / PostgreSQL 支持 |
| `internal/web/` | Gin Web 服务、控制器、中间件、会话、API 和业务服务 |
| `internal/sub/` | Raw / JSON / Clash 订阅服务、模板和订阅页面数据 |
| `internal/xray/` | Xray 配置生成、生命周期、API、热更新和指标 |
| `frontend/` | React 面板、API 客户端、页面、组件、Schema、Storybook、测试 |
| `frontend/src/generated/` | 由 `tools/openapigen` 生成的类型、Schema 和示例 |
| `tools/openapigen/` | OpenAPI、TypeScript 类型和 Zod Schema 生成器 |
| `docs/` | Next.js / Fumadocs 文档站 |
| `deploy/` | cloud-init、云市场和部署冒烟测试相关文件 |
| `media/` | 项目图片和截图资源 |
| `windows_files/` | Windows 平台辅助文件与说明 |
| `docker-compose.yml` | Docker Compose 本地构建和可选 PostgreSQL profile |
| `Dockerfile` | 多阶段容器构建：前端、Go 二进制和最终运行镜像 |

## 十四、常见问题

**Q：Docker 启动后访问不到面板？**

确认容器端口映射和 `XUI_PORT` 一致。默认映射是 `2053:2053`；如果设置 `XUI_PORT=8080`，Compose 端口也要改为 `8080:8080`。

**Q：启动 PostgreSQL profile 后仍在使用 SQLite？**

这是预期行为。`docker compose --profile postgres up -d` 只启动数据库容器；还需要启用 `XUI_DB_TYPE=postgres` 和 `XUI_DB_DSN=...`。

**Q：前端开发服务器提示后端不可达？**

先在仓库根目录启动 Go 后端：`go run main.go`，然后在 `frontend/` 中运行 `npm run dev`。

**Q：Go 测试或 race detector 因 SQLite 编译失败？**

安装 C 编译器和对应平台的构建工具后重试。`make race` 也需要 C 编译器。

**Q：`/panel/api/openapi.json` 返回 404？**

确认已经构建前端或运行了代码生成。生产构建会把 `openapi.json` 放入 `internal/web/dist/` 并嵌入 Go 二进制。

**Q：脚本安装出来的不是我本地改过的代码？**

脚本安装依赖发布包和 raw 资源，不等同于把本地未发布源码编译安装。验证当前分支改动时，请使用 `docker compose up -d --build` 或 `make build` 后运行生成的二进制。

## 十五、安全提醒

- 不要在公网以明文 HTTP 直接暴露管理面板；建议使用 HTTPS、反向代理、SSH 隧道或可信网络访问控制。
- 安装后立即保存随机生成的凭据，并定期轮换密码和 API token。
- 开启并验证 Fail2ban / IP 限制前，确保容器或宿主机具备实际封禁所需权限。
- 不要把数据库、订阅链接、API token、私钥、证书、Telegram token、邮件凭据提交到仓库或 Issue。
- 调用远程 URL、导入外部配置或使用 Reality 扫描功能前，应确认目标属于你授权测试的范围。
- 发布问题报告时请附带版本、系统、部署方式和脱敏日志。

## 十六、交流与反馈

- QQ 群：`1016190748`
- Telegram：<https://t.me/lenmoC>

反馈问题时请说明是否只出现在当前二改分支，并附上复现步骤、日志和部署方式。

## 十七、参与贡献

欢迎提交 Issue 和 Pull Request。提交前建议至少运行：

```bash
make gen
make lint
make typecheck
make test
```

较大改动或准备合并前建议运行：

```bash
make verify
```

涉及前端 API 类型、Zod Schema 或 OpenAPI 的改动，请运行 `make gen` 并提交生成文件。

## 十八、更新日志说明

当前仓库未维护独立 `CHANGELOG.md`。请通过 Git 提交历史、Release 页面或上游项目记录追踪变更；不要假设本分支与上游发布内容完全一致。

## 十九、开源协议

本项目采用 GNU General Public License v3.0 or later 授权，完整文本见 [LICENSE](LICENSE)。
