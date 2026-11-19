# 自动更新与发布

本指南适用于 v3.13.1 起的 GitHub 发行归档。Linux 服务安装使用面板首页或 `x-ui update`。Windows 维护版是便携服务，Docker 和 NAS 使用镜像更新。

## Linux 服务更新

1. 从 [Sigstore 官方安装指南](https://docs.sigstore.dev/cosign/system_config/installation/)选择适用于主机架构的 Cosign，并确保面板服务和管理员的 PATH 可找到 `cosign`。Cosign 需要访问其信任根服务。
2. 备份面板数据库和本地配置。SQLite 可使用面板数据库备份。PostgreSQL 应保留与本次更新对应的可恢复快照。
3. 在面板首页检查新版本并确认，或在服务器上执行 `x-ui update`。开发渠道沿用 `dev-latest`，默认使用正式版。
4. 更新器验证指定正式标签或开发分支的 GitHub Actions 身份，再比对当前架构归档的 SHA-256。Gitee 和 GitHub 可互为下载源，每次尝试都使用同源的签名、清单和归档。
5. 下载完成后，更新器检查归档内容；解包并运行新版本的 `-v` 检查。只有这些检查通过才停止旧服务。面板会轮询当前操作的结果。

从 v3.13.0 首次迁移时，旧面板的 Web 更新入口仍会从仓库下载一次更新脚本。如果希望在执行之前自行验证，可用下节命令校验新版本归档，再从已验证的归档运行其中的 `update.sh`。升级到 v3.13.1 后，更新入口使用安装目录里的脚本。

## 独立验证归档

从同一 [GitHub Release](https://github.com/yuchen0x1/Xray-server-management-panel/releases)下载目标归档、`SHA256SUMS` 和 `SHA256SUMS.sigstore.json`。下面以正式版和 Linux amd64 为例。将标签和文件名改为实际目标。

```bash
tag=v3.13.1
asset=x-ui-linux-amd64.tar.gz
cosign verify-blob --bundle SHA256SUMS.sigstore.json \
  --certificate-identity "https://github.com/yuchen0x1/Xray-server-management-panel/.github/workflows/release.yml@refs/tags/$tag" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com SHA256SUMS &&
grep -F "  $asset" SHA256SUMS | sha256sum --check --strict
```

开发版验证时使用 `refs/heads/main` 身份。旧版本可能只有 SHA-256。签名无效或缺少签名时，新版 Linux 自动更新会停止并保留运行中的版本。

## Windows、Docker 和首次安装

- Windows amd64 可用 Cosign 验证同一清单，再用 PowerShell `Get-FileHash -Algorithm SHA256` 核对 ZIP。停止面板并备份数据库和配置后解包。证书依赖和保存位置见归档内说明。ZIP 的签名证明不提供 Authenticode。
- Docker 和 NAS 在保留持久卷的前提下拉取指定版本的 GHCR 镜像并重建，按镜像标签和架构验收。面板内的 Linux 更新器适用于 systemd/OpenRC 服务安装。容器镜像签名由独立发布项跟进。
- 首次安装脚本保留历史版本的 SHA-256 兼容路径。需要签名验收的部署应在启动任何归档程序前使用上面的独立验证步骤。

## 失败处理

下载、签名、摘要、兼容性或解包失败时，旧服务留在原位。服务替换开始后，如果新服务启动失败，脚本尝试恢复旧安装和服务文件，会在日志中报告回滚文件位置。数据库位于单独配置的数据目录。应用程序回滚不会撤销已提交的数据库迁移。

使用 `journalctl -u 'x-ui-web-update-*'` 查看通过 systemd 调度的更新日志。通过终端执行更新时，终端会显示签名校验、下载和失败阶段。经历断电或强制终止后，先检查安装父目录的 `.x-ui-update-backup.*` 和数据库备份，再执行恢复。

## 发布流程

推送 `main` 生成开发预发行。正式版标签必须与 `internal/config/version` 一致并指向 `main` 的历史，发布等待完整 CI、Linux 和 Windows 归档矩阵，以及 Linux amd64 和 Windows amd64 解包启动检查。工作流生成并验证签名后，将资产上传到草稿再公开。已公开正式版的同名资产受工作流保护。

发布后检查 GitHub 下载、标签、签名和 CI，以及 Docker 架构与 Linux 安装 smoke。支持矩阵、维护节奏和已知边界见 [平台与维护](PLATFORMS.md)。
