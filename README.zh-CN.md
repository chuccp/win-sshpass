## 相关项目

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — 轻松解决鉴权问题——路由声明需要什么权限，Filter 一处校验，handler 里干干净净。泛型 Model 解决全栈 CRUD——定义好 struct，增删改查直接能用。轻巧，需要的组件按需安装。不需要代码生成，无需CI工具，目前最精巧得go web全栈框架。

# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

跨平台 sshpass 工具（Windows 和 Linux），实现类似 Linux sshpass 的功能。

> 💡 **如果这个项目对你有帮助，欢迎点个 ⭐ Star！** 让更多人发现这个工具。

## 功能特性

- 支持密码或私钥认证的 SSH 登录
- 执行远程命令或打开交互式 Shell
- 通过 SFTP 上传/下载文件（带进度条）
- SCP 风格和 Rsync 风格的文件传输
- 配置文件支持，方便管理多台服务器
- 交互式 Shell 使用 raw 终端模式（正确的回显、Ctrl+C、vim/top 全屏程序支持）
- 交互式 Shell 模式下动态调整终端大小
- Git Bash 路径转换检测与自动修复
- 支持 IPv6 地址
- 支持 Windows（x64、ARM64）、Linux（amd64、arm64）和 macOS（amd64、arm64）
- **可复用 Go SDK** — 作为库引入（`package sshpass`），在自有应用中嵌入 SSH/SFTP/Shell 能力，支持注入 I/O 流与进度回调
- **代理支持** — 通过 SOCKS5/SOCKS4/HTTP/HTTPS 代理隧道连接 SSH
- **断点续传** — 中断的 SFTP 文件传输可从断点处恢复
- **文件哈希与校验** — 计算和校验本地文件哈希（MD5、SHA-1、SHA-256、SHA-512）
- **密钥生成** — SSH 密钥对生成（Ed25519/RSA）

## 下载

从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下载最新版本：

### Windows

| 架构 | Zip | MSI 安装包 |
|------|-----|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

### Linux

| 架构 | Tarball |
|------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

### macOS

| 架构 | PKG 安装包 | Tarball |
|------|-----------|---------|
| **amd64 (Intel)** | `win-sshpass-*-darwin-amd64.pkg` | `win-sshpass-*-darwin-amd64.tar.gz` |
| **arm64 (Apple Silicon)** | `win-sshpass-*-darwin-arm64.pkg` | `win-sshpass-*-darwin-arm64.tar.gz` |

> `.pkg` 安装包会自动将二进制文件安装到 `/usr/local/bin/win-sshpass`。

