# win-sshpass 使用手冊

> Windows 平台的 sshpass 實作：密碼/金鑰 SSH 登入、互動式 Shell、SFTP 檔案傳輸、SCP/Rsync 風格傳輸，以及可重複使用的 Go SDK。

## 什麼是 win-sshpass？

win-sshpass 是 Linux sshpass 工具的 Windows 實作。它是一個獨立的可執行檔，無需安裝 OpenSSH 或其他軟體，下載即可使用。

它同時支援**命令列工具**和 **Go SDK** 兩種使用方式：

- **密碼/金鑰認證**：支援密碼、私鑰、環境變數、設定檔等多種認證方式。
- **互動式 Shell**：原始終端模式，支援 vim、top、htop 等全螢幕應用，動態終端調整大小。
- **SFTP 檔案傳輸**：上傳/下載檔案和目錄，帶進度條顯示。
- **SCP/Rsync 風格**：相容 scp 和 rsync 命令語法。
- **可重複使用 Go SDK**：作為 Go 函式庫匯入，嵌入到你自己的應用中，支援注入 I/O 串流和進度回呼。

## 30 秒快速體驗

```bash
# 透過 Scoop 安裝
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# 密碼登入並執行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 上傳檔案
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 開啟互動式 Shell
win-sshpass -p 'password' ssh user@host
```

## 核心功能

### 1. 多種認證方式

**密碼認證**：直接指定密碼、從檔案讀取、從環境變數讀取。

```bash
# 直接指定密碼
win-sshpass -p 'secret' ssh user@host

# 從檔案讀取
win-sshpass -f pass.txt ssh user@host

# 從環境變數讀取
SSHPASS='secret' win-sshpass -e ssh user@host
```

**私鑰認證**：支援 Ed25519、RSA 等金鑰格式。

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

**設定檔**：管理多台伺服器的連線資訊。

```bash
win-sshpass -f server.config -c 'docker ps'
```

### 2. 互動式 Shell（原始終端模式）

當不指定命令時，win-sshpass 會開啟互動式 Shell：

- **正確的回顯** — 輸入的字元正確顯示（無雙重回顯）
- **Ctrl+C / Ctrl+Z** — 訊號正確轉發到遠端程序
- **全螢幕應用** — vim、top、htop、nano 等正常工作
- **動態終端調整** — 遠端終端自動匹配本地視窗大小
- **Tab 補全** — 遠端 Shell 的 Tab 補全正常工作

```bash
win-sshpass -p 'password' ssh user@host
```

### 3. 檔案傳輸

**SFTP**：上傳/下載檔案和目錄，帶進度條顯示。

```bash
# 上傳檔案
win-sshpass -h host -p 'pass' -local file.txt -remote /tmp/file.txt

# 上傳多個檔案
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 下載目錄
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

**SCP 風格**：相容 scp 命令語法。

```bash
win-sshpass -p 'pass' scp ./app.jar user@server:/opt/app/
win-sshpass -p 'pass' scp -r ./dist user@server:/var/www/html
```

**Rsync 風格**：相容 rsync 命令語法。

```bash
win-sshpass -p 'pass' rsync -avz ./ user@server:/backup/
```

**Shell 內 rz/sz**：在互動式 Shell 中直接使用 rz/sz 命令。

```bash
# 在遠端 Shell 中：
rz                              # 上傳檔案（開啟檔案選擇器）
sz /remote/path/to/file         # 下載檔案
```

### 4. 可重複使用 Go SDK

win-sshpass 也是一個 Go 函式庫（`package sshpass`），可以嵌入到你自己的應用中：

```go
import sshpass "github.com/chuccp/win-sshpass"

cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
defer client.Close()

// 執行命令
client.Exec("uname -a")

// SFTP 傳輸
sftp, _ := client.SFTP()
sftp.Upload("./local.txt", "/tmp/remote.txt")
```

SDK **不包含任何 UI 程式碼**（無進度條、無檔案對話框），透過函式選項注入：

| 選項 | 用途 |
|------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | 重導向 I/O 串流 |
| `WithProgress(fn)` | 設定傳輸進度回呼 |
| `WithFileSelector(s)` | 設定 rz/sz 檔案選擇器 |
| `WithSignalHandler()` | 註冊 Ctrl+C 訊號處理器 |

## 參數一覽

| 參數 | 說明 | 範例 |
|------|------|------|
| `-p` | 密碼 | `-p 'secret123'` |
| `-i` | 私鑰路徑 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密碼檔案 / 設定檔 | `-f pass.txt` |
| `-e` | 從 SSHPASS 環境變數讀取密碼 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主機位址 | `-h example.com` |
| `-u` | 使用者名稱，預設 root | `-u ubuntu` |
| `-P` | 連接埠，預設 22 | `-P 2222` |
| `-c` | 要執行的命令 | `-c 'ls -la'` |
| `-local` | 本地路徑（逗號或空格分隔） | `-local "a.txt,b.txt"` |
| `-remote` | 遠端路徑 | `-remote /tmp/file.txt` |
| `-d` | 下載模式 | `-d` |
| `-k` | 啟用嚴格主機金鑰驗證 | `-k` |
| `-t` | 操作逾時（秒），0 表示無限制 | `-t 30` |
| `-ct` | TCP 連線逾時（秒），預設 10 | `-ct 5` |
| `-retry` | 連線重試次數（預設 3） | `-retry 5` |
| `-v` | 顯示版本 | `-v` |
| `-help` | 顯示說明 | `-help` |

## 快速連結

### 快速開始

- [安裝](getting-started/installation.md) - 下載與安裝方式
- [快速開始](getting-started/quick-start.md) - 第一個連線

### 使用者指南

- [SSH 連線](guide/ssh.md) - 密碼、金鑰、環境變數認證
- [檔案傳輸](guide/file-transfer.md) - SFTP 上傳下載
- [互動式 Shell](guide/shell.md) - 原始終端模式與 rz/sz
- [SCP 與 Rsync](guide/scp-rsync.md) - 相容 scp/rsync 語法
- [設定檔](guide/config-file.md) - 管理多台伺服器

### 進階與參考

- [Go SDK](advanced/sdk.md) - 作為 Go 函式庫使用
- [最佳實踐](advanced/best-practices.md) - 安全與效率建議
- [更新日誌](changelog.md)

## 社群

- [GitHub](https://github.com/chuccp/win-sshpass)
- [問題回報](https://github.com/chuccp/win-sshpass/issues)
