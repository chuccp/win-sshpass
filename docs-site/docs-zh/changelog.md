# 更新日志

## v0.8.1

- 新增 SSH 密钥对生成子命令（`keygen`）—— 支持 Ed25519 和 RSA，部署公钥以实现免密登录
- 新增基于 Docker 的集成测试套件（71 项测试覆盖所有功能）
- 文档站点更新为 go-web-frame 风格，支持 Material 图标、样式化按钮和更丰富的主题
- 新增密钥生成 SDK 函数：`GenerateKeyPair`、`GenerateRSAKeyPair`、`SaveKeyPair`、`DeployPublicKey`、`DefaultKeyPath`
- 更新 4 种语言的 README 和文档站点，新增 keygen、Docker 测试、代理和 hash/verify 章节
- 多项文档修复和改进

## v0.7.2

- 使用最新功能更新所有 README 和文档站点
- 在发布工作流中启用 WinGet 自动提交
- 新增 Linux 交叉编译支持
- 新增代理支持：SOCKS5/SOCKS4/HTTP/HTTPS 代理隧道（`-proxy` 参数）
- 新增 SFTP 文件传输断点续传（`-resume` 参数）
- 新增 hash 和 verify 子命令（MD5、SHA-1、SHA-256、SHA-512）
- 修复代理超时处理，附全面测试覆盖

## v0.7.1

- 新增 MkDocs 文档站点，支持英语、简体中文、繁体中文和日语翻译
- 提取 SDK 包（`package sshpass`）和 CLI 入口点（`cmd/sshpass`）
- 新增 Scoop 安装说明
- 为所有 README 添加 Star 提醒

## v0.6.4

- 更新发布工作流

## v0.6.3

- 更新发布工作流和流水线
- 新增交互式 Shell 中的 rz/sz 文件传输支持（带 SFTP 回退）
- 使用最新功能同步所有 README

## v0.6.2

- 为所有 README 添加交互式 Shell 和配置文件位置命令示例
- 修复原始终端模式（正确的回显、信号转发、全屏应用支持）
- 修复错误退出码和配置文件命令处理
- 修复指数退避溢出
- 重构以消除代码重复并提高健壮性

## v0.6.1

- 新增 ARM64 架构构建支持
- 在文档中添加 ARM64 下载表
- 在配置中添加端口号验证

## v0.5.1

- 新增可配置的连接超时（`-ct`）和操作超时（`-t`）
- 多项错误修复和改进
- 修复 Git Bash 路径转换示例（`//` 前缀）

## v0.4.x

- WiX v7 MSI 安装程序支持
- macOS DMG/pkg 打包
- 带指数退避的连接重试
- 带进度条的 SFTP 上传/下载
- SCP 风格和 Rsync 风格命令语法
- 配置文件支持（`-f`）
- 带原始终端模式的交互式 Shell
- 动态终端大小调整（SIGWINCH / 轮询）

## v0.3.x

- macOS .pkg 安装程序和 .tar.gz 打包
- 流水线改进

## v0.2.x

- SCP 和 Rsync 传输支持
- 配置文件支持
- 带指数退避的连接重试
- 构建流水线和发布自动化

## v0.1.0

- 初始版本
- SSH 密码和私钥认证
- 远程命令执行
- 带原始终端模式的交互式 Shell
- SFTP 文件传输
- 动态终端大小调整
