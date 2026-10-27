---
hide:
  - navigation
  - toc
---

# win-sshpass

> win-sshpass — 跨 Windows、Linux、macOS 的 SSH 用戶端：密碼/金鑰/agent 登入、SFTP 與 SCP/Rsync 傳輸、連接埠轉送、SOCKS/HTTP 代理、JSON 自動化、全螢幕 Web SSH 工作區、MCP 工具與可重複使用 Go SDK。

[快速開始](getting-started/quick-start.md){ .md-button .md-button--primary }
[安裝](getting-started/installation.md){ .md-button }
[原始碼 :simple-github:](https://github.com/chuccp/win-sshpass){ .md-button }

---

## :material-star: 功能特性

- **:material-console: SSH 與 Shell** — 密碼或私鑰認證、遠端命令執行、互動式 Shell 原始終端模式（正確回顯、Ctrl+C、vim/top 支援、動態視窗大小調整）。
- **:material-file-sync: 檔案傳輸** — SFTP 上傳/下載帶進度條、SCP 風格與 Rsync 風格傳輸語法、多檔案支援、遞迴目錄上傳。
- **:material-cloud-download: Shell 內 rz/sz** — 在互動式 Shell 中直接使用 `rz`/`sz` 命令 — 遠端伺服器無需安裝任何軟體；傳輸透過 SFTP 進行。
- **:material-shield-key: 金鑰產生** — 內建 SSH 金鑰對產生（Ed25519 和 RSA），無需 ssh-keygen。透過 SSH 部署公鑰實現免密碼登入。
- **:material-lan-connect: 代理通道** — 透過 SOCKS5（含認證）、SOCKS4、SOCKS4A、HTTP CONNECT 和 HTTPS CONNECT 代理通道化 SSH 連線。
- **:material-swap-horizontal: 連接埠轉送** — 本機（`-L`）和遠端（`-R`）TCP 連接埠轉送，透過 SSH 隧道傳輸，提供 `Forwarder` SDK API 供程式化使用。
- **:material-shield-account: SSH Agent 與 JSON 模式** — 自動檢測並使用本機 ssh-agent 進行認證，支援 agent 轉送（`-A`）。結構化 JSON 輸出（`-json`）適用於 AI 代理和自動化。
- **:material-reload: 斷點續傳** — 使用 `-resume` 標誌從中斷處恢復 SFTP 上傳/下載。
- **:material-fingerprint: 檔案雜湊與校驗** — 計算並校驗本地檔案校驗和（MD5、SHA-1、SHA-256、SHA-512）— 無需 SSH 連線。
- **:material-monitor-dashboard: Web SSH 工作區** — SFTP 檔案管理、拖曳傳輸、上傳進度、可拖曳分隔線、明暗主題與真實 xterm.js SSH PTY。
- **:material-robot: MCP** — 在 `/mcp` 提供 `ssh_exec` 與 SFTP 檔案工具，供 AI 用戶端使用。
- **:material-package-variant: 可重複使用 Go SDK** — 匯入 `package sshpass`，將 SSH/SFTP/Shell 嵌入你自己的應用，支援注入 I/O、進度回呼，零 UI 依賴。

---

## :material-rocket-launch: 30 秒快速開始

```bash
# 透過 WinGet 安裝
winget install chuccp.win-sshpass

# 或透過 Scoop 安裝
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# 使用密碼登入並執行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 上傳檔案
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 開啟互動式 Shell
win-sshpass -p 'password' ssh user@host

# 產生 SSH 金鑰對
win-sshpass keygen

# 開啟 Web SSH 工作區
win-sshpass -web
```

---

## :material-compass: 快速導覽

| | | |
|---|---|---|
| [:material-download: 安裝](getting-started/installation.md) | [:material-rocket-launch: 快速開始](getting-started/quick-start.md) | [:material-console: SSH 連線](guide/ssh.md) |
| [:material-file-sync: 檔案傳輸](guide/file-transfer.md) | [:material-monitor: 互動式 Shell](guide/shell.md) | [:material-folder-multiple: SCP 與 Rsync](guide/scp-rsync.md) |
| [:material-cog: 設定檔](guide/config-file.md) | [:material-code-braces: Go SDK](advanced/sdk.md) | [:material-security: 最佳實踐](advanced/best-practices.md) |
| [:material-monitor-dashboard: Web 控制台與 MCP](guide/web-console.md) | [:material-history: 更新日誌](changelog.md) | |

---

## :material-layers: 依賴

| 依賴 | 用途 |
|---|---|
| `golang.org/x/crypto/ssh` | SSH 協定實作 |
| `github.com/pkg/sftp` | SFTP 檔案傳輸 |
| `github.com/schollz/progressbar/v3` | CLI 進度條（僅 CLI） |
| `github.com/ncruces/zenity` | rz/sz 檔案對話框（僅 CLI，可選） |

win-sshpass 是一個**獨立的可執行檔** — 無需外部執行時依賴，下載即用。

---

## :simple-github: 社群

- [GitHub 倉庫](https://github.com/chuccp/win-sshpass)
- [問題回報](https://github.com/chuccp/win-sshpass/issues)
- [發布頁面](https://github.com/chuccp/win-sshpass/releases)
- [更新日誌](changelog.md)
