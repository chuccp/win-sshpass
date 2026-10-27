# win-sshpass

[![Release](https://badgen.net/github/release/chuccp/win-sshpass)](https://github.com/chuccp/win-sshpass/releases)
[![License](https://badgen.net/badge/License/Apache%202.0/blue)](LICENSE)
[![Go Version](https://badgen.net/badge/Go/1.26.0/00ADD8?icon=go)](go.mod)
[![Downloads](https://img.shields.io/github/downloads/chuccp/win-sshpass/total)](https://github.com/chuccp/win-sshpass/releases)

[English](README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md)

`win-sshpass` 是跨 Windows、Linux 與 macOS 的 SSH 用戶端與自動化工具。
支援密碼/金鑰登入、SFTP 與 SCP/Rsync 風格傳輸、連接埠轉送、瀏覽器 SSH
工作區、MCP 工具與可重用 Go SDK。

## 主要功能

- 密碼、私鑰、ssh-agent 與顯式空密碼登入。
- 遠端命令、JSON、重試、逾時、代理與 `known_hosts`。
- 互動式終端、背景命令 `--bg`、`rz`/`sz` 檔案傳輸。
- SFTP 上傳下載、目錄遞迴、多檔案與斷點續傳。
- 本機/遠端連接埠轉送，SOCKS4/SOCKS5/HTTP/HTTPS 代理。
- 全螢幕 Web SSH：SFTP 檔案管理員與 xterm.js 真實 PTY。
- 拖曳傳輸、上傳百分比、可拖曳分隔線、明暗主題。
- SSH 命令與 SFTP 操作的 MCP 工具。
- 自更新、SSH 金鑰產生、Hash 與驗證。

## 安裝

```powershell
winget install chuccp.win-sshpass
```

```powershell
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

也可從 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases)
下載壓縮檔、MSI 或 PKG。程式沒有外部執行階段依賴。

## 快速開始

```bash
# 執行遠端命令
win-sshpass -p '密碼' ssh root@example.com 'whoami'

# 開啟互動式 Shell
win-sshpass -p '密碼' ssh root@example.com

# 上傳與下載
win-sshpass -h example.com -p '密碼' -local ./app.jar -remote /tmp/app.jar
win-sshpass -h example.com -p '密碼' -d -remote /tmp/app.log -local ./app.log

# Web SSH 控制台，預設 http://127.0.0.1:6326
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p '密碼'
```

命令必須透過 `-h` 或 `-f` 提供 SSH 伺服器位址。開啟
<http://127.0.0.1:6326> 會自動連線並直接進入工作區。

## Web SSH 工作區

- 左側：SFTP 目錄瀏覽、上傳下載、建立目錄、刪除、拖曳傳輸。
- 右側：xterm.js 呈現的真實 SSH PTY，支援 vim/top 與 Tab 補全。
- 分隔線：拖曳或方向鍵調整寬度，自動記憶。
- 主題：預設跟隨系統，可切換淺色/深色。
- 上傳進度：檔名、序號、進度條與總體百分比。

直接執行 `win-sshpass -web` 會報錯結束。必須在命令中提供伺服器位址，並視需要
提供使用者與驗證參數，也可以使用設定檔：

```bash
win-sshpass -f server.config -web
```

變更監聽位址：

```bash
win-sshpass -web -h example.com -u root -p '密碼' -web-addr 127.0.0.1:6326
```

預設僅監聽回環位址。

## MCP

Web 服務預設啟用 MCP Streamable HTTP：

```text
http://127.0.0.1:6326/mcp
```

Codex 設定範例：

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

MCP 使用 `-web` 命令提供的 SSH 連線，並透過 `Mcp-Session-Id` 維護工作階段。

## 常用命令

| 操作 | 命令 |
|---|---|
| 私鑰登入 | `win-sshpass -i ~/.ssh/id_ed25519 ssh user@host` |
| 環境變數密碼 | `SSHPASS='pass' win-sshpass -e ssh user@host` |
| 密碼檔 | `win-sshpass -f pass.txt ssh user@host` |
| 設定檔 | `win-sshpass -f server.config 'docker ps'` |
| 指定連接埠 | `win-sshpass -p 'pass' ssh -p 2222 user@host` |
| 操作逾時 | `win-sshpass -t 30 -p 'pass' ssh user@host 'long-task'` |
| SOCKS5 代理 | `win-sshpass -proxy socks5://127.0.0.1:1080 ssh user@host` |
| SCP 上傳 | `win-sshpass -p 'pass' scp -P 2222 file.txt user@host:/tmp/` |
| Rsync 上傳 | `win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/` |
| 產生金鑰 | `win-sshpass keygen -algo ed25519` |
| Hash / 驗證 | `win-sshpass hash sha256 file.iso` |
| 檢查更新 | `win-sshpass update -check` |

## 設定檔

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

完整說明見[設定檔文件](https://chuccp.github.io/win-sshpass/tw/guide/config-file/)。

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

詳見 [Go SDK 文件](https://chuccp.github.io/win-sshpass/tw/advanced/sdk/)。

## 線上文件

- [English](https://chuccp.github.io/win-sshpass/en/)
- [简体中文](https://chuccp.github.io/win-sshpass/zh/)
- [繁體中文](https://chuccp.github.io/win-sshpass/tw/)
- [日本語](https://chuccp.github.io/win-sshpass/ja/)
- [文件原始碼](https://github.com/chuccp/win-sshpass/tree/main/docs-site)

## 開發

```bash
go test -race ./...
go vet ./...
go build -o win-sshpass.exe ./cmd/sshpass
```

交叉編譯：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
```

## License

Apache-2.0，詳見 [LICENSE](LICENSE)。
