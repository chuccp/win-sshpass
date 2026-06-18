# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 sshpass 工具，實現類似 Linux sshpass 的功能。

## 功能特色

- 支援密碼或私鑰認證的 SSH 登入
- 執行遠端命令或開啟互動式 Shell
- 透過 SFTP 上傳/下載檔案（附進度條）
- SCP 風格和 Rsync 風格的檔案傳輸
- 設定檔支援，方便管理多台伺服器
- 互動式 Shell 使用 raw 終端模式（正確的回顯、Ctrl+C、vim/top 全螢幕程式支援）
- 互動式 Shell 模式下動態調整終端大小
- Git Bash 路徑轉換偵測與自動修復
- 支援 IPv6 位址
- 支援 x64 (amd64) 和 ARM64 架構

## 下載

從 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下載最新版本：

| 架構 | Zip | MSI 安裝包 |
|------|-----|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. 開啟 [Releases](https://github.com/chuccp/win-sshpass/releases) 頁面
2. 下載對應架構的 zip 或 MSI 安裝包（x64 或 ARM64）
3. 如果使用 MSI 安裝包：執行安裝程式即可，安裝目錄會自動加入系統 PATH 中

> **零依賴**：`win-sshpass.exe` 是一個獨立的可執行檔案，無需安裝 OpenSSH 或任何其他軟體。下載後放入 PATH 目錄即可直接使用。

## 快速開始

```bash
# 密碼登入執行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 私鑰登入執行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# 上傳檔案
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 下載檔案
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## 命令格式

### SSH 登入

```bash
# 密碼認證
win-sshpass -p <密碼> ssh [user@host] [命令]
win-sshpass -p <密碼> ssh -p <端口> user@host '命令'
win-sshpass -p <密碼> ssh -o StrictHostKeyChecking=no user@host

# 互動式 Shell（raw 終端模式：正確的回顯、Ctrl+C、vim/top 支援）
win-sshpass -p <密碼> ssh user@host

# 私鑰認證
win-sshpass -i <私鑰路徑> ssh [user@host] [命令]

# 環境變數密碼
SSHPASS=<密碼> win-sshpass -e ssh user@host

# 密碼檔案
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 設定檔（多行格式）
win-sshpass -f server.config
```

### 檔案傳輸

> **Git Bash 使用者**：遠端路徑需使用 `//` 前綴，例如 `-remote //tmp/file.txt`。詳見下方 [Git Bash 注意事項](#git-bash-注意事項)。

```bash
# 上傳檔案
win-sshpass -h <主機> -p <密碼> -local <本地路徑> -remote <遠端路徑>

# 上傳多個檔案（逗號分隔）
win-sshpass -h <主機> -p <密碼> -local "a.txt,b.txt,c.txt" -remote //tmp/

# 上傳多個檔案（空格分隔，僅適用於不含 / 或 \ 的簡單路徑）
win-sshpass -h <主機> -p <密碼> -local "a.txt b.txt c.txt" -remote //tmp/

# 上傳目錄（自動遞迴）
win-sshpass -h <主機> -p <密碼> -local <本地目錄> -remote <遠端目錄>

# 下載檔案/目錄
win-sshpass -h <主機> -p <密碼> -d -remote <遠端路徑> -local <本地路徑>
```

### SCP 風格

```bash
# 上傳檔案
win-sshpass -p <密碼> scp <本地檔案> user@host:<遠端路徑>
win-sshpass -p <密碼> scp -P <端口> <本地檔案> user@host:<遠端路徑>

# 上傳目錄
win-sshpass -p <密碼> scp -r <本地目錄> user@host:<遠端路徑>

# 下載檔案/目錄
win-sshpass -p <密碼> scp user@host:<遠端檔案> <本地路徑>
```

### Rsync 風格

```bash
# 上傳
win-sshpass -p <密碼> rsync -avz <本地路徑> user@host:<遠端路徑>

# 下載
win-sshpass -p <密碼> rsync -avz user@host:<遠端路徑> <本地路徑>
```

## 參數說明

| 參數 | 說明 | 範例 |
|------|------|------|
| `-p` | 密碼 | `-p 'secret123'` |
| `-i` | 私鑰路徑 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密碼檔案/設定檔 | `-f pass.txt` |
| `-e` | 從環境變數 SSHPASS 讀密碼 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主機位址 | `-h example.com` |
| `-u` | 使用者名稱，預設 root | `-u ubuntu` |
| `-P` | 端口，預設 22 | `-P 2222` |
| `-c` | 執行的命令 | `-c 'ls -la'` |
| `-local` | 本地路徑（逗號或空格分隔） | `-local "a.txt,b.txt"` |
| `-remote` | 遠端路徑（上傳/下載） | `-remote /tmp/file.txt` |
| `-d` | 下載模式 | `-d` |
| `-k` | 啟用嚴格主機金鑰驗證 | `-k` |
| `-t` | 總操作逾時時間（秒），0 表示不限 | `-t 30` |
| `-ct` | TCP 連線逾時時間（秒），預設 10 | `-ct 5` |
| `-v` | 顯示版本 | `-v` |
| `-help` | 顯示帮助資訊 | `-help` |

## 設定檔格式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # 可選，使用私鑰代替密碼
# timeout: 0              # 可選，總操作逾時時間（秒），0 表示不限
# connect_timeout: 10     # 可選，TCP 連線逾時時間（秒）
# strict_host_key: false  # 可選，啟用嚴格主機金鑰驗證
```

使用方式：
```bash
win-sshpass -f server.config -c 'ls -la'
win-sshpass -f server.config 'ls -la'
```

## 完整範例

```bash
# 1. 密碼登入執行命令
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 私鑰登入執行 sudo 命令
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. 上傳整個目錄到伺服器
win-sshpass -h server.com -p 'mypass' -local ./dist -remote //var/www/html

# 4. 下載伺服器日誌目錄
win-sshpass -h server.com -p 'mypass' -d -remote //var/log/nginx -local ./logs

# 5. SCP 上傳檔案
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 環境變數傳遞密碼（更安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'

# 7. 操作逾時（30 秒後自動中斷）
win-sshpass -p 'mypass' -t 30 ssh user@server.com 'long-running-command'

# 8. 設定檔 + 位置參數命令
win-sshpass -f server.config 'docker ps'
```

## Git Bash 注意事項

遠端路徑用 `//` 開頭避免路徑轉換：
```bash
# 錯誤：/tmp 會被轉換為 Windows 路徑
win-sshpass ... -remote /tmp/file.txt

# 正確：使用雙斜線
win-sshpass ... -remote //tmp/file.txt
```

## 編譯

```bash
go build -o win-sshpass.exe .
```

## 相依套件

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3

## 相關專案

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — 一個 Go Web 框架，整合依賴注入、泛型資料存取和開箱即用的生產工具，內建 HTTP 路由、快取、日誌、Redis 等。