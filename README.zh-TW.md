## 相關專案

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — 輕鬆解決鑑權問題——路由宣告需要什麼權限，Filter 一處校驗，handler 裡乾乾淨淨。泛型 Model 解決全棧 CRUD——定義好 struct，增刪改查直接能用。輕巧，需要的組件按需安裝。不需要代碼生成，無需CI工具，目前最精巧的go web全棧框架。

# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

跨平台 sshpass 工具（Windows、Linux 和 macOS），實現類似 Linux sshpass 的功能。

> 💡 **如果這個專案對你有幫助，歡迎點個 ⭐ Star！** 讓更多人發現這個工具。

## 功能特色

- 支援密碼、私鑰或 ssh-agent 認證的 SSH 登入
- 執行遠端命令或開啟互動式 Shell
- 透過 SFTP 上傳/下載檔案（附進度條）
- SCP 風格和 Rsync 風格的檔案傳輸
- 設定檔支援，方便管理多台伺服器
- 互動式 Shell 使用 raw 終端模式（正確的回顯、Ctrl+C、vim/top 全螢幕程式支援）
- 互動式 Shell 模式下動態調整終端大小
- Git Bash 路徑轉換偵測與自動修復
- 支援 IPv6 位址
- 支援 Windows（x64、ARM64）、Linux（amd64、arm64）和 macOS（amd64、arm64）
- **可複用 Go SDK** — 作為函式庫引入（`package sshpass`），在自有應用中嵌入 SSH/SFTP/Shell 能力，支援注入 I/O 串流與進度回呼
- **連接埠轉送** — 本機（`-L`）和遠端（`-R`）TCP 連接埠轉送，透過 SSH 隧道傳輸
- **SSH Agent 轉送**（`-A`）— 轉送本機 ssh-agent 到遠端伺服器，無憑據時自動檢測 agent
- **JSON 輸出模式**（`-json`）— 為 AI 代理和自動化輸出結構化 JSON 結果
- **代理支援** — 透過 SOCKS5/SOCKS4/HTTP/HTTPS 代理通道連線 SSH
- **斷點續傳** — 中斷的 SFTP 檔案傳輸可從斷點處恢復
- **檔案雜湊與校驗** — 計算和校驗本地檔案雜湊（MD5、SHA-1、SHA-256、SHA-512）
- **金鑰產生** — SSH 金鑰對產生（Ed25519/RSA）
- **自動更新**（`update`）— 檢查 GitHub Releases 並就地替換為最新版本的執行檔
- **背景執行指令**（`--bg`）— 透過 SSH 啟動服務後立即返回，工作階段不會卡在服務上

## 下載

從 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下載最新版本：

### Windows

| 架構 | Zip | MSI 安裝包 |
|------|-----|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

### Linux

| 架構 | Tarball |
|------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

### macOS

| 架構 | PKG 安裝包 | Tarball |
|------|-----------|---------|
| **amd64 (Intel)** | `win-sshpass-*-darwin-amd64.pkg` | `win-sshpass-*-darwin-amd64.tar.gz` |
| **arm64 (Apple Silicon)** | `win-sshpass-*-darwin-arm64.pkg` | `win-sshpass-*-darwin-arm64.tar.gz` |

> `.pkg` 安裝包會自動將二進位檔案安裝到 `/usr/local/bin/win-sshpass`。

