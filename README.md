<div align="center">

# 3X-UI · Xray 服务器管理面板

> 用于管理 Xray-core 服务器的 Web 控制面板：协议配置、客户端与流量管理、订阅服务、节点协同、路由工具和运行状态监控

[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-61DAFB?logo=react&logoColor=white)](https://react.dev/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)

</div>

> 本仓库是基于原项目进行的**二次修改维护分支**，仅建议用于个人和合法用途。

## 项目归属与致谢

- 当前仓库：[yuchen0x1/3x-ui](https://github.com/yuchen0x1/3x-ui)
- 原项目：[MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui)（作者：MHSanaei）
- 更早的 X-UI 项目：[FranzKafkaYu/x-ui](https://github.com/FranzKafkaYu/x-ui)（作者：Franz Kafka Yu）

重新发布或继续修改时，请保留上述归属说明。本仓库的功能和修复可能与原项目存在差异。

## 📖 项目简介

3X-UI 是一个用于管理 Xray-core 服务器的 Web 控制面板，提供协议配置、客户端与流量管理、订阅服务、节点协同、路由工具和运行状态监控。

## 🎯 功能特性

- 支持 VLESS、VMess、Trojan、Shadowsocks、WireGuard、Hysteria2、HTTP、SOCKS、Dokodemo-door、Tunnel 和 TUN 入站
- 支持 TCP Raw、mKCP、WebSocket、gRPC、HTTPUpgrade、XHTTP 等传输方式
- 支持 TLS、XTLS、REALITY、Fallback、路由规则、Balancer 和出站工具
- 支持客户端流量配额、到期时间、IP 限制、在线状态、分享链接、二维码和订阅
- 内置 Raw、JSON、Clash 及自定义模板订阅输出
- 支持多节点管理、节点健康状态、证书和节点升级工具
- 支持 Telegram 机器人、邮件通知、Fail2ban 和系统指标
- 提供面板内置 OpenAPI 文档的 REST API
- 默认使用 SQLite，也支持 PostgreSQL
- 支持多语言界面、浅色、深色和极暗主题

## 🧱 技术栈

| 模块 | 技术 |
| --- | --- |
| 后端 | Go |
| 前端 | React |
| 数据库 | SQLite（默认）/ PostgreSQL（可选） |
| 部署 | 安装脚本 / Docker |

## 🚀 快速开始

### 方式一：安装脚本

```bash
bash <(curl -Ls https://raw.githubusercontent.com/yuchen0x1/3x-ui/main/install.sh)
```

安装完成后运行 `x-ui`，可以管理服务、登录凭据、证书、数据库和 Xray 设置。

### 方式二：Docker

```bash
docker compose up -d
```

默认发布端口：`2053`（可在 `docker-compose.yml` 通过 `XUI_PORT` 调整）。数据库默认保存在 `db/`（挂载到容器内 `/etc/x-ui/`）。

## ⚙️ 默认配置与注意事项

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| 默认账号 | `admin` / `admin` | **初始化项目，首次登录请立即修改密码** |
| 面板端口 | `2053` | Docker 默认；脚本安装以安装输出为准 |
| 数据库 | SQLite `/etc/x-ui` | 忘记密码时可用 sqlitebrowser 打开数据库查询 |
| 环境变量 | `XUI_*` | Docker 通过环境变量配置，见 `.env.example` |

> 只建议用于个人和合法用途，请遵守所在地法律法规和服务条款。

## 📬 联系方式

- QQ 交流群：`1016190748`
- Telegram：[https://t.me/lenmoC](https://t.me/lenmoC)

联系或反馈时，请备注原项目地址 `https://github.com/MHSanaei/3x-ui`、原创作者 `MHSanaei`，并说明你使用的版本、运行平台和具体问题。