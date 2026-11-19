# 平台与维护

面板是运行在服务器上的控制平面。浏览器管理入站、客户端、订阅、节点、流量、备份及配置。代理连接由同主机或远程节点上的 Xray 提供，MTProto 由对应的侧车进程承载。账号、权限、会话和数据始终以服务端为准。Go/Gin/GORM 与 React/Ant Design 的边界见 [架构](architecture.md)。

| 平台 | 判断 | 功能、权限与分发维护 |
| --- | --- | --- |
| Ubuntu / Debian | A | Linux 原生服务，systemd、root 部署，可访问 SQLite/PostgreSQL 和运行时目录。amd64、arm64 有容器安装 smoke，amd64 另验实际归档启动。签名归档更新。 |
| Fedora / Arch Linux | A | 复用 Linux 服务和发行版 unit，包管理器/服务权限各自处理。沿用多架构 tar.gz 和签名更新。发行版原机验收独立记录。 |
| 其他 Linux（含 Alpine） | A | Debian/RHEL/Arch 系及 Alpine OpenRC 部署沿用现有脚本。高阶能力取决于侧车资产与主机权限。旧架构保持构建覆盖，逐个记录实机验收。 |
| Windows 10 / 11 x64 | B | Go/CGo Windows 便携服务，浏览器管理。同一 API 和数据模型，保留 Windows 运行时文件。归档启动由 Windows Runner 检查。后台托管、自动安装与 Authenticode 需要额外独立维护。 |
| Windows ARM64 | C | CGo 和各侧车需单独准备和验证；可在浏览器使用远程部署。 |
| macOS Intel / Apple Silicon | C | 原生服务需要两套 CGo/侧车、launchd、签名公证及系统兼容测试。日常管理直接使用远程 Web。 |
| Android 手机 / Android 平板 / 折叠屏 | A（Web），C（原生） | 浏览器可使用全部服务端管理权限，布局按视口适配。原生客户端需要单独处理窄屏、凭据存储、多实例、签名和更新；移动系统上的代理核心与管理服务器属于不同部署。 |
| iOS / iPadOS | A（Web），C（原生） | Safari 管理已部署实例。原生应用增加本地凭据和商店发布成本，后台服务不能承担服务器的监听和进程生命周期。 |
| 桌面 Web / 移动 Web / 响应式 Web | A | 同一服务提供带指纹的前端资产、后端 API、权限和存储。随服务或容器更新，无独立客户端版本。 |
| PWA | C | 离线时无法运行服务端管理；缓存面板和订阅数据需要额外隔离、更新和注销策略。收益低于现有响应式 Web。 |
| Chrome / Edge / Firefox 扩展 | C | 核心功能均经由站点提供；扩展会增加跨域权限、凭据存储及三个商店的审核和更新成本。 |
| Docker / NAS | A | Linux 容器，使用多架构 GHCR 标签、运行时端口和持久卷。NAS 须支持对应架构和容器网络；旁路或低端硬件需另验。 |
| 私有化部署 / 云服务器 | A | 使用 Linux 或 Docker 分发，可结合 HTTPS、访问控制、持久化和数据库备份。 |
| CLI | A | Go CLI 管理设置、证书和迁移；Linux 管理菜单处理服务启停和更新。Windows 使用 Go CLI。 |
| Linux 桌面 AppImage / DEB / RPM / Flatpak | C | 本轮沿用现有服务安装方式。每种分发形态都需要处理服务权限、侧车与更新，其维护成本高于当前路径。 |

A 直接维护，B 作为现有平台改进，C 暂缓。Linux Release 构建 amd64、arm64、armv7、armv6、armv5、386、s390x；Docker 构建 amd64、arm64、armv7、armv6、386；Windows Release 构建 amd64。以具体版本的工作流结果为验收证据。移动 Web 的改进应优先覆盖搜索、列表密度、触控和批量任务回执。

## 维护节奏

- 每次改动通过 `main` 的 CI。正式版由语义版本标签发布，稳定标签受保护。更新接口、数据模型、API 及相应平台共同验收。
- 每 30 天查看 Releases、CodeQL、Dependabot 和 [竞品记录](../COMPETITOR_ANALYSIS.md)，记录变化和实际处置。依赖升级经测试后进入 `main`。
- 每次正式发布记录归档矩阵、镜像、校验清单、签名验证、安装 smoke 及数据库恢复结果。Windows Authenticode、Linux OpenRC 实机升级、PostgreSQL 跨版本恢复和全架构运行验收按证据更新。
- 当前长期改进项和验收定义见 [路线图](../UPGRADE_ROADMAP.md)。历史运行证据见 [交付报告](../UPGRADE_REPORT.md)。
