# 会话状态

- 任务启动：2026-11-10 12:25 (+0800)
- 分支：`autonomous-upgrade`（已 fast-forward 合入 `main`）
- 基线检查点：`49925bfb`
- 已发布版本：`v3.12.0`（tag 已推送，Release 已设为 latest 正式版）

## 当前阶段

轮次 1–3 已完成并发布。后续轮次尚未开始。

## 已完成

### 轮次 1 — 订阅限流 + Webhook

- `internal/sub`：记录表满时不再放行新来源（原实现 fail-open），IPv6 按 /64 归并
- `internal/web/service/webhook`：拒绝跨主机重定向（防 `X-Webhook-Secret` 泄露）、
  指标事件按订阅端阈值过滤、投递并发有界、修正阈值文案
- `internal/web/web.go`：仅配置 webhook 的部署不再收不到 `cpu.high`/`memory.high`

### 轮次 2 — TUIC 资源边界

- 无状态 Retry + 连接准入（全局与按来源 /64 上限）、认证前流数量上限
- 每账号会话数 32、UDP relay 512，配额在失败/关闭时归还
- 修复 `SocksUDPSession.Close()` 空指针崩溃（由新测试发现）

### 轮次 3 — API 令牌作用域与过期

- scope：admin / monitor / node-sync / node-admin，越权返回 403
- expiresAt（Unix 毫秒，0 为永不过期）
- 按 SHA-256 索引单条查询替代全表扫描
- 未知 scope 不再回退为管理员
- 可选 `expectedScope` 防止按行号误删
- 旧库 scope 回填为 admin，已签发令牌不失效
- `-getApiToken` 改为轮换单个 `cli-fallback` 令牌，不再累积
- 前端：scope 选择器、过期时间、每行 scope 标签
- node-sync 白名单由 `runtime/remote.go` 真实调用点反推，并有测试固定对应关系

## 最后一次成功测试

发布前完整执行：

- `go test -shuffle=on -count=1 ./...` 全绿
- `golangci-lint run` → `0 issues`
- `npm run typecheck` / `npm run lint` 通过
- `npm run test` → 84 文件 / 971 测试通过
- `npm run build` 产出 `internal/web/dist/`

## 下一步

按 `TODO_NEXT.md` 继续：节点 token 明文落库、`internal/sub` 测试顺序污染、
UI/UX 审查（空态/加载态/大表分页/移动端）、`vendor` chunk 分包。