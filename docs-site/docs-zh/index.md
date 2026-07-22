---
hide:
  - navigation
  - toc
---

# win-sshpass

> win-sshpass — 跨平台 sshpass 实现：密码/密钥 SSH 登录、交互式 Shell、SFTP/SCP/Rsync 文件传输、SOCKS5/SOCKS4/HTTP 代理隧道、断点续传、文件哈希校验、密钥生成，以及可复用的 Go SDK。

[快速开始](getting-started/quick-start.md){ .md-button .md-button--primary }
[安装](getting-started/installation.md){ .md-button }
[源代码 :simple-github:](https://github.com/chuccp/win-sshpass){ .md-button }

---

## :material-star: 功能特性

- **:material-console: SSH 与 Shell** — 密码或私钥认证、远程命令执行、交互式 Shell 原始终端模式（正确回显、Ctrl+C、vim/top 支持、动态窗口大小调整）。
- **:material-file-sync: 文件传输** — SFTP 上传/下载带进度条、SCP 风格与 Rsync 风格传输语法、多文件支持、递归目录上传。
- **:material-cloud-download: Shell 内 rz/sz** — 在交互式 Shell 中直接使用 `rz`/`sz` 命令 — 远程服务器无需安装任何软件；传输通过 SFTP 进行。
- **:material-shield-key: 密钥生成** — 内置 SSH 密钥对生成（Ed25519 和 RSA），无需 ssh-keygen。通过 SSH 部署公钥实现免密登录。
- **:material-lan-connect: 代理隧道** — 通过 SOCKS5（含认证）、SOCKS4、SOCKS4A、HTTP CONNECT 和 HTTPS CONNECT 代理隧道化 SSH 连接。
- **:material-reload: 断点续传** — 使用 `-resume` 标志从中断处恢复 SFTP 上传/下载。
- **:material-fingerprint: 文件哈希与校验** — 计算并校验本地文件校验和（MD5、SHA-1、SHA-256、SHA-512）— 无需 SSH 连接。
- **:material-package-variant: 可复用 Go SDK** — 导入 `package sshpass`，将 SSH/SFTP/Shell 嵌入你自己的应用，支持注入 I/O、进度回调，零 UI 依赖。

---

## :material-rocket-launch: 30 秒快速开始

```bash
# 通过 WinGet 安装
winget install chuccp.win-sshpass

# 或通过 Scoop 安装
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# 使用密码登录并执行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 上传文件
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 打开交互式 Shell
win-sshpass -p 'password' ssh user@host

# 生成 SSH 密钥对
win-sshpass keygen
```

---

## :material-compass: 快速导航

| | | |
|---|---|---|
| [:material-download: 安装](getting-started/installation.md) | [:material-rocket-launch: 快速开始](getting-started/quick-start.md) | [:material-console: SSH 连接](guide/ssh.md) |
| [:material-file-sync: 文件传输](guide/file-transfer.md) | [:material-monitor: 交互式 Shell](guide/shell.md) | [:material-folder-multiple: SCP 与 Rsync](guide/scp-rsync.md) |
| [:material-cog: 配置文件](guide/config-file.md) | [:material-code-braces: Go SDK](advanced/sdk.md) | [:material-security: 最佳实践](advanced/best-practices.md) |
| [:material-history: 更新日志](changelog.md) | | |

---

## :material-layers: 依赖

| 依赖 | 用途 |
|---|---|
| `golang.org/x/crypto/ssh` | SSH 协议实现 |
| `github.com/pkg/sftp` | SFTP 文件传输 |
| `github.com/schollz/progressbar/v3` | CLI 进度条（仅 CLI） |
| `github.com/ncruces/zenity` | rz/sz 文件对话框（仅 CLI，可选） |

win-sshpass 是一个**独立的可执行文件** — 无需外部运行时依赖，下载即用。

---

## :simple-github: 社区

- [GitHub 仓库](https://github.com/chuccp/win-sshpass)
- [问题反馈](https://github.com/chuccp/win-sshpass/issues)
- [发布页面](https://github.com/chuccp/win-sshpass/releases)
- [更新日志](changelog.md)