1. 打开 [Releases](https://github.com/chuccp/win-sshpass/releases) 页面
2. 下载对应平台和架构的安装包
3. **Windows MSI / macOS PKG**：运行安装程序即可，二进制文件会自动添加到系统 PATH
4. **Windows Zip / Linux tar.gz / macOS tar.gz**：解压后将二进制文件放入 PATH 目录

> **零依赖**：`win-sshpass.exe` 是一个独立的可执行文件，无需安装 OpenSSH 或任何其他软件。下载后放入 PATH 目录即可直接使用。

### 通过 Scoop 安装

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 通过 WinGet 安装

```bash
winget install chuccp.win-sshpass
```

## 快速开始

```bash
# 密码登录执行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 私钥登录执行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# 上传文件
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 下载文件
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## 交互式 Shell

不指定命令时，`win-sshpass` 会打开一个 **raw 终端模式** 的交互式 Shell：

```bash
win-sshpass -p 'password' ssh user@host
```

**Raw 终端模式** 特性：

- **正确的回显** — 输入的字符正确显示（不会出现双重回显）
- **Ctrl+C / Ctrl+Z** — 信号正确转发到远程进程
- **全屏程序** — vim、top、htop、nano 等全屏应用正常运行
- **动态终端大小调整** — 远程终端自动匹配本地窗口大小
- **Tab 补全** — 远程 Shell 的 Tab 补全功能正常工作

### 交互式 Shell 中的文件传输

连接状态下，使用 `rz` / `sz` 命令传输文件（远程服务器无需安装任何软件）：

```bash
# 上传文件到远程当前目录（弹出文件选择器）
rz

# 上传指定本地文件
rz /本地/文件/路径

# 下载远程文件（弹出保存对话框）
sz /远程/文件/路径

# 下载远程文件到指定本地路径
sz /远程/文件/路径 /本地/保存/路径
```

> **原理**：当远程 Shell 报 `rz`/`sz: command not found` 时，工具自动拦截并通过 SFTP 完成传输。支持文件和目录，带进度条。

## 命令格式

### SSH 登录

```bash
# 密码认证
win-sshpass -p <密码> ssh [user@host] [命令]
win-sshpass -p <密码> ssh -p <端口> user@host '命令'
win-sshpass -p <密码> ssh -o StrictHostKeyChecking=no user@host

# 交互式 Shell（raw 终端模式：正确的回显、Ctrl+C、vim/top 支持）
win-sshpass -p <密码> ssh user@host

# 私钥认证
win-sshpass -i <私钥路径> ssh [user@host] [命令]

# 环境变量密码
SSHPASS=<密码> win-sshpass -e ssh user@host

# 密码文件
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 配置文件（多行格式）
win-sshpass -f server.config
```

### 文件传输

> **Git Bash 用户**：远程路径需使用 `//` 前缀，例如 `-remote //tmp/file.txt`。详见下方 [Git Bash 注意事项](#git-bash-注意事项)。

```bash
# 上传文件
win-sshpass -h <主机> -p <密码> -local <本地路径> -remote <远程路径>

# 上传多个文件（逗号分隔）
win-sshpass -h <主机> -p <密码> -local "a.txt,b.txt,c.txt" -remote //tmp/

# 上传多个文件（空格分隔，仅适用于不含 / 或 \ 的简单路径）
win-sshpass -h <主机> -p <密码> -local "a.txt b.txt c.txt" -remote //tmp/

# 上传目录（自动递归）
win-sshpass -h <主机> -p <密码> -local <本地目录> -remote <远程目录>

# 下载文件/目录
win-sshpass -h <主机> -p <密码> -d -remote <远程路径> -local <本地路径>
```

### SCP 风格

```bash
# 上传文件
win-sshpass -p <密码> scp <本地文件> user@host:<远程路径>
win-sshpass -p <密码> scp -P <端口> <本地文件> user@host:<远程路径>

# 上传目录
win-sshpass -p <密码> scp -r <本地目录> user@host:<远程路径>

# 下载文件/目录
win-sshpass -p <密码> scp user@host:<远程文件> <本地路径>
```

### Rsync 风格

```bash
# 上传
win-sshpass -p <密码> rsync -avz <本地路径> user@host:<远程路径>

# 下载
win-sshpass -p <密码> rsync -avz user@host:<远程路径> <本地路径>
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-p` | 密码 | `-p 'secret123'` |
| `-i` | 私钥路径 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密码文件/配置文件 | `-f pass.txt` |
| `-e` | 从环境变量 SSHPASS 读密码 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主机地址 | `-h example.com` |
| `-u` | 用户名，默认 root | `-u ubuntu` |
| `-P` | 端口，默认 22 | `-P 2222` |
| `-c` | 执行的命令 | `-c 'ls -la'` |
| `-local` | 本地路径（逗号或空格分隔） | `-local "a.txt,b.txt"` |
| `-remote` | 远程路径（上传/下载） | `-remote /tmp/file.txt` |
| `-d` | 下载模式 | `-d` |
| `-k` | 启用严格主机密钥验证 | `-k` |
| `-t` | 总操作超时时间（秒），0 表示不限 | `-t 30` |
| `-ct` | TCP 连接超时时间（秒），默认 10 | `-ct 5` |
| `-retry` | 总连接尝试次数（默认：3） | `-retry 5` |
| `-resume` | 从断点恢复中断的文件传输 | `-resume` |
| `-proxy` | 代理 URL（socks5/socks4/http/https） | `-proxy socks5://127.0.0.1:1080` |
| `-algo` | 密钥算法（ed25519/rsa），默认 ed25519 | `-algo rsa` |
| `-out` | 密钥输出路径前缀，默认 id_ed25519 | `-out ~/.ssh/mykey` |
| `-comment` | 密钥注释 | `-comment "my-laptop"` |
| `-v` | 显示版本 | `-v` |
| `-help` | 显示帮助信息 | `-help` |

## 哈希与校验

无需 SSH 连接，即可计算和校验本地文件哈希：

```bash
# 计算哈希
win-sshpass hash md5 ./file.iso
win-sshpass hash sha256 ./file.iso

# 校验文件
win-sshpass verify sha256 d1dc38f6df... ./file.iso
# 输出: OK  (或: FAILED)
```

支持的算法：`md5`、`sha1`、`sha256`、`sha512`。

## 密钥生成

无需 SSH 连接，即可生成 SSH 密钥对：

```bash
win-sshpass keygen                  # 生成 Ed25519 密钥（默认）
win-sshpass keygen -algo rsa        # 生成 RSA 密钥
win-sshpass keygen -out ~/.ssh/mykey   # 指定输出路径
win-sshpass keygen -comment "my-laptop" # 添加注释
```

支持的算法：`ed25519`、`rsa`。

生成的文件：
- `<名称>` — 私钥
- `<名称>.pub` — 公钥

默认输出：`id_ed25519` 和 `id_ed25519.pub`（或 `id_rsa` 和 `id_rsa.pub`）。

**部署公钥以实现无密码登录：**

```bash
# 将公钥内容读入变量，再通过 SSH 部署
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'password' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

# 然后使用私钥登录
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

## 配置文件格式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # 可选，使用私钥代替密码
# timeout: 0              # 可选，总操作超时时间（秒），0 表示不限
# connect_timeout: 10     # 可选，TCP 连接超时时间（秒）
# strict_host_key: false  # 可选，启用严格主机密钥验证
# proxy: socks5://user:pass@127.0.0.1:1080  # 可选，代理 URL（socks5/socks4/http/https）
```

使用方式：
```bash
win-sshpass -f server.config -c 'ls -la'
win-sshpass -f server.config 'ls -la'
```

## 完整示例

```bash
# 1. 密码登录执行命令
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 私钥登录执行 sudo 命令
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. 上传整个目录到服务器
win-sshpass -h server.com -p 'mypass' -local ./dist -remote //var/www/html

# 4. 下载服务器日志目录
win-sshpass -h server.com -p 'mypass' -d -remote //var/log/nginx -local ./logs

# 5. SCP 上传文件
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 环境变量传递密码（更安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'

# 7. 操作超时（30 秒后自动中断）
win-sshpass -p 'mypass' -t 30 ssh user@server.com 'long-running-command'

# 8. 配置文件 + 位置参数命令
win-sshpass -f server.config 'docker ps'

# 9. 断点续传上传
win-sshpass -p 'mypass' -h server.com -local ./bigfile.iso -remote //data/bigfile.iso -resume

# 10. 计算文件哈希
win-sshpass hash sha256 ./download.iso

# 11. 校验文件完整性
win-sshpass verify sha256 d1dc38f6dfb1e4c8... ./download.iso

# 12. 生成 Ed25519 密钥对
win-sshpass keygen -out ~/.ssh/my_key -comment "my-work-laptop"

# 13. 生成 RSA 密钥对
win-sshpass keygen -algo rsa -out ~/.ssh/my_rsa_key
```

## 代理支持

通过代理服务器建立 SSH 隧道连接。支持协议：SOCKS5、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

```bash
# SOCKS5 代理
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@host

# SOCKS5 带认证
win-sshpass -p 'pass' -proxy socks5://proxyuser:proxypass@127.0.0.1:1080 ssh user@host

# SOCKS4 代理
win-sshpass -p 'pass' -proxy socks4://192.168.1.1:1080 ssh user@host

# HTTP CONNECT 代理
win-sshpass -p 'pass' -proxy http://proxy.local:8080 ssh user@host

# HTTPS CONNECT 代理（带认证）
win-sshpass -p 'pass' -proxy https://user:pass@proxy.local:8443 ssh user@host

# 代理 + 文件传输
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 -h host -local ./file.txt -remote /tmp/file.txt

# 代理 + SCP
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 scp ./app.jar user@host:/opt/app/

# 配置文件中设置代理
# proxy: socks5://user:pass@127.0.0.1:1080
```

## Git Bash 注意事项

远程路径用 `//` 开头避免路径转换：
```bash
# 错误：/tmp 会被转换为 Windows 路径
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## 作为 Go SDK 使用

`win-sshpass` 也是一个可复用的 Go 库（`package sshpass`）。引入它即可在自有应用中嵌入 SSH/SFTP/Shell 能力：

```bash
go get github.com/chuccp/win-sshpass
```

```go
package main

import (
	"log"

	sshpass "github.com/chuccp/win-sshpass"
)

func main() {
	cfg := sshpass.NewConfig()
	cfg.Host = "example.com"
	cfg.User = "root"
	cfg.Password = "secret"

	// NewClient 拨号并返回一个即开即用的客户端。
	client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 执行命令（输出默认流向 os.Stdout/os.Stderr）。
	if err := client.Exec("uname -a"); err != nil {
		log.Fatal(err)
	}

	// 通过 SFTP 上传文件。
	sftp, err := client.SFTP()
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()
	if err := sftp.Upload("./local.txt", "/tmp/remote.txt"); err != nil {
		log.Fatal(err)
	}
}
```

### 自定义选项

通过传入 `NewClient` 的函数式选项配置行为：

| 选项 | 用途 |
|------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | 重定向 I/O 流（默认 `os.Stdin`/`os.Stdout`/`os.Stderr`）。 |
| `WithProgress(fn)` | 设置 `ProgressFunc` 回调，在 SFTP 传输时接收 `(description string, sent, total int64)`。SDK 自身不做任何渲染，由调用方决定如何展示进度。默认不设置（适合无头环境）。 |
| `WithFileSelector(s)` | 设置 rz/sz 文件传输回退用的 `FileSelector`。SDK 不提供默认实现；未设置时 rz/sz 从 stdin 读取路径。 |
| `WithSignalHandler()` | 注册 Ctrl+C 处理器以关闭连接。默认不注册，库不会干扰宿主进程的信号处理。 |

SDK 有意**不内置任何 UI 代码**（无进度条、无文件对话框）。这些职责位于 CLI 包
（`cmd/sshpass/ui.go`），它将基于 progressbar 的 `ProgressFunc` 和基于 zenity 的
`FileSelector` 接入客户端。库用户需自行提供。

如需通过代理隧道连接 SSH，在调用 `NewClient` 前设置 `Config.ProxyURL`：

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // 或 http://、https://、socks4://

client, err := sshpass.NewClient(cfg)
```

底层辅助函数也已导出供高级使用：`Dial`、`NewConfig`、`LoadConfig`、
`LoadConfigOrPasswordFile`、`ParseSSHArgs`、`ParseSCPArgs`、`ParseRsyncArgs`、
`DetectCommandType`、`RunSCP`、`RunRsync`、`CleanRemotePath`、`SplitPaths`、
`ParseUserHostPath`、`ExitCodeFromError`。

## 编译

```bash
# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass

# 交叉编译
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=windows GOARCH=amd64               go build -o win-sshpass.exe ./cmd/sshpass
GOOS=darwin  GOARCH=arm64               go build -o win-sshpass ./cmd/sshpass
```

## 依赖

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3（仅 CLI 进度条）
- github.com/ncruces/zenity（仅 CLI 文件对话框）

