# 升级方案与执行记录

日期：2026-11-10
仓库：`C:\Users\lenmo\Desktop\Xray-server-management-panel`
基线：`main` @ `654ee9a2`（v3.8.0 + 3 提交）

本文件记录对分支与上游关系的实测结论、已完成的阶段 0 修复，以及后续阶段的执行清单。

## 1. 与上游的真实关系

早期审计报告中的“领先 3298 / 落后 316”是在**浅克隆**的 remote-tracking ref 上计算的，不能作为决策依据。补齐历史后（`git fetch --unshallow`）的实测结果：

```
上游真实分叉点（内容层面）: 87ebcc7a  2026-07-29  feat(ui): tag settings that sit at their shipped default value
  · 该提交在 upstream/main 上，且本仓库分叉点父提交的 tree 与之完全一致

分叉点之后:
  上游新增            761 个提交（其中 756 个为本仓库真正缺失）
  本仓库自有提交       64 个（patch-id 判定）
  与上游 patch-id 等价 3112 个
```

`git cherry upstream/main HEAD` 输出 `64 +` / `3112 -`，说明本分支历史被重写过。**因此 `git merge upstream/main` 与 `git rebase upstream/main` 都不可用**：前者会产生数千个无意义冲突，后者会重放 3297 个提交。

内容层面的距离：

| 对比点 | 与 HEAD 的差异 |
|---|---:|
| 上游 v3.6.0 | 323 文件, +9,061 / -19,928 |
| 上游真实 v3.7.0 | 995 文件, +82,375 / -38,193 |
| 上游真实 v3.8.0 | 1,261 文件, +139,004 / -42,132 |
| 上游 v3.9.0 | 1,476 文件, +168,757 / -37,220 |
| 上游 v3.11.0 | 1,769 文件, +225,066 / -45,627 |

### 真实冲突面

以分叉点 `87ebcc7a` 为基准：

```
我方改动文件 263 个
上游改动文件 1818 个
双方都改（真实冲突面）208 个
仅我方改动（纯 fork 资产）55 个
```

208 个冲突文件按目录分布：`frontend/src` 85、`internal/web` 45、`internal/tuic` 21、`internal/sub` 7、`internal/database` 4、`.github/workflows` 3、其余零散。

### tag 命名冲突

本地 `v3.7.0`（5b8f8da3）与 `v3.8.0`（42ad9188）与上游同名 tag 指向不同提交，`git fetch --tags upstream` 会被拒绝（`would clobber existing tag`）。`v3.8.5`、`v3.9.0`、`v3.10.0`、`v3.10.2`、`v3.11.0` 与上游一致。

## 2. 阶段 0（已完成）

### 2.1 Xray 二进制与库版本对齐

**问题**：`go.mod` 的 xray-core 库是 `2026-10-21` 快照，而运行时下发的 Xray 二进制 pin 在 `v26.7.28`（2026-07-28），相差约三个月。库版本由 `83f17e92`（TUIC 移植）提升，但二进制 pin 未同步。

这个错配是实际风险而非理论风险：`0093e9f8` 已为 xray-core `20261021` 的 finalmask/xdns 新结构添加构建期兼容层，而 `v26.7.28` 二进制不认识该结构。

**修复**：三处 pin 统一到 `v26.10.21`，与当前库版本配套（这是上游 v3.11.0 的配对关系，已用 `git show v3.11.0:DockerInit.sh` 核对）。

| 文件 | 变更 |
|---|---|
| `DockerInit.sh:35` | `v26.7.28` → `v26.10.21` |
| `.github/workflows/release.yml:117` | `v26.7.28` → `v26.10.21` |
| `.github/workflows/release.yml:275` | `v26.7.28` → `v26.10.21` |

已确认 `v26.10.21` 发布包含全部所需资产：`Xray-linux-{64,32,arm64-v8a,arm32-v7a,arm32-v6,arm32-v5,s390x}.zip`、`Xray-windows-{64,32,arm64-v8a}.zip`，与 `DockerInit.sh` 的 `ARCH` 映射及 release 矩阵完全对应。

### 2.2 2FA 密钥替换门禁（GHSA-xqqw-jqqv-99h6）

**问题**：`internal/web/controller/setting.go` 的确认门禁只覆盖 `twoFactorEnable: true → false`。当保存时保持 `twoFactorEnable=true` 并携带非空 `twoFactorToken`，`preserveRedactedSecrets` 只在提交值为空时才回填存储值，于是非空值直接写入且没有任何分支要求验证码——可静默重绑认证器，把合法管理员锁在外面。

**修复**：门禁改为“当前开启 且（要关闭 或 提交的 secret 与存储值不同）时验证码必需”。空白 secret 仍表示“不变”，正常保存路径不受影响。

对应上游提交 `c8a3a2d`。回归测试 `TestUpdateSettingRequiresCodeToReplaceTwoFactorToken` 覆盖两条路径：无码重绑必须被拒绝且存储值不变；空 secret 保存必须成功且保留原值。

### 2.3 TUIC 未认证连接加固

**问题**：本仓库 TUIC 为 `83f17e92` 的移植实现，缺少上游 2026-10-31 的三项资源边界（`fix(tuic): bound what unauthenticated connections cost`、`bound one client's connections and UDP relays`、`end a UDP response relay with its association`）。公网监听下这是资源耗尽面。

