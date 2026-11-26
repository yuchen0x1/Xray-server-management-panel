# 版本记录

本页记录此维护分支实际发布的版本，依据对应标签、提交差异和 GitHub Release 整理。日期使用北京时间（Asia/Shanghai）；技术细节见 [CHANGELOG](../../CHANGELOG.md)。

| 版本                                                                                      | 发布日期   | 主要变化                                         |
| ----------------------------------------------------------------------------------------- | ---------- | ------------------------------------------------ |
| [v3.13.2](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.13.2) | 2026-11-27 | 安全与运行时修复、完整发行包及构建缓存治理 |
| [v3.13.1](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.13.1) | 2026-11-20 | 发布验证未完成，无发行附件，请使用 v3.13.2 |
| [v3.13.0](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.13.0) | 2026-11-16 | 观测、订阅、备份与入站运维体验增强                 |
| [v3.12.1](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.12.1) | 2026-11-13 | 一致性备份、恢复确认、更新校验和 Docker 架构修复 |
| [v3.12.0](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.12.0) | 2026-11-10 | API 令牌权限与有效期、TUIC 和通知安全加固        |
| [v3.8.0](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.8.0)   | 2026-11-08 | TUIC v5、全局命令面板、入站主机备注              |
| [v3.7.0](https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.7.0)   | 2026-11-07 | 多客户端订阅导入、Webhook 通知、订阅防爆破       |

`v3.9.0` 至 `v3.11.0` 等标签属于上游项目，本维护分支没有以这些编号发布版本。它们不应混入本页的发布记录。

## 未发布

- 暂无新的未发布变更。

## v3.13.2 — 2026-11-27

### 📢 版本概述

本次更新修复了反向代理来源地址识别、MTProto IPv6 监听与流量统计、Geo 数据下载和页面缓存刷新等问题，并修正发布包验证流程。发行附件上传及签名验证成功后才会将此版本设为最新稳定版，替代没有发行附件的 v3.13.1。

### 🔧 改进与优化

- 网站与面板使用同一份生成的 OpenAPI 规范，并同步四种语言 API 页的操作引用，避免网站遗漏已支持的接口或引用旧参数名。
- 开发版归档按实际提交版本校验；正式版继续严格校验稳定版本号，Linux 与 Windows 均保留归档启动检查。
- Actions 构建产物保留 7 天，自动清理超过保留期且工作流已结束的产物；历史 Release 附件不受清理影响。

### 🐛 问题修复

- 连续收到不同资源的 WebSocket 通知时，各类页面数据均能刷新；设置保存成功后同步刷新客户端与入站页面共享的默认配置。
- 修复 MTProto 裸 IPv6 地址无法监听的问题，同步子进程状态访问，补计首次采样及计数器重置后的流量，并丢弃进程替换前的过期统计。
- Geo 数据下载中断或内容为空时保留原文件；只有实际替换数据后才重启 Xray。
- 修改账户凭据后，API 返回的会话使用正确的登录代次；页面原有重新登录流程保持不变。
- 非法 VMess 外链包含 JSON null 时不再导致订阅请求 panic。
- 整合 PR #26 的 TUIC 测试流读取修复，避免测试提前取消响应流影响后续回显验证。

### 🔒 安全与兼容性

- 将 Gorilla WebSocket 锁定到上游修复提交，解决 GO-2026-6278 / GHSA-w67g-5rqw-f597 的客户端掩码随机数问题。
- 对已配置的可信反向代理，从可信端解析转发链，防止客户端伪造前缀影响登录限流和审计来源地址。
- 账户凭据与双因素认证重置在同一事务提交，避免改密失败却提前关闭双因素认证；成功改密后的既有重置策略不变。
- 复核 CodeQL 的 UDP 缓冲区、证书路径及订阅 ETag 告警：分别已有严格大小上限、属于管理员证书路径功能及非密码内容摘要；记录误报理由，保留既有功能。
- 保留 SHA256SUMS 与 Sigstore 发布者签名。Linux 在线更新需先通过可信渠道安装 Cosign，并确保面板服务可以找到它。

### 📦 下载与安装

