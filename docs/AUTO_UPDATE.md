# 自动更新与发布

本项目的在线更新适用于 Linux 安装版面板。更新入口位于面板首页的版本信息中，后端通过 GitHub Releases 查询版本，并在中国大陆环境下保留 Gitee 镜像回退路径。

## 更新流程

1. 面板调用 GitHub Releases API 查询最新稳定版；启用 Dev 渠道时查询固定标签 `dev-latest`。
2. 用户确认后，面板以独立进程启动 `update.sh`，并通过 `update-status.json` 回传 `pending`、`success` 或 `failed`。
3. 更新脚本从 GitHub/Gitee 下载当前架构的归档和 `SHA256SUMS`。
4. 归档必须通过 SHA-256 校验、tar 路径安全检查，并包含 `x-ui/x-ui` 主程序后，才会进入安装阶段。
5. 校验失败、归档损坏或下载失败时，旧版本不会被替换；安装阶段会先把旧目录和服务文件放入同文件系统的临时回滚目录，更新后的服务启动失败时自动恢复旧版本。

Windows 构建可以通过 Release 下载，但面板自更新脚本目前只支持 Linux 服务安装，不会在 Windows 上执行静默替换。

## GitHub Actions 发布

`.github/workflows/release.yml` 负责构建 Linux 和 Windows 产物，并在发布阶段统一生成 `SHA256SUMS`：

- 推送 `vMAJOR.MINOR.PATCH` 标签：创建正式 Release，上传全部归档和校验清单。
- 推送 `main`：更新固定的 `dev-latest` 预发布 Release，正文包含完整 `commit=<sha>` 标记，供 Dev 渠道比较提交。
- Release 工作流使用 `GITHUB_TOKEN` 的 `contents: write` 权限，不需要额外的发布密钥。
- 日常开发和自动构建统一在 `main` 分支；其他分支不会触发该发布工作流。正式发布需要显式创建版本标签。

发布前检查：

```bash
git tag v3.13.0
git push origin v3.13.0
```

Release 发布后，应确认归档、`SHA256SUMS`、目标架构和 Release 类型均正确。工作流不会替用户配置代码签名证书。

## 安全边界

- 下载地址只允许固定的 GitHub/Gitee 公共域名；脚本拒绝本地地址、IP 字面量和不安全路径。
- 归档在验证 SHA-256 之前不会安装。
- Release 资产使用 HTTPS 传输；校验清单与归档来自同一发布标签，并在主/回退镜像间自动切换。
- 当前没有接入 Windows Authenticode 或 Linux 二进制签名验证。SHA-256 可以检测传输损坏和非预期替换，但不能替代签名验证；正式生产环境应在具备证书后再启用签名门禁。
- 更新脚本由固定 `main` 分支的 `update.sh` 提供，因此修改更新脚本本身仍需要可信的代码仓库、分支保护和发布审查。

## 失败处理与恢复

下载失败、校验失败、归档结构错误和不支持的平台会在替换前退出，并通过状态接口报告失败。旧版本保持原位。

如果服务停止后发生系统级故障，脚本会优先恢复同文件系统回滚目录中的旧安装。若主机在恢复过程中断电，使用服务器控制台或 SSH 重新运行安装脚本/指定版本安装；不要删除数据库目录。数据库位于配置的 `XUI_DB_FOLDER`，更新过程不应覆盖它。

## 测试与限制

本地可验证：

- `bash -n update.sh`
- 前端类型检查、Lint、Vitest 和生产构建
- 发布工作流的 YAML 变更审查

当前环境没有 Go 工具链，也不是 Linux systemd 主机，因此无法在本机真实执行服务停止、归档替换、重启和回滚流程；这些属于发布前的外部验收项，不能标记为已通过。