**修复**（移植上游 `pending_conns.go` + 相关接线）：

| 加固点 | 常量 | 作用 |
|---|---|---|
| 未认证连接总数 | `maxPendingConns = 1024` | 每个 pending 连接持有握手状态与 goroutine |
| 单来源配额 | `maxPendingPerSource = 32`（IPv4 单地址 / IPv6 /64） | 只统计已验证来源 |
| 源地址验证阈值 | `pendingRetryThreshold = 64` | 超过后强制走 QUIC Retry，伪造源不产生状态 |
| 未认证流数/类型 | `maxPreAuthStreams = 16` | 防止单连接开流换取 goroutine |
| 单客户端会话数 | `maxUserSessions = 32` | 跨连接累计 |
| 单客户端 UDP 中继数 | `maxUserUdpRelays = 512` | 每个中继占一个宿主端口 |

实现要点：`Start()` 从 `quic.Listen` 改为显式 `quic.Transport`，挂载 `VerifySourceAddress` 与 `ConnContext`；认证成功时 `releasePendingSlot(conn.Context())` 提前归还槽位；`Close()` 补上 transport 关闭。

新增 `internal/tuic/pending_conns.go` 与 3 个回归测试：`TestPendingConnsRetryAndCeilings`、`TestUnauthenticatedStreamsShareASmallHandlerBudget`（90 个未认证流不得产生超过 `2*16` 个 goroutine）、`TestRegisterConnRefusesPastTheSessionCap`。

### 2.4 验证结果

| 检查 | 结果 |
|---|---|
| `go test -shuffle=on -count=1 ./...`（CGO=1） | 通过 |
| `golangci-lint run` | 0 issues |
| `govulncheck ./...` | No vulnerabilities found |
| `npm run typecheck` | 通过 |
| `npm run lint` | 通过 |
| `npm test` | 84 文件 / 971 用例通过 |

## 3. 阶段 1：按目录收敛分叉面

以 `87ebcc7a..upstream/main` 为清单，208 个冲突文件分批推进，每批独立提交。仓库历史不可 merge，因此**全部按 cherry-pick 或手工移植处理**。

| 批次 | 范围 | 优先级依据 |
|---|---|---|
| A | `internal/web/service/*`、`internal/database/*` | 上游 56 个 `fix(sub)`、42 个 `fix(node)`、10 个 `fix(database)`；含多处数据一致性修复 |
| B | `frontend/src/*`（85 文件） | 建议跟随上游 `routes/` 目录结构与 a11y 修复；上游已迁 oxlint/oxfmt + TS 7 + Node 26 |
| C | `internal/tuic/*`（21 文件） | 建议直接采用上游实现替换本地移植版，避免长期双维护 |
| D | tag 与版本号规范 | 重命名本地 `v3.7.0` / `v3.8.0` 为带 fork 后缀的形式，消除与上游 tag 冲突 |

### 尚未处理的已知缺口

以下为上游已有、本仓库缺失且**本轮未修复**的项目，按风险排序：

| 项目 | 上游提交 | 风险 |
|---|---|---|
| API token scope / 过期 / hash 索引查询 | `1230559e`、`perf(auth): hash index` | 中：本仓库为逐行常量时间比较，无 scope 概念 |
| 节点 token 静态加密 | `1793a9b8`（`internal/crypto/nodetoken/`） | 中：节点 API token 明文落库 |
| `-getApiToken` 累积 admin token | `34c248bb` | 中 |
| Happ 本地 SOCKS/HTTP 代理强制鉴权 | `6ab718f8` | 中 |
| 公网 IP 查询边界与格式校验 | `06cb1742` | 低 |
| Xray 库/二进制升到 `v26.10.31` | `72c9092d` 等 3 个适配提交 | 需连同适配一并移植 |

## 4. 阶段 2（可选）：重建基线

若长期需要跟随上游，最干净的做法是：以 `upstream/main` 为基，把 64 个自有提交中仍然需要的作为独立补丁重新应用，而不是合并历史。预计需保留 15–20 个：webhook、订阅限流、Gitee 更新通道、玻璃主题、Ctrl+K 命令面板、入站主机备注、默认白天主题；TUIC 若采用上游实现则归零。

## 5. 不建议执行的操作

- `git merge upstream/main` / `git rebase upstream/main`：历史重写会让冲突量爆炸。
- 为清 3 个 low 告警执行 `npm audit fix --force`：会把 swagger-ui 降到 4.1.3（主版本降级）；上游同样保留 `deepmerge@4.3.1`，属于上游未解决项而非本仓库回归。
- 引入 amneziawg / PIA / oxlint 全家桶：属于上游独立产品线，与本分支定位无关。

## 6. 环境备注

- Go 1.27.3、Node 24.20、npm 11.19。
- `make` 未安装；`make verify` 的各步骤已分别单独执行。
- Windows 下 SQLite 测试需 CGO。可用编译器：`D:\DevTools\llvm-mingw\llvm-mingw-20261020-ucrt-x86_64\bin\x86_64-w64-mingw32-gcc.exe`。
- `proxy.golang.org` 不通，改用 `GOPROXY=https://goproxy.cn,direct`。