| 文件名 | 平台 / 架构 | 说明 |
| --- | --- | --- |
| x-ui-linux-amd64.tar.gz | Linux x64 | 服务安装与升级归档 |
| x-ui-linux-arm64.tar.gz | Linux arm64 | 服务安装与升级归档 |
| x-ui-linux-armv7.tar.gz | Linux ARMv7 | 服务安装与升级归档 |
| x-ui-linux-armv6.tar.gz | Linux ARMv6 | 服务安装与升级归档 |
| x-ui-linux-armv5.tar.gz | Linux ARMv5 | 服务安装与升级归档 |
| x-ui-linux-386.tar.gz | Linux x86 | 服务安装与升级归档 |
| x-ui-linux-s390x.tar.gz | Linux s390x | 服务安装与升级归档 |
| x-ui-windows-amd64.zip | Windows x64 | 便携归档，手动升级 |
| SHA256SUMS | 通用 | 发行归档的 SHA-256 校验清单 |
| SHA256SUMS.sigstore.json | 通用 | 校验清单的 Sigstore 签名及验证材料 |

- 首次安装见 [项目说明](https://github.com/yuchen0x1/Xray-server-management-panel#readme)。
- 已安装的 Linux 用户可在满足 Cosign 要求后使用面板更新功能或安装脚本升级；Windows 用户验证归档后停止面板，再替换程序文件。

### 🔄 升级与数据迁移

- 本版本没有新增数据库迁移。升级前备份数据库、配置及当前程序；保留现有数据目录和环境配置。
- 已有版本仍通过本仓库的 latest 发布入口发现新版本，此入口在完整附件就绪后切换。
- Docker 用户拉取对应版本镜像并重建容器；回退时使用升级前的程序与相匹配的数据备份。
- v3.13.1 的公开发布没有发行附件，请直接使用本版本。

### 🧪 验证结果

- 本地 `make verify` 完整通过，涵盖生成文件一致性、Go 与前端代码检查、类型检查、完整测试、生产构建及 Storybook 构建。
- 修复后 `govulncheck` 未发现漏洞；前端生产及开发依赖 `npm audit` 均为 0 漏洞；暂存变更凭据扫描通过。
- 前端完整测试通过：89 个文件、983 项测试；类型检查、ESLint 及全部 13 种语言的翻译键检查通过。
- MTProto 竞态检测与订阅、TUIC、Geo 下载、代理来源地址等针对性回归测试通过。
- GitHub CI 全仓 Go 测试在 `-race -shuffle=on -count=1` 下通过。
- 稳定版发布由完整 CI、Linux 与 Windows 归档启动检查、校验清单及签名验证共同把关。
- Linux 非 amd64 架构仅完成交叉构建；未进行真实跨网络吞吐或长期运行测试。

### ⚠️ 已知问题

- Windows 不支持面板内自动更新，需要手动替换程序。新版本提示仍可用。
- Linux 在线更新依赖可用的 Cosign；缺少时会拒绝更新，请先按 [更新指南](https://github.com/yuchen0x1/Xray-server-management-panel/blob/main/docs/AUTO_UPDATE.md) 完成配置。

---

**项目地址：** https://github.com/yuchen0x1/Xray-server-management-panel

**问题反馈：** https://github.com/yuchen0x1/Xray-server-management-panel/issues

**版本下载：** https://github.com/yuchen0x1/Xray-server-management-panel/releases/tag/v3.13.2

## v3.13.1 — 2026-11-18

### 版本概述

Linux 面板更新在停止服务前验证归档的发布者签名、架构和版本。正式发布需通过 CI 与 Linux、Windows 归档启动检查。Windows 可验证发行归档后手动更换程序文件。

### 升级说明

- Linux 在线更新需要通过可信渠道安装 Cosign，并让面板服务的 PATH 能找到它。
- 从 v3.13.0 升级时，先备份数据库和配置。升级前可独立验证 Release 的签名和校验清单。
- Docker 仍通过拉取指定版本的镜像并重建容器升级。安装要求和回退范围见 [更新指南](../AUTO_UPDATE.md)。

## v3.13.0 — 2026-11-16

### 入站与任务稳定性

- 入站列表的复制分享链接动作恢复正常，并继续复用现有链接生成和剪贴板流程。
- 入站搜索词和节点筛选会在刷新后恢复；节点被删除后，失效筛选会自动清理。
- LDAP 同步配置和数据读取错误会记录具体字段并结束当前任务轮次，不再直接触发调度器 panic。
- 页面级错误边界保留侧栏，并提供重试和返回首页入口。

### 订阅、备份与监控

- raw、JSON 和 Clash 订阅支持 ETag 与 `If-None-Match`；内容未变化时返回 `304 Not Modified`。
- 可选本地数据库备份保留按数量或天数清理快照，默认关闭；启用后使用原子写入、同步落盘和 0600 权限。
- 新增受认证保护的低基数 Prometheus 指标和告警规则下载，指标不包含客户端、入站或凭据标识。

### 发布与维护

- CI 增加 13 个语言包的 key、额外 key 和占位符一致性检查。
- Windows 后续归档不再附带固定版本的第三方 OpenSSL 安装器；证书管理可按文档单独安装 OpenSSL，已经发布的历史附件保持原样。
- 依赖更新保持在 `main` 上验证，关闭自动更新 PR 创建，同时保留漏洞告警和 CI 安全检查。

### 升级说明

- 本版没有新增数据库字段或迁移；备份保留策略只有显式设置 `XUI_BACKUP_RETENTION_COUNT` 或 `XUI_BACKUP_RETENTION_DAYS` 时才启用。
- Prometheus 端点使用面板会话或 Bearer token 认证，建议仅向受信任的监控系统开放。
- Windows 更新不再从归档中提供固定版本 OpenSSL 安装器，证书功能使用前请按文档准备 OpenSSL。

[完整变更记录](https://github.com/yuchen0x1/Xray-server-management-panel/compare/v3.12.1...v3.13.0)

## v3.12.1 — 2026-11-13

### 备份与恢复

- SQLite 数据库下载采用一致性快照，包含已经提交到 WAL 的数据，避免复制正在写入的数据库文件。
- SQL 迁移转储保留二进制字段类型、文本中的零字节以及含引号的表名，减少恢复后的数据差异。
- 恢复前显示所选文件名和覆盖提醒；可以取消，空文件会被拒绝，恢复期间不接受重复操作。
- 同一面板进程会拒绝同时进行的第二次恢复，异常结束后界面恢复可操作状态。
- 新提示覆盖全部 13 种语言，窄屏下长文件名也能正常换行。

### 更新与使用体验

- 升级 Go 1.27.4，修复 `net/http` 中可能导致 HTTP 服务崩溃的漏洞（GO-2026-6859）。
- Linux 在线更新在安装前校验归档和 SHA-256 清单，可使用配置的主站或备用镜像。
- 更新后服务启动失败时尝试恢复旧版本；数据库目录保持独立。
- 侧栏项目、社区和帮助链接可以直接打开。
- Release 提供 Linux、Windows 归档及 `SHA256SUMS`；正式版本在附件上传完毕后公开。
- Docker 镜像从对应架构的正式安装包构建并校验 SHA-256，修正 32 位 x86、ARMv6 和 ARM 核心文件匹配。

### 升级说明

- 本版不改变数据库结构或 API 路径。升级或恢复前请先另存备份。
- Linux 服务安装可使用面板在线更新；Windows 请下载对应归档进行更新。
- SHA-256 校验用于验证文件完整性，不等同于发布者签名。Windows 产物尚未配置 Authenticode 签名。
- 恢复互斥只覆盖单个面板进程；共享 PostgreSQL 的多个实例仍需协调停写。

### 发布与验证

- 提供 Linux 七种架构和 Windows amd64 安装包；Docker 提供 amd64、arm64、arm/v7、arm/v6、386 镜像。
- CI、Linux amd64/arm64 干净安装、Windows 下载校验与版本检查通过；Docker 五架构均验证面板、Xray 和 MTProto 程序可启动。
- 干净安装验证不代表已经完成生产环境跨版本更新与故障回滚演练。
- 应用标签对应 `76f49f14`；Docker 打包修复在后续 `main` 提交 `05d59020` 落地，镜像仍使用此版本的已校验归档。
- [应用提交差异](https://github.com/yuchen0x1/Xray-server-management-panel/compare/v3.12.0...v3.12.1) · [Docker 打包修复](https://github.com/yuchen0x1/Xray-server-management-panel/commit/05d590204bf94b7ee5952b19110514d8f02287e4)

## v3.12.0 — 2026-11-10

### API 令牌与账户安全

- API 令牌可指定管理员、只读监控、节点同步、节点管理四种权限范围，并设置到期时间；界面显示权限与有效期。
- 已有令牌迁移为管理员权限，未设置有效期时继续长期有效；异常权限值不再按管理员放行。
- 删除、启用和停用令牌时可校验预期权限，减少旧页面误操作；数据库按令牌哈希查询，不保存令牌明文。
- 命令行获取 API 令牌改为轮换同一枚 CLI 令牌，避免反复生成无法管理的长期令牌。
- 重新绑定双重认证需要验证当前验证码，防止保存设置时直接替换 TOTP 密钥。

### 连接、订阅与通知

- TUIC 增加 QUIC Retry 和连接准入限制，约束全局与每来源的待认证连接、认证前数据流、账户并发会话及 UDP 中继。
- 修复 TUIC 会话尚未建立完整连接就关闭时的空指针错误，失败或关闭后释放中继配额。
- 订阅限流表满时回收过期和最旧记录，IPv6 来源按 /64 网段统计，减少轮换地址绕过限制的机会。
- Webhook 拒绝跨主机重定向，防止通知密钥随请求泄漏；CPU、内存告警遵循配置阈值，限制同时发送的请求数量。
- 随包 Xray 固定为 `v26.10.21`，与面板使用的核心库版本对齐。

### 升级说明

- API 令牌字段会自动迁移，原有令牌权限保持兼容；新增令牌可按实际用途缩小权限范围。
- TUIC 的连接与会话限制会影响超过配额的连接；Webhook 接收地址应直接指向最终主机。
- [完整提交差异](https://github.com/yuchen0x1/Xray-server-management-panel/compare/v3.8.0...v3.12.0)

## v3.8.0 — 2026-11-08

### 新功能与体验

- 新增 TUIC v5 入站协议，支持对应客户端连接。
- 入站列表直接显示关联主机备注，便于识别多节点配置。
- 新增 `Ctrl+K` 全局命令面板，快速跳转入站、设置及常用操作。
- 新安装默认使用浅色主题，可在外观设置切换深色主题。

### 修复与兼容

- 客户端流量写入失败时返回错误，避免统计保存失败被静默忽略。
- 兼容 Xray xdns finalmask 的旧字符串配置格式，减少升级后的配置解析错误。
- 对齐 quic-go 依赖并修正 TUIC 测试中的流关闭时序假设。
- 更新 Go 至 1.27.3 并处理当时发现的 Go、前端及文档依赖安全告警；这些结果不代表该旧版本长期无漏洞。
- 修复版本标签推送可能被工作流路径过滤跳过的问题，使标签能正常触发构建。

### 升级说明

- 使用 TUIC 时需选择支持 TUIC v5 的客户端，并放通入站使用的 UDP 端口。
- 原有 xdns 字符串配置保留兼容；主题可以在面板中调整。
- [完整提交差异](https://github.com/yuchen0x1/Xray-server-management-panel/compare/v3.7.0...v3.8.0)

## v3.7.0 — 2026-11-07

### 新功能

- 订阅页新增 Clash、Sing-box、Hiddify、Surge、Loon 的客户端唤醒链接和桌面导入入口，按订阅格式选择导入源。
- 新增 Webhook 通知，可适配钉钉、企业微信、飞书和 Discord；通过 `XUI_WEBHOOK_URL`、`XUI_WEBHOOK_SECRET` 配置。
- 订阅端针对 subId 猜测和路径探测进行按 IP 滑动窗口限流，超限后封禁 15 分钟。
- 订阅服务接入可信代理设置，避免直接信任任意来源的 `X-Forwarded-For`。

### 修复与升级说明

- 安装和更新时遇到 Gitee/GitHub API 403、429 限流，立即切换备用通道，减少无效重试。
- Webhook 为按需配置功能；反向代理部署应核对可信代理设置，以便正确识别来源 IP。
- 这是本维护仓库当前保留的首个正式 Release，继承功能以对应标签源码为准；不将全部上游历史功能计作本版新增。
- [对应源码](https://github.com/yuchen0x1/Xray-server-management-panel/tree/v3.7.0)

## dev-latest — 滚动预发布

- 跟随 `main` 更新，用于提前验证尚未进入下一稳定版的改动，不是新的正式版本号。
- GitHub 发布说明首行 `commit=<完整提交哈希>` 表示该批附件实际构建的提交；同名附件会被后续构建替换。
- 开发版会包含最近稳定版的功能和之后的维护变更；部署前另存数据库备份，需要固定版本时选择正式 Release。
- 保留此渠道是为了兼容面板的开发版更新入口；它是发布标签，不是开发分支。

[查看全部发布附件](https://github.com/yuchen0x1/Xray-server-management-panel/releases)
