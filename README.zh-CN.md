# win-sshpass

[![Release](https://badgen.net/github/release/chuccp/win-sshpass)](https://github.com/chuccp/win-sshpass/releases)
[![License](https://badgen.net/badge/License/Apache%202.0/blue)](LICENSE)
[![Go Version](https://badgen.net/badge/Go/1.26.0/00ADD8?icon=go)](go.mod)
[![Downloads](https://img.shields.io/github/downloads/chuccp/win-sshpass/total)](https://github.com/chuccp/win-sshpass/releases)

[English](README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md)

`win-sshpass` 是一个跨 Windows、Linux 和 macOS 的 SSH 客户端与自动化工具。
支持密码/密钥登录、SFTP 与 SCP/Rsync 风格传输、端口转发、浏览器 SSH
工作区、MCP 工具和可复用 Go SDK。

## 主要功能

- 密码、私钥、ssh-agent 和显式空密码登录。
- 远程命令、JSON、重试、超时、代理与 `known_hosts`。
- 交互式终端、后台命令 `--bg`、`rz`/`sz` 文件传输。
- SFTP 上传下载、目录递归、多文件和断点续传。
- 本地/远程端口转发，SOCKS4/SOCKS5/HTTP/HTTPS 代理。
- 全屏 Web SSH：SFTP 文件管理器与 xterm.js 真实 PTY。
- 拖拽传输、上传百分比、可拖动分隔线、明暗主题。
- SSH 命令与 SFTP 操作的 MCP 工具。
- 自更新、SSH 密钥生成、Hash 与校验。

## 安装

```powershell
winget install chuccp.win-sshpass
```

```powershell
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

也可以从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases)
下载压缩包、MSI 或 PKG。程序无外部运行时依赖。

## 快速开始

```bash
# 执行远程命令
win-sshpass -p '密码' ssh root@example.com 'whoami'

# 打开交互式 Shell
win-sshpass -p '密码' ssh root@example.com

# 上传和下载
win-sshpass -h example.com -p '密码' -local ./app.jar -remote /tmp/app.jar
win-sshpass -h example.com -p '密码' -d -remote /tmp/app.log -local ./app.log

# Web SSH 控制台，默认 http://127.0.0.1:6326
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p '密码'
```

命令必须通过 `-h` 或 `-f` 提供 SSH 服务器地址。打开
<http://127.0.0.1:6326> 会自动连接并直接进入工作区。

## Web SSH 工作区

- 左侧：SFTP 目录浏览、上传下载、新建目录、删除、拖拽传输。
- 右侧：xterm.js 渲染的真实 SSH PTY，支持 vim/top 和 Tab 补全。
- 分隔线：拖动或方向键调整宽度，自动记忆。
- 主题：默认跟随系统，可切换浅色/深色。
- 上传进度：文件名、文件序号、进度条和总体百分比。

裸执行 `win-sshpass -web` 会报错退出。必须在命令中提供服务器地址，并按需
提供用户和认证参数，也可以使用配置文件：

```bash
win-sshpass -f server.config -web
```

更改监听地址：

```bash
win-sshpass -web -h example.com -u root -p '密码' -web-addr 127.0.0.1:6326
```

默认仅监听回环地址。

## MCP

Web 服务默认启用 MCP Streamable HTTP：

```text
http://127.0.0.1:6326/mcp
```

Codex 配置示例：

```toml
[mcp_servers.win-sshpass]
url = "http://127.0.0.1:6326/mcp"
```

工具：

- `ssh_exec`
- `sftp_list`
- `sftp_read`
- `sftp_write`
- `sftp_mkdir`
- `sftp_delete`

MCP 使用 `-web` 命令提供的 SSH 连接，并通过 `Mcp-Session-Id` 维护会话。

## 常用命令

| 操作 | 命令 |
|---|---|
| 私钥登录 | `win-sshpass -i ~/.ssh/id_ed25519 ssh user@host` |
| 环境变量密码 | `SSHPASS='pass' win-sshpass -e ssh user@host` |
| 密码文件 | `win-sshpass -f pass.txt ssh user@host` |
| 配置文件 | `win-sshpass -f server.config 'docker ps'` |
| 指定端口 | `win-sshpass -p 'pass' ssh -p 2222 user@host` |
| 操作超时 | `win-sshpass -t 30 -p 'pass' ssh user@host 'long-task'` |
| SOCKS5 代理 | `win-sshpass -proxy socks5://127.0.0.1:1080 ssh user@host` |
| SCP 上传 | `win-sshpass -p 'pass' scp -P 2222 file.txt user@host:/tmp/` |
| Rsync 上传 | `win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/` |
| 生成密钥 | `win-sshpass keygen -algo ed25519` |
| Hash / 校验 | `win-sshpass hash sha256 file.iso` |
| 检查更新 | `win-sshpass update -check` |

## 配置文件

```yaml
host: 192.168.1.10
username: root
password: secret
port: 22
timeout: 0
strict_host_key: false
```

```bash
win-sshpass -f server.config 'docker ps'
```

完整配置说明见[配置文件文档](https://chuccp.github.io/win-sshpass/zh/guide/config-file/)。

## Go SDK

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

if err := client.Exec("uname -a"); err != nil {
    log.Fatal(err)
}
```

```bash
go get github.com/chuccp/win-sshpass
```

详见 [Go SDK 文档](https://chuccp.github.io/win-sshpass/zh/advanced/sdk/)。

## 在线文档

- [English](https://chuccp.github.io/win-sshpass/en/)
- [简体中文](https://chuccp.github.io/win-sshpass/zh/)
- [繁體中文](https://chuccp.github.io/win-sshpass/tw/)
- [日本語](https://chuccp.github.io/win-sshpass/ja/)
- [文档源码](https://github.com/chuccp/win-sshpass/tree/main/docs-site)

## 开发

```bash
go test -race ./...
go vet ./...
go build -o win-sshpass.exe ./cmd/sshpass
```

交叉编译：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
```

## License

Apache-2.0，详见 [LICENSE](LICENSE)。