1. 開啟 [Releases](https://github.com/chuccp/win-sshpass/releases) 頁面
2. 下載對應平台和架構的安裝包
3. **Windows MSI / macOS PKG**：執行安裝程式即可，二進位檔案會自動加入系統 PATH
4. **Windows Zip / Linux tar.gz / macOS tar.gz**：解壓後將二進位檔案放入 PATH 目錄

> **零依賴**：`win-sshpass.exe` 是一個獨立的可執行檔案，無需安裝 OpenSSH 或任何其他軟體。下載後放入 PATH 目錄即可直接使用。

### 透過 Scoop 安裝

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 透過 WinGet 安裝

```bash
winget install chuccp.win-sshpass
```

## 快速開始

```bash
# 密碼登入執行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 私鑰登入執行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# SSH agent 認證（自動偵測，無需 -p/-i）
win-sshpass ssh user@example.com 'whoami'

# JSON 輸出模式（適用於 AI 代理與自動化）
win-sshpass -json -p 'password' ssh user@example.com 'uptime'

# 上傳檔案
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 下載檔案
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## 互動式 Shell

不指定命令時，`win-sshpass` 會開啟一個 **raw 終端模式** 的互動式 Shell：

```bash
win-sshpass -p 'password' ssh user@host
```

**Raw 終端模式** 特性：

- **正確的回顯** — 輸入的字元正確顯示（不會出現雙重回顯）
- **Ctrl+C / Ctrl+Z** — 訊號正確轉發到遠端程序
- **全螢幕程式** — vim、top、htop、nano 等全螢幕應用正常運作
- **動態終端大小調整** — 遠端終端自動匹配本地視窗大小
- **Tab 補全** — 遠端 Shell 的 Tab 補全功能正常運作

### 互動式 Shell 中的檔案傳輸

連線狀態下，使用 `rz` / `sz` 命令傳輸檔案（遠端伺服器無需安裝任何軟體）：

```bash
# 上傳檔案到遠端目前目錄（彈出檔案選擇器）
rz

# 上傳指定本地檔案
rz /本機/檔案/路徑

# 下載遠端檔案（彈出儲存對話框）
sz /遠端/檔案/路徑

# 下載遠端檔案到指定本機路徑
sz /遠端/檔案/路徑 /本機/儲存/路徑
```

> **原理**：當遠端 Shell 回報 `rz`/`sz: command not found` 時，工具自動攔截並透過 SFTP 完成傳輸。支援檔案和目錄，附進度條。

## 命令格式

### SSH 登入

```bash
# 密碼認證
win-sshpass -p <密碼> ssh [user@host] [命令]
win-sshpass -p <密碼> ssh -p <端口> user@host '命令'
win-sshpass -p <密碼> ssh -o StrictHostKeyChecking=no user@host

# 主機名稱可不帶 user@（使用者名稱預設 root，主機名稱交由系統解析）
win-sshpass -p <密碼> ssh example.com 'uptime'

# 空密碼（伺服器未設定密碼時使用；必須明確寫 -p ''，不能省略）
win-sshpass -p '' ssh root@192.168.1.100 'hostname'

# 互動式 Shell（raw 終端模式：正確的回顯、Ctrl+C、vim/top 支援）
win-sshpass -p <密碼> ssh user@host

# 私鑰認證
win-sshpass -i <私鑰路徑> ssh [user@host] [命令]

# SSH agent 認證（自動偵測，無需 -p/-i）
win-sshpass ssh user@host 'whoami'

# SSH agent 轉送（-A 參數）
win-sshpass -A -i ~/.ssh/id_ed25519 ssh user@jumphost

# 環境變數密碼
SSHPASS=<密碼> win-sshpass -e ssh user@host

# 密碼檔案
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 設定檔（多行格式）
win-sshpass -f server.config
```

### 背景執行指令（`--bg`）

啟動常駐程式（服務、守護程式）後立即返回，不等待它結束。

```bash
# 部署完新二進位檔後啟動服務
win-sshpass -p 'pass' ssh --bg root@host 'cd /app && ./myapp > /tmp/myapp.log 2>&1'

# 同一條指令裡替換二進位檔並啟動
win-sshpass -p 'pass' ssh --bg root@host 'cd /app && cp /tmp/myapp-new ./myapp && chmod +x ./myapp && ./myapp > /tmp/myapp.log 2>&1'

# 重啟服務，再用另一條連線驗證
win-sshpass -p 'pass' ssh --bg root@host 'pkill -x myapp; sleep 1; cd /app && ./myapp > /tmp/myapp.log 2>&1'
win-sshpass -p 'pass' ssh root@host 'ps -ef | grep [m]yapp; curl -s localhost:8080/api/status/ping'
```

為什麼需要它：`ssh host 'nohup ./myapp &'` 這種寫法**會卡住** —— 背景程式繼承了工作階段的 stdout/stderr，SSH 通道就一直不 EOF，客戶端會一直等到服務結束（或 `-t` 逾時）。`--bg` 會把指令透過 `setsid` 啟動（沒有 setsid 的系統自動退化為 `nohup`），並將三個標準流全部重新導向，因此 shell 一返回通道就關閉，而服務繼續執行。

注意：

- 輸出**必須寫在指令內部**重新導向（`> /tmp/xxx.log 2>&1`），否則會被 `--bg` 丟棄
- 結束碼只代表指令已啟動，不代表服務起來了 —— 請像上面那樣用另一條連線驗證
- 也可以在[設定檔](#設定檔格式)裡按主機設定 `background: true`

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
| `-p` | 密碼。明確傳入空值 `-p ''` 表示伺服器沒有設定密碼 | `-p 'secret123'` |
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
| `-retry` | 總連線嘗試次數（預設：3） | `-retry 5` |
| `-resume` | 從斷點恢復中斷的檔案傳輸 | `-resume` |
| `-proxy` | 代理 URL（socks5/socks4/http/https） | `-proxy socks5://127.0.0.1:1080` |
| `-L` | 本機連接埠轉送（可重複使用） | `-L 8080:db.internal:3306` |
| `-R` | 遠端連接埠轉送（可重複使用） | `-R 9090:localhost:8080` |
| `-A` | 啟用 ssh-agent 轉送 | `-A` |
| `--bg` | 讓指令脫離工作階段在背景啟動並立即返回（用於啟動服務，如部署完二進位檔後）。輸出需在指令內重新導向 | `--bg './server > /tmp/server.log 2>&1'` |
| `-json` | JSON 格式輸出（適用於 AI/自動化） | `-json` |
| `-algo` | 金鑰演算法（ed25519/rsa），預設 ed25519 | `-algo rsa` |
| `-out` | 金鑰輸出路徑前綴，預設 id_ed25519 | `-out ~/.ssh/mykey` |
| `-comment` | 金鑰註解 | `-comment "my-laptop"` |
| `-v` | 顯示版本 | `-v` |
| `-help` | 顯示帮助資訊 | `-help` |

## 雜湊與校驗

無需 SSH 連線，即可計算和校驗本地檔案雜湊：

```bash
# 計算雜湊
win-sshpass hash md5 ./file.iso
win-sshpass hash sha256 ./file.iso

# 校驗檔案
win-sshpass verify sha256 d1dc38f6df... ./file.iso
# 輸出: OK  (或: FAILED)
```

支援的演算法：`md5`、`sha1`、`sha256`、`sha512`。

## 金鑰產生

無需 SSH 連線，即可產生 SSH 金鑰對：

```bash
win-sshpass keygen                  # 產生 Ed25519 金鑰（預設）
win-sshpass keygen -algo rsa        # 產生 RSA 金鑰
win-sshpass keygen -out ~/.ssh/mykey   # 指定輸出路徑
win-sshpass keygen -comment "my-laptop" # 加入註解
```

支援的演算法：`ed25519`、`rsa`。

產生的檔案：
- `<名稱>` — 私鑰
- `<名稱>.pub` — 公鑰

預設輸出：`id_ed25519` 和 `id_ed25519.pub`（或 `id_rsa` 和 `id_rsa.pub`）。

**部署公鑰以實現無密碼登入：**

```bash
# 將公鑰內容讀入變數，再透過 SSH 部署
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'password' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

# 然後使用私鑰登入
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

## 自動更新

直接從 GitHub 更新到最新版本，無需手動下載壓縮檔：

```bash
# 更新到最新版本
win-sshpass update

# 僅檢查是否有新版本
win-sshpass update -check

# 重新安裝目前版本，或切換到指定版本
win-sshpass update -force
win-sshpass update -version v1.0.0
```

此命令會先向 GitHub 查詢目前的最新版本號，與正在執行的版本比較——只需一次請求，不會下載任何檔案。只有版本確實較新時，才會下載符合目前系統與架構的安裝包，並就地替換正在執行的執行檔：

```
$ win-sshpass update
Updating win-sshpass v0.9.1 (windows/amd64)...
downloading win-sshpass-v0.9.4-amd64.zip (3.4 MiB)
Downloading win-sshpass-v0.9.4-amd64.zip 100% |████████████████| (3.6/3.6 MB, 1.2 MB/s)
Updated win-sshpass v0.9.1 -> v0.9.4 (C:\Tools\win-sshpass.exe)
```

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `-check` | 僅檢查是否有新版本，不下載 | — |
| `-force` | 即使已是最新版本也重新安裝 | — |
| `-version <tag>` | 安裝指定的 release 版本（隱含 `-force`） | 最新版本 |
| `-target <path>` | 要替換的執行檔路徑 | 目前執行的程式 |

注意事項：

- Windows 上舊版本會保留為 `win-sshpass.exe.old`（執行中的 `.exe` 無法刪除），下次更新時自動清理。Linux 與 macOS 則是原子重命名替換。
- 安裝目錄必須可寫入。如果 win-sshpass 是透過 **scoop**、**winget** 或 **MSI/PKG** 安裝的，請改用對應的套件管理器更新，以免檔案歸屬混亂。
- 替換之前會校驗下載內容確實是執行檔，因此下載失敗或被攔截不會導致無法使用的 `win-sshpass`。
- 版本檢查走的是 github.com 的 `/releases/latest` 重導向，而非 REST API，因此無需 token、不受 API 每小時 60 次的匿名限制影響，在 `api.github.com` 被封鎖的網路下同樣可用（API 僅作為後備）。
- 使用代理時，`update` 會讀取 `HTTPS_PROXY` / `HTTP_PROXY` 環境變數。

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
# proxy: socks5://user:pass@127.0.0.1:1080  # 可選，代理 URL（socks5/socks4/http/https）
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

# 9. 斷點續傳上傳
win-sshpass -p 'mypass' -h server.com -local ./bigfile.iso -remote //data/bigfile.iso -resume

# 10. 計算檔案雜湊
win-sshpass hash sha256 ./download.iso

# 11. 校驗檔案完整性
win-sshpass verify sha256 d1dc38f6dfb1e4c8... ./download.iso

# 12. 產生 Ed25519 金鑰對
win-sshpass keygen -out ~/.ssh/my_key -comment "my-work-laptop"

# 13. 產生 RSA 金鑰對
win-sshpass keygen -algo rsa -out ~/.ssh/my_rsa_key

# 14. SSH agent 認證（無需密碼或金鑰）
win-sshpass ssh user@host 'whoami'

# 15. SSH agent 轉送到跳板機
win-sshpass -A ssh user@jumphost

# 16. JSON 輸出模式（適用於自動化）
win-sshpass -json -p 'pass' ssh user@host 'uptime'

# 17. 本機連接埠轉送（透過跳板機存取內部資料庫）
win-sshpass -p 'pass' -L 3306:db.internal:3306 ssh user@jumphost

# 18. 遠端連接埠轉送（暴露本機開發伺服器）
win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server
```

## 代理支援

透過代理伺服器建立 SSH 通道連線。支援協定：SOCKS5、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

```bash
# SOCKS5 代理
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@host

# SOCKS5 帶認證
win-sshpass -p 'pass' -proxy socks5://proxyuser:proxypass@127.0.0.1:1080 ssh user@host

# SOCKS4 代理
win-sshpass -p 'pass' -proxy socks4://192.168.1.1:1080 ssh user@host

# HTTP CONNECT 代理
win-sshpass -p 'pass' -proxy http://proxy.local:8080 ssh user@host

# HTTPS CONNECT 代理（帶認證）
win-sshpass -p 'pass' -proxy https://user:pass@proxy.local:8443 ssh user@host

# 代理 + 檔案傳輸
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 -h host -local ./file.txt -remote /tmp/file.txt

# 代理 + SCP
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 scp ./app.jar user@host:/opt/app/

# 設定檔中設定代理
# proxy: socks5://user:pass@127.0.0.1:1080
```

## 連接埠轉送

透過 SSH 伺服器建立 TCP 隧道連線，支援本機轉送（`-L`）和遠端轉送（`-R`）：

```bash
# 本機轉送：透過跳板機存取 db.internal:3306 → localhost:8080
win-sshpass -p 'pass' -L 8080:db.internal:3306 ssh user@jumphost

# 多個本機轉送
win-sshpass -p 'pass' -L 8080:db1.internal:3306 -L 8081:db2.internal:3306 ssh user@jumphost

# 遠端轉送：將 localhost:8080 暴露在遠端伺服器的 9090 連接埠
win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server

# 僅轉送模式（無命令，阻塞直到 Ctrl+C）
win-sshpass -p 'pass' -L 8080:db.internal:3306 -L 6379:redis.internal:6379 ssh user@jumphost
```

> 連接埠轉送使用標準 OpenSSH 格式：`[繫結位址:]連接埠:主機:主機連接埠`。不支援與 SCP、Rsync 或檔案傳輸模式同時使用。

## Git Bash 注意事項

遠端路徑用 `//` 開頭避免路徑轉換：
```bash
# 錯誤：/tmp 會被轉換為 Windows 路徑
win-sshpass ... -remote /tmp/file.txt

# 正確：使用雙斜線
win-sshpass ... -remote //tmp/file.txt
```

## 作為 Go SDK 使用

`win-sshpass` 也是一個可複用的 Go 函式庫（`package sshpass`）。引入它即可在自有應用中嵌入 SSH/SFTP/Shell 能力：

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

	// NewClient 撥號並回傳一個即開即用的客戶端。
	client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 執行命令（輸出預設流向 os.Stdout/os.Stderr）。
	if err := client.Exec("uname -a"); err != nil {
		log.Fatal(err)
	}

	// 透過 SFTP 上傳檔案。
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

### 自訂選項

透過傳入 `NewClient` 的函數式選項設定行為：

| 選項 | 用途 |
|------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | 重定向 I/O 串流（預設 `os.Stdin`/`os.Stdout`/`os.Stderr`）。 |
| `WithProgress(fn)` | 設定 `ProgressFunc` 回呼，在 SFTP 傳輸時接收 `(description string, sent, total int64)`。SDK 自身不做任何渲染，由呼叫方決定如何展示進度。預設不設定（適合無頭環境）。 |
| `WithFileSelector(s)` | 設定 rz/sz 檔案傳輸回退用的 `FileSelector`。SDK 不提供預設實作；未設定時 rz/sz 從 stdin 讀取路徑。 |
| `WithSignalHandler()` | 註冊 Ctrl+C 處理器以關閉連線。預設不註冊，函式庫不會干擾宿主程序的訊號處理。 |

SDK 有意**不內建任何 UI 程式碼**（無進度條、無檔案對話框）。這些職責位於 CLI 套件
（`cmd/sshpass/ui.go`），它將基於 progressbar 的 `ProgressFunc` 和基於 zenity 的
`FileSelector` 接入客戶端。函式庫使用者需自行提供。

如需透過代理通道連線 SSH，在呼叫 `NewClient` 前設定 `Config.ProxyURL`：

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // 或 http://、https://、socks4://

client, err := sshpass.NewClient(cfg)
```

底層輔助函式也已匯出供進階使用：`Dial`、`NewConfig`、`LoadConfig`、
`LoadConfigOrPasswordFile`、`ParseSSHArgs`、`ParseSCPArgs`、`ParseRsyncArgs`、
`DetectCommandType`、`RunSCP`、`RunRsync`、`CleanRemotePath`、`SplitPaths`、
`ParseUserHostPath`、`ExitCodeFromError`。

## 編譯

```bash
# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass

# 交叉編譯
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=windows GOARCH=amd64               go build -o win-sshpass.exe ./cmd/sshpass
GOOS=darwin  GOARCH=arm64               go build -o win-sshpass ./cmd/sshpass
```

## 相依套件

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3（僅 CLI 進度條）
- github.com/ncruces/zenity（僅 CLI 檔案對話框）

