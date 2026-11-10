# 待办与下一轮

## 当前状态

轮次 1 已完成（订阅限流、webhook 五项缺陷）。轮次 2 待启动。

## 下一轮优先事项（按 P 级）

### P0/P1

- [ ] **TUIC `maxUserUdpRelays` 未经测试验证**：轮次 1 只加了计数，未见回归测试覆盖"超限后被拒绝"。需补测试并确认 `onClose` 在关联过期路径上也会触发（`expireLocked` 与 `closeAll` 直接关 relay，未走 `SocksUDPSession.Close`？需核实计数是否会泄漏）。
- [ ] **面板 API token 无 scope/过期**（上游 `1230559e`）：本仓库 token 权限等同管理员。
- [ ] **节点 token 明文落库**（上游 `1793a9b8`）：`internal/crypto/nodetoken/` 完全缺失。

### P1

- [ ] `-getApiToken` 累积 admin token（上游 `34c248bb`）。
- [ ] Happ 本地 SOCKS/HTTP 代理未强制鉴权（上游 `6ab718f8`）。
- [ ] `internal/sub` 测试顺序污染：`TestGenHostRemark_PerClient` 全包运行失败、单独运行通过（子代理报告）。

### P2（UI/UX）

- [ ] 审查 `frontend/src/pages/index`、`inbounds`、`clients` 三页的空状态、加载态、错误态。
- [ ] 检查表格在大数据量下的分页/虚拟滚动表现。
- [ ] 移动端宽度适配（Ant Design 6 响应式断点）。

### P3（功能）

- [ ] 订阅页导出/备份入口的可发现性。
- [ ] 全局命令面板（Ctrl+K）覆盖面扩展。

### P4（性能）

- [ ] `vendor` chunk 约 1.35MB，评估手动分包收益。
- [ ] Storybook iframe bundle 1.33MB。

## 已知限制

- `make` 未安装，`make verify` 需拆步执行。
- GitHub API 限流，`gh release view` 可用。
- `proxy.golang.org` 不通，用 `goproxy.cn`。
- Windows 需 `PATH=.../llvm-mingw/bin` + `CGO_ENABLED=1` + `CC=x86_64-w64-mingw32-gcc`。
