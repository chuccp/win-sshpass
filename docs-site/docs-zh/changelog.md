# 更新日志

## v0.3.2

- 新增 Linux 支持（amd64、arm64）
- 新增代理支持：SOCKS5/SOCKS4/HTTP/HTTPS 代理隧道
- 新增 SFTP 断点续传（`-resume` 参数）
- 新增 hash 和 verify 子命令（md5、sha1、sha256、sha512）
- 新增 WinGet 安装支持
- 改进代理超时处理
- 初始文档站点

## v0.3.1

- 修复 SFTP 传输中的超时重置问题
- 改进 rz/sz 文件传输的稳定性

## v0.3.0

- 提取 SDK 包和 CLI 入口点
- 新增 `WithProgress`、`WithFileSelector`、`WithSignalHandler` 选项
- SDK 不再包含 UI 代码，CLI 侧适配器独立实现

## v0.2.0

- 新增 SCP 风格传输支持
- 新增 Rsync 风格传输支持
- 新增配置文件支持
- 新增连接重试和指数退避

## v0.1.0

- 初始版本
- SSH 密码/密钥认证
- 交互式 Shell（原始终端模式）
- SFTP 文件传输
- 动态终端调整大小
- rz/sz Shell 内传输
