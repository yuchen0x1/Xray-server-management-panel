# 3X-UI

3X-UI 是一个用于管理 Xray-core 服务器的 Web 控制面板，提供协议配置、
客户端与流量管理、订阅服务、节点协同、路由工具和运行状态监控。

本仓库是基于原项目进行的**二次修改维护分支**，仅建议用于个人和合法用途。

## 项目归属与致谢

- 当前仓库：[yuchen0x1/Xray-server-management-panel](https://github.com/yuchen0x1/Xray-server-management-panel)
- 原项目：[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)，作者 MHSanaei
- 更早的 X-UI 项目：[FranzKafkaYu/x-ui](https://github.com/FranzKafkaYu/x-ui)，作者 Franz Kafka Yu

重新发布或继续修改时，请保留上述归属说明。本仓库的功能和修复可能与
原项目存在差异。

## 社区交流

- QQ 群：`1016190748`
- Telegram：[t.me/lenmoC](https://t.me/lenmoC)

反馈问题时请附上具体版本、操作系统、相关日志，并说明问题是否只出现于
本二改分支。

## 功能特性

- 支持 VLESS、VMess、Trojan、Shadowsocks、WireGuard、Hysteria2、HTTP、
  SOCKS、Dokodemo-door、Tunnel 和 TUN 入站
- 支持 TCP Raw、mKCP、WebSocket、gRPC、HTTPUpgrade、XHTTP 等传输方式
- 支持 TLS、XTLS、REALITY、Fallback、路由规则、Balancer 和出站工具
- 支持客户端流量配额、到期时间、IP 限制、在线状态、分享链接、二维码和订阅
- 内置 Raw、JSON、Clash 及自定义模板订阅输出
- 支持多节点管理、节点健康状态、证书和节点升级工具
- 支持 Telegram 机器人、邮件通知、Fail2ban 和系统指标
- 提供面板内置 OpenAPI 文档的 REST API
- 默认使用 SQLite，也支持 PostgreSQL
- 支持多语言界面、浅色、深色和极暗主题

## 快速开始

### 安装脚本

```bash
bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/Xray-server-management-panel/main/install.sh)
```

安装完成后运行 `x-ui`，可以管理服务、登录凭据、证书、数据库和 Xray 设置。

### Docker

```bash
git clone https://github.com/yuchen0x1/Xray-server-management-panel.git
cd Xray-server-management-panel
docker compose up -d
```

默认 compose 配置使用 SQLite。如需 PostgreSQL，请参考
`docker-compose.yml` 中的数据库环境变量。

本分支当前提供源码和 Docker 构建方式，尚未在当前仓库声明自己的发布产物。
仓库内的安装脚本仍兼容原项目的发布渠道；直接执行它安装的是原项目发布版本，
不是当前二改分支的未发布二进制。

```bash
bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/Xray-server-management-panel/main/install.sh)
```

### Windows

仓库包含 Windows 支持，请阅读
[`windows_files/readme.txt`](windows_files/readme.txt) 中的平台说明。

## 开发与验证

环境要求：

- Go 1.26
- Node.js 24 或更高版本
- npm 10 或更高版本
- 支持 SQLite 编译的 C 编译器

前端开发：

```bash
cd frontend
npm install
npm run dev
```

从仓库根目录运行完整检查：

```bash
make gen
make lint
make typecheck
make test
make build
make verify
```

`make verify` 是本地 CI 风格的完整检查。前端构建会生成
`internal/web/dist/`，该目录随后会被嵌入 Go 二进制文件。

## 目录说明

- `internal/`：Go 后端、数据库、Xray 生命周期、任务、API 和业务服务
- `frontend/`：React、TypeScript、Ant Design、Vite、Schema 和测试
- `tools/openapigen/`：前端类型、Schema 和 API 文档生成工具
- `deploy/`：cloud-init、云市场和部署冒烟测试
- `docs/`：架构和订阅模板文档
- `media/`：截图及项目媒体资源

## 安全提示

请勿在没有认证和传输保护的情况下公开面板。请使用强密码，尽可能限制
可信网络访问，定期轮换 API Token，并在应用前检查生成的 Xray 配置。
提交问题时请删除密码、Token、私钥和订阅地址。

## 许可证

本项目使用 GNU General Public License v3.0，完整文本请参阅
[`LICENSE`](LICENSE)。
