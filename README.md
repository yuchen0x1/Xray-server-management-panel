# 3X-UI

3X-UI is a web control panel for managing Xray-core servers. It provides
protocol configuration, client and traffic management, subscriptions, node
coordination, routing tools, and operational monitoring from one interface.

This repository is a maintained **derivative fork / secondary modification**
of the original project. It is intended for personal and lawful use only.

## Attribution

- Current repository: [yuchen0x1/3x-ui](https://github.com/yuchen0x1/3x-ui)
- Original project: [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui), by MHSanaei
- Earlier X-UI project: [FranzKafkaYu/x-ui](https://github.com/FranzKafkaYu/x-ui), by Franz Kafka Yu

Please keep this attribution when redistributing modified builds. Features and
fixes in this repository may differ from the original project.

## Community

- QQ group: `1016190748`
- Telegram: [t.me/lenmoC](https://t.me/lenmoC)

When asking for help, include the exact version, operating system, relevant
logs, and whether the issue is from the original project or this fork.

## Features

- VLESS, VMess, Trojan, Shadowsocks, WireGuard, Hysteria2, HTTP, SOCKS,
  Dokodemo-door, Tunnel, and TUN inbounds
- TCP Raw, mKCP, WebSocket, gRPC, HTTPUpgrade, and XHTTP transports
- TLS, XTLS, REALITY, fallbacks, routing rules, balancers, and outbound tools
- Per-client traffic quotas, expiry dates, IP limits, online status, share links,
  QR codes, and subscription links
- Built-in subscription output for raw, JSON, Clash, and custom templates
- Multi-node management with node health, certificate, and update tooling
- Telegram bot, email notifications, Fail2ban integration, and system metrics
- REST API with in-panel OpenAPI documentation
- SQLite by default, with optional PostgreSQL support
- Multiple UI languages plus light, dark, and ultra-dark themes

## Quick Start

### Source or Docker

This fork currently ships source and Docker build instructions rather than
claiming upstream release artifacts. Use the Docker command below for a
reproducible local build, or build the binary with `make build`.

### Upstream-compatible installer

The bundled installer remains compatible with the original project's published
release artifacts:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/MHSanaei/3x-ui/main/install.sh)
```

That command installs the original project's release channel, not an
unpublished binary from this fork. Run `x-ui` after installation to manage the
service, credentials, certificates, database, and Xray settings.

### Docker

```bash
git clone https://github.com/yuchen0x1/3x-ui.git
cd 3x-ui
docker compose up -d
```

The default compose setup uses SQLite. PostgreSQL can be enabled through the
database environment variables documented in `docker-compose.yml`.

### Windows

Windows support is included in the repository. See
[`windows_files/readme.txt`](windows_files/readme.txt) for the platform-specific
startup notes.

## Configuration

SQLite is the default database and is stored under the configured X-UI data
directory. PostgreSQL can be selected with:

```text
XUI_DB_TYPE=postgres
XUI_DB_DSN=postgres://xui:password@127.0.0.1:5432/xui?sslmode=disable
```

Other commonly used settings include:

| Variable | Purpose |
| --- | --- |
| `XUI_INIT_WEB_BASE_PATH` | Initial web panel path |
| `XUI_PORT` | Panel listen port |
| `XUI_LOG_LEVEL` | Log verbosity |
| `XUI_DEBUG` | Enable debug behavior |
| `XUI_ENABLE_FAIL2BAN` | Enable Fail2ban integration |
| `XUI_DB_FOLDER` | SQLite data directory |
| `XUI_DB_MAX_OPEN_CONNS` | PostgreSQL open connection limit |
| `XUI_DB_MAX_IDLE_CONNS` | PostgreSQL idle connection limit |

See `.env.example`, `docker-compose.yml`, and `docs/` for the complete list.

## Development

### Requirements

- Go 1.26
- Node.js 24 or newer
- npm 10 or newer
- A C compiler for SQLite-enabled Go builds

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The Vite server runs on `http://localhost:5173/` and proxies panel API
requests to a Go server on port `2053`.

### Verification

From the repository root:

```bash
make gen
make lint
make typecheck
make test
make build
make verify
```

`make verify` is the local CI-style gate. The frontend build regenerates
`internal/web/dist/`, which is embedded into the Go binary.

## Repository Layout

- `internal/`: Go backend, database, Xray lifecycle, jobs, API, and services
- `frontend/`: React, TypeScript, Ant Design, Vite, schemas, and tests
- `tools/openapigen/`: generated frontend schema and API tooling
- `deploy/`: cloud-init, marketplace, and smoke-test deployment material
- `docs/`: architecture and subscription customization documentation
- `media/`: screenshots and project media

## Security

Do not expose the panel without authentication and transport protection. Use
strong credentials, restrict access to trusted networks where possible, rotate
API tokens, and review generated Xray configurations before applying them.
Security reports should not contain passwords, API tokens, private keys, or
subscription URLs.

## License

This project is distributed under the GNU General Public License v3.0. See
[`LICENSE`](LICENSE) for the full license text.
