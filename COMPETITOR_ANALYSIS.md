# 全球开源竞品与相邻项目对标

研究日期：2026-11-16（Asia/Shanghai）

## 研究方法

本轮研究覆盖 GitHub 上的直接竞品、协议核心、客户端、网络管理面板、可观测性项目和发布/更新工具。每个项目至少检查了官方仓库、README、目录结构、许可证、发布或提交记录；重点项目进一步检查了核心源码、工作流、Issues/PR 主题和安全/发布文档。没有复制任何项目代码。

许可证只用于判断能否借鉴设计或重新实现，不把 GPL、AGPL、LGPL、Apache、MIT 或混合许可证代码直接带入本仓库。当前仓库保持 GPLv3 兼容边界。

## 研究项目总表

| 项目 | 类型与已验证重点 | 许可证 | 对本项目的可借鉴点 | 取舍结论 |
|---|---|---|---|---|
| [MHSanaei/3x-ui](https://github.com/MHSanaei/3x-ui) | 上游面板；检查了 controller/service、Xray 管理、订阅、CI 和 release 工作流 | GPLv3 | 当前项目继承的协议模型、多节点和 Xray 生命周期基础 | 保持兼容，优先向上游同步通用修复 |
| [Gozargah/Marzban](https://github.com/Gozargah/Marzban) | 直接竞品；检查 Python API、React 管理界面、数据库选项、备份服务和用户批量 API | AGPLv3 | 显式备份服务、用户/管理员 API、批量运维、数据库后端选择 | 借鉴产品流程和 API 设计，重新实现，不引入 AGPL 代码 |
| [hiddify/Hiddify-Manager](https://github.com/hiddify/Hiddify-Manager) | 直接竞品；检查多协议、角色、限额、安装、更新和备份文档 | GPLv3 | 新手引导、角色化运维、自动更新/备份、多服务器视图 | 借鉴 onboarding 和运维入口，保留本项目更深的 Xray 配置能力 |
| [2dust/v2rayN](https://github.com/2dust/v2rayN) | Windows 客户端；检查配置导入、订阅更新、节点测试和跨核心适配 | GPLv3 | 订阅更新反馈、批量节点测试、配置导入错误解释 | 只借鉴客户端可理解的链接和订阅反馈 |
| [2dust/v2rayNG](https://github.com/2dust/v2rayNG) | Android 客户端；检查移动端配置、订阅和连接状态 | GPLv3 | 移动端空状态、连接状态和配置分享的低门槛表达 | 借鉴移动端 UX，不复制客户端实现 |
| [SagerNet/sing-box](https://github.com/SagerNet/sing-box) | 协议核心；检查配置 schema、route、DNS、outbound 和 release 结构 | GPLv3 | 结构化配置、清晰的 route/DNS 边界、版本兼容提示 | 继续把 schema 作为 source of truth，关注 Xray/sing-box 互操作 |
| [XTLS/Xray-core](https://github.com/XTLS/Xray-core) | 当前运行核心；检查 API、配置类型、路由和 release | MPL 系列许可边界 | gRPC 热更新、配置验证、协议能力跟随核心版本 | 通过依赖升级和测试跟进，不直接改核心 |
| [wg-easy/wg-easy](https://github.com/wg-easy/wg-easy) | WireGuard 管理面板；检查 peer CRUD、二维码、容器和状态展示 | 以仓库 LICENSE 为准 | Peer 卡片、二维码、配置下载、最小化运维流程 | 继续强化 WireGuard 友好流程，保持协议模型独立 |
| [Nginx Proxy Manager](https://github.com/NginxProxyManager/nginx-proxy-manager) | Web 反向代理面板；检查代理主机、证书、访问列表、审计和 Docker 部署 | MIT | 资源卡片、证书状态、访问控制和向导式配置 | 借鉴资源状态卡和证书 UX |
| [Cockpit](https://github.com/cockpit-project/cockpit) | 主机管理面板；检查模块化页面、journal、网络、服务和安全响应 | LGPL/GPL/BSD/MIT 混合 | 小而清晰的运维模块、服务状态、测试门禁、安全发布节奏 | 借鉴模块化运维信息架构，不嵌入代码 |
| [Netdata](https://github.com/netdata/netdata) | 主机/网络可观测性；检查 agent、健康检查、发布 manifest、签名和多架构产物 | GPLv3 | 指标命名、健康状态、签名 manifest、资产校验和故障上下文 | 优先借鉴发布供应链和健康模型 |
| [Prometheus](https://github.com/prometheus/prometheus) | 指标与告警；检查 exposition、规则、TLS、版本和签名 tag | Apache-2.0 | `/metrics`、告警规则、稳定标签、不可变签名 release | 采用 Prometheus 兼容指标，不引入 Prometheus 服务本体 |
| [Grafana](https://github.com/grafana/grafana) | 可观测性 UI；检查 dashboard、数据源、release checksums、SBOM 和 CVE 说明 | AGPLv3 与例外 | 仪表盘信息层级、SBOM、校验和、安全发布说明 | 借鉴展示和发布材料，避免 AGPL 代码耦合 |
| [Uptime Kuma](https://github.com/louislam/uptime-kuma) | 黑盒监控；检查探针、状态页、通知、supported tags 和安全公告 | MIT | 面板/Xray/订阅探针、状态页和通知历史 | 可实现轻量内置探针 |
| [Watchtower](https://github.com/containrrr/watchtower) | 容器自动更新；检查 health gate、token file、JSON 日志和多架构 tag | Apache-2.0/BSD | 更新前健康检查、凭据文件、结构化日志和失败停留 | 将 health gate 应用于面板更新，不照搬容器逻辑 |
| [Renovate](https://github.com/renovatebot/renovate) | 依赖自动更新；检查配置迁移、发布节奏、签名和 PR 测试 | AGPLv3 | Go/npm/Xray/mtg 依赖更新、分组策略、自动验证 | 使用同类工作流思想，保留项目自有 CI |
| [GoReleaser](https://github.com/goreleaser/goreleaser) | 发布工程；检查声明式配置、SBOM、provenance、归档和跨平台矩阵 | MIT | 统一产物、checksums、SBOM、provenance 和发布元数据 | 借鉴 release 配置设计，逐步补齐工作流 |

研究数量：17 个项目。

## 深度对标

### 直接产品：Marzban 与 Hiddify Manager

| 对比项目 | 对方优势 | 我们的现状 | 差距 | 建议 |
|---|---|---|---|---|
| 核心功能 | 更聚焦的用户/管理员运维和安装路径 | 协议、Xray 配置、MTProto、多节点更深 | 新手从安装到第一条链接的路径仍可缩短 | 增加 onboarding、状态摘要和任务反馈 |
| 用户体验 | 备份、批量用户、面向运营者的入口更直接 | 功能多，但页面分散，入站与客户端操作密度较高 | 批量动作和失败解释仍可增强 | 引入批量任务结果、dry-run 和清晰空状态 |
| 技术架构 | Marzban 的 API/DB 选项明确；Hiddify 的多服务器叙事清晰 | service/runtime 分层更适合 Xray 热应用和远程节点 | 可观测状态没有统一抽象 | 建立健康、同步和任务状态 DTO |
| 安全性 | Hiddify 的更新/备份入口明显 | CSRF、mTLS、token scope、SSRF 已较完整 | 更新签名和供应链证明不足 | P0 补签名 manifest、SBOM、provenance |
| 结论 | 借鉴产品流程 | 保留 Xray 深度和多节点差异 | 不做技术栈替换 | 重新实现 UI/API 设计 |

### 可观测性：Netdata、Prometheus、Grafana、Uptime Kuma

这些项目共同体现了三个可迁移模式：

1. 所有健康状态都有稳定的名称、时间戳、最近错误和恢复状态，而不是只有一条日志。
2. 指标和告警有机器可读接口，UI 只是其中一个消费者。
3. 发布产物、校验和、SBOM、签名和安全公告是同一条交付链的一部分。

本项目已有事件总线、Xray metrics、节点 heartbeat、tunnel monitor 和本轮新增的只读 Prometheus exposition/告警规则下载，但还没有统一的探针历史与健康摘要。后续继续沿用只读指标路径，避免引入新的存储服务。

### 运维面板：Cockpit、Nginx Proxy Manager、wg-easy

这些项目的共同优势不是功能数量，而是把一个高风险动作拆成清晰的资源卡片、状态标签、确认步骤和可回退结果。当前项目的入站、节点、证书、备份页面已经具备一部分基础，下一步应把同步、更新、恢复和 Xray 重启也表现为同样的可追踪任务。

### 发布与更新：Watchtower、GoReleaser、Renovate、Netdata、Grafana

当前 `update.sh` 已有 HTTPS 下载、镜像回退、SHA-256、归档路径检查、备份和失败回滚，这是基础能力。对标项目进一步提供：

- 发布者身份验证，而不是只有内容摘要。
- SBOM 和 provenance，便于知道产物由哪个提交和构建环境生成。
- 更新前健康检查，避免把无法启动的二进制替换到服务目录。
- 结构化更新日志、明确的支持版本和安全公告。
- 依赖更新自动化，但所有更新都必须经过项目测试。

这些改进不改变现有协议或数据库结构，适合按阶段落地。

## 我们保持的优势

- Xray-core 配置表单和链接生成覆盖面深，且有 schema、golden fixtures 和后端 parity 测试。
- 本地与远程 runtime 统一了状态改变路径，能处理多节点和 mTLS，而许多轻量面板只覆盖单机。
- MTProto 多 secret、广告标签、配额/过期 hot apply 是明确的差异能力。
- PostgreSQL、SQLite 快照、迁移、流量写入和恢复路径比简单 SQLite 面板更完整。
- 事件总线、cron、WebSocket、Xray hot diff 和 TUIC/MTProto 子进程监督形成了较强工程基础。

## 研究后的技术采用决定

| 方案 | 决定 | 原因 |
|---|---|---|
| 签名 manifest + SHA-256 + SBOM | 采用，重新实现 | 直接解决当前更新供应链缺口，依赖成熟工具即可 |
| Prometheus `/metrics` | 采用，新增只读端点 | 与现有 metrics/eventbus 兼容，不需要新数据库 |
| 黑盒健康探针 | 采用，局部实现 | 可复用现有 tunnel monitor 和节点 heartbeat |
| Watchtower 式 health gate | 采用其原则 | 更新失败时保留旧版本，减少不可启动替换 |
| Renovate 式依赖分组 | 采用工作流思想 | 减少依赖升级噪声，仍由 CI 负责验证 |
| 直接复制 Marzban/Hiddify/Grafana/Cockpit 代码 | 不采用 | 许可证、架构和维护边界不适合直接嵌入 |
| 更换 React/Go/Xray 技术栈 | 不采用 | 当前 schema、测试和 runtime 投资较大，替换会扩大回归面 |
