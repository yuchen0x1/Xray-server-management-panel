# 项目分析

生成时间：2026-11-10 12:25 (+0800)

## 基本识别

| 项 | 值 |
|---|---|
| 项目 | 3x-ui 维护分支（Xray 服务端管理面板） |
| 用途 | 管理 Xray-core 代理服务、客户端配额、订阅分发、多节点 |
| 后端 | Go 1.27.3 + Gin + GORM |
| 前端 | React 19 + Ant Design 6 + Vite 8 + TypeScript 6.0.3 |
| 数据库 | SQLite（默认，CGo），PostgreSQL 可选 |
| 构建 | Vite → `internal/web/dist/` → `embed.FS` 嵌入单二进制 |
| 测试 | Go stdlib testing（297 个测试文件）+ Vitest（84 文件 / 971 用例） |
| 前端版本 | package.json 0.4.3 / panel 版本 3.8.0 |

## 模块地图

- `main.go` — 入口 + CLI（run / migrate / setting / cert）
- `internal/database/` — GORM schema、AutoMigrate、手写迁移（`db.go` 66 个函数）
- `internal/xray/` — Xray 子进程生命周期、配置生成、gRPC stats
- `internal/mtproto/` — `mtg-multi` 子进程（MTProto 多 secret）
- `internal/tuic/` — 进程内 QUIC TUIC v5 服务端（21 文件）
- `internal/sub/` — 订阅服务（raw / JSON / Clash）
- `internal/web/controller|service|job|middleware|runtime` — HTTP 与业务逻辑
- `internal/web/service/webhook/` — 多渠道告警（fork 独有）
- `frontend/src/pages/` — 页面：index / inbounds / clients / nodes / hosts / sub / settings / xray / groups / api-docs
- `tools/openapigen/` — Go 类型 → Zod/OpenAPI 生成器

## 分支状态

- 分支：`autonomous-upgrade`（从 `main` @ `654ee9a2` 创建）
- 检查点：`49925bfb`
- 与上游关系：见 `UPGRADE_PLAN.md`（分叉点 `87ebcc7a`，756 个上游提交未同步，208 个真实冲突文件）

## 已验证的基线质量

| 检查 | 结果 |
|---|---|
| `go test -shuffle=on ./...` | 通过 |
| `golangci-lint run` | 0 issues |
| `govulncheck ./...` | 无漏洞 |
| `npm run typecheck` / `lint` | 通过 |
| `npm test` | 971 通过 |

## 已知结构性问题

1. 历史被重写过（3112 个 patch-id 等价提交），无法 merge/rebase 上游。
2. 本地 tag `v3.7.0`/`v3.8.0` 与上游同名 tag 冲突。
3. 上游 6 项安全加固未移植（API token scope、节点 token 加密等），见 `UPGRADE_PLAN.md`。
