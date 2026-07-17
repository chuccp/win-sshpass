# win-sshpass 使用手册

> 跨平台 sshpass 实现（Windows 和 Linux）：密码/密钥 SSH 登录、交互式 Shell、SFTP 文件传输、SCP/Rsync 风格传输、代理隧道、断点续传、文件哈希校验，以及可复用的 Go SDK。

## 什么是 win-sshpass？

win-sshpass 是 Linux sshpass 工具的跨平台实现，支持 Windows 和 Linux。它是一个独立的可执行文件，无需安装 OpenSSH 或其他依赖，下载即用。

它同时支持**命令行工具**和**Go SDK**两种使用方式：

- **密码/密钥认证**：支持密码、私钥、环境变量、配置文件等多种认证方式。
- **交互式 Shell**：原始终端模式，支持 vim、top、htop 等全屏应用，动态终端调整大小。
- **SFTP 文件传输**：上传/下载文件和目录，带进度条显示。
- **SCP/Rsync 风格**：兼容 scp 和 rsync 命令语法。
- **代理支持**：通过 SOCKS5/SOCKS4/HTTP/HTTPS 代理隧道连接 SSH。
- **断点续传**：中断的 SFTP 文件传输可从断点处恢复。
- **文件哈希与校验**：计算和校验本地文件哈希（MD5、SHA-1、SHA-256、SHA-512）。
- **可复用 Go SDK**：作为 Go 库导入，嵌入到你自己的应用中，支持注入 I/O 流和进度回调。

## 30 秒快速体验

```bash
# 通过 WinGet 安装
winget install chuccp.win-sshpass

# 通过 Scoop 安装
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# 密码登录并执行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 上传文件
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 打开交互式 Shell
win-sshpass -p 'password' ssh user@host
```

## 核心功能

### 1. 多种认证方式

**密码认证**：直接指定密码、从文件读取、从环境变量读取。

```bash
# 直接指定密码
win-sshpass -p 'secret' ssh user@host

# 从文件读取密码
win-sshpass -f pass.txt ssh user@host

# 从环境变量读取
SSHPASS='secret' win-sshpass -e ssh user@host
```

**私钥认证**：支持 Ed25519、RSA 等密钥格式。

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

**配置文件**：管理多台服务器的连接信息。

```bash
win-sshpass -f server.config -c 'docker ps'
```

### 2. 交互式 Shell（原始终端模式）

当不指定命令时，win-sshpass 打开交互式 Shell，支持：

- **正确的回显** — 输入的字符正确显示（无双重回显）
- **Ctrl+C / Ctrl+Z** — 信号正确转发到远程进程
- **全屏应用** — vim、top、htop、nano 等正常工作
- **动态终端调整** — 远程终端自动匹配本地窗口大小
- **Tab 补全** — 远程 Shell 的 Tab 补全正常工作

```bash
win-sshpass -p 'password' ssh user@host
```

### 3. 文件传输

**SFTP 方式**：上传/下载文件和目录，带进度条。

```bash
# 上传文件
win-sshpass -h host -p 'pass' -local file.txt -remote /tmp/file.txt

# 上传多个文件
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 下载目录
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

**SCP 风格**：兼容 scp 命令语法。

```bash
win-sshpass -p 'pass' scp ./app.jar user@server:/opt/app/
win-sshpass -p 'pass' scp -r ./dist user@server:/var/www/html
```

**Rsync 风格**：兼容 rsync 命令语法。

```bash
win-sshpass -p 'pass' rsync -avz ./ user@server:/backup/
```

**Shell 内 rz/sz**：在交互式 Shell 中直接使用 rz/sz 命令。

```bash
# 在远程 Shell 中
rz                              # 上传文件（打开文件选择器）
sz /remote/path/to/file         # 下载文件
```

### 4. 可复用 Go SDK

win-sshpass 也是一个 Go 库（`package sshpass`），可以嵌入到你自己的应用中：

```go
import sshpass "github.com/chuccp/win-sshpass"

cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
defer client.Close()

// 执行命令
client.Exec("uname -a")

// SFTP 传输
sftp, _ := client.SFTP()
sftp.Upload("./local.txt", "/tmp/remote.txt")
```

SDK 不包含任何 UI 代码（无进度条、无文件对话框），通过函数选项注入：

| 选项 | 用途 |
|------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | 重定向 I/O 流 |
| `WithProgress(fn)` | 设置传输进度回调 |
| `WithFileSelector(s)` | 设置 rz/sz 文件选择器 |
| `WithResume()` | 启用文件传输断点续传 |
| `WithSignalHandler()` | 注册 Ctrl+C 信号处理 |

## 参数一览

| 参数 | 说明 | 示例 |
|------|------|------|
| `-p` | 密码 | `-p 'secret123'` |
| `-i` | 私钥路径 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密码文件 / 配置文件 | `-f pass.txt` |
| `-e` | 从 SSHPASS 环境变量读取密码 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主机地址 | `-h example.com` |
| `-u` | 用户名，默认 root | `-u ubuntu` |
| `-P` | 端口，默认 22 | `-P 2222` |
| `-c` | 要执行的命令 | `-c 'ls -la'` |
| `-local` | 本地路径（逗号或空格分隔） | `-local "a.txt,b.txt"` |
| `-remote` | 远程路径 | `-remote /tmp/file.txt` |
| `-d` | 下载模式 | `-d` |
| `-k` | 启用严格主机密钥验证 | `-k` |
| `-t` | 操作超时（秒），0 表示无限制 | `-t 30` |
| `-ct` | TCP 连接超时（秒），默认 10 | `-ct 5` |
| `-retry` | 连接重试次数，默认 3 | `-retry 5` |
| `-resume` | 从断点恢复中断的文件传输 | `-resume` |
| `-proxy` | 代理 URL（socks5/socks4/http/https） | `-proxy socks5://127.0.0.1:1080` |
| `-v` | 显示版本 | `-v` |
| `-help` | 显示帮助 | `-help` |

## 快速链接

### 快速开始

- [安装](getting-started/installation.md) - 下载与安装方式
- [快速开始](getting-started/quick-start.md) - 第一个连接

### 用户指南

- [SSH 连接](guide/ssh.md) - 密码、私钥、环境变量认证
- [文件传输](guide/file-transfer.md) - SFTP 上传下载
- [交互式 Shell](guide/shell.md) - 原始终端模式与 rz/sz
- [SCP 与 Rsync](guide/scp-rsync.md) - 兼容 scp/rsync 语法
- [配置文件](guide/config-file.md) - 管理多台服务器

### 高级与参考

- [Go SDK](advanced/sdk.md) - 作为 Go 库使用
- [最佳实践](advanced/best-practices.md) - 安全与效率建议
- [更新日志](changelog.md)

## 社区

- [GitHub](https://github.com/chuccp/win-sshpass)
- [问题反馈](https://github.com/chuccp/win-sshpass/issues)
