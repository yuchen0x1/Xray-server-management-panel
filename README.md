<div align="center">

# 3X-UI · Xray 服务器管理面板

> 用于在浏览器里管理 Xray-core 服务器：入站、出站、客户端、订阅、节点协同、路由、通知和运行状态。

[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-61DAFB?logo=react&logoColor=white)](https://react.dev/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)

</div>

> 本仓库是维护分支，适合个人和合法用途。

## 项目解决什么问题

3X-UI 解决的是 Xray-core 服务器“靠命令行和配置文件不好管”的问题。它把常见运维动作收进一个 Web 面板里，方便集中管理多协议入站、客户端生命周期、订阅链接、节点协同和运行状态。

## 主要功能

- 管理 VLESS、VMess、Trojan、Shadowsocks、WireGuard、Hysteria2、HTTP、SOCKS、Tunnel、TUN 等入站
- 管理 TCP、WebSocket、gRPC、HTTPUpgrade、XHTTP、mKCP 等传输
- 配置 TLS、REALITY、路由规则、Balancer 和常用出站
- 管理客户端流量配额、到期时间、IP 限制、在线状态和分享链接
- 生成 Raw、JSON、Clash 和自定义模板订阅
- 管理多节点、证书、面板更新和 Xray 重启
- 支持 Telegram 机器人、邮件通知、Fail2ban、API Token 和系统指标
- 默认使用 SQLite，也支持 PostgreSQL
- 提供多语言界面和深色/浅色主题

## 安装方法

### 方式一：官方脚本

```bash
bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/xray-panel/main/install.sh)
```

脚本安装会生成随机的用户名、密码、端口和 Web base path，并把结果写到 `/etc/x-ui/install-result.env`。

### 方式二：Docker

```bash
docker compose up -d
```

默认是 SQLite。若要使用仓库自带的 PostgreSQL，再执行：

```bash
docker compose --profile postgres up -d
```

Docker 数据默认挂载在 `db/`，证书挂载在 `cert/`。

### 方式三：无人值守 / cloud-init

```bash
XUI_NONINTERACTIVE=1 bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/xray-panel/main/install.sh)
```

适合自动化部署。生成的凭据会写入 `/etc/x-ui/install-result.env`，同时会打印访问地址。

## 使用方法

1. 打开 `http://<server-ip>:<port>/<web-base-path>`
2. 用安装时生成的用户名和密码登录
3. 在面板里添加入站、客户端、路由、订阅和节点
4. 需要时用 `x-ui` 命令查看设置、修改端口、重置账号、迁移数据库或管理服务
5. 如果要调用 API，使用 `Authorization: Bearer <token>` 访问 `/panel/api/*`

## 输入输出示例

### 示例 1：脚本安装

输入：

```bash
bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/xray-panel/main/install.sh)
```

输出：

```text
/etc/x-ui/install-result.env (mode 600)
XUI_USERNAME=...
XUI_PASSWORD=...
XUI_PANEL_PORT=...
XUI_WEB_BASE_PATH=...
XUI_ACCESS_URL=...
XUI_API_TOKEN=...
```

### 示例 2：Docker 启动

输入：

```bash
docker compose up -d
```

输出：

```text
面板在 2053 端口提供服务
SQLite 数据持久化到 ./db/
```

## 归属与致谢

- 当前仓库：[yuchen0x1/xray-panel](https://github.com/yuchen0x1/xray-panel)
- 原项目：[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)
- 更早的 X-UI 项目：[FranzKafkaYu/x-ui](https://github.com/FranzKafkaYu/x-ui)

如需反馈或继续跟进，请说明你使用的版本、运行平台和具体问题。
