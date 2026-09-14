# SSH 連線

win-sshpass 支援多種 SSH 認證方式，滿足不同場景需求。

## 密碼認證

### 直接指定密碼

```bash
win-sshpass -p 'mypassword' ssh user@host
win-sshpass -p 'mypassword' ssh user@host 'whoami'
```

### 從檔案讀取密碼

建立一個只包含密碼的文字檔（單行）：

```bash
echo 'mypassword' > pass.txt
win-sshpass -f pass.txt ssh user@host
```

### 從環境變數讀取密碼

```bash
export SSHPASS='mypassword'
win-sshpass -e ssh user@host
```

或在 Windows CMD 中：

```cmd
set SSHPASS=mypassword
win-sshpass -e ssh user@host
```

!!! tip "安全性建議"
    使用環境變數或設定檔比在命令列中直接傳遞密碼更安全，因為命令歷史中不會記錄密碼。

### 空密碼

當伺服器帳號沒有設定密碼時，明確傳入空密碼。省略 `-p` 意義不同——那會回退到 ssh-agent：

```bash
win-sshpass -p '' ssh root@192.168.1.100 'hostname'
```

!!! note "引號寫法"
    PowerShell 與 Git Bash 用 `''`；Windows CMD 用 `""`。

## 私鑰認證

```bash
# 使用 Ed25519 金鑰
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 使用 RSA 金鑰
win-sshpass -i ~/.ssh/id_rsa ssh user@host

# 執行遠端命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'uname -a'
```

!!! note "注意"
    win-sshpass 不支援加密（有密碼保護）的私鑰。如果私鑰有密碼保護，請先解密或使用 ssh-agent。

## SSH Agent 認證

win-sshpass 可以自動使用本機 ssh-agent（如 OpenSSH agent、Windows 上的 Pageant）進行認證。當未指定 `-p`（密碼）或 `-i`（金鑰路徑）時，預設啟用 ssh-agent 自動檢測：

```bash
# 自動檢測並使用 ssh-agent（無需 -p 或 -i）
win-sshpass ssh user@host 'whoami'

# 同樣適用於檔案傳輸和 SCP/Rsync
win-sshpass -h host -local file.txt -remote /tmp/file.txt
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
```

!!! tip "ssh-agent 設定"
    確保 ssh-agent 正在執行並已載入金鑰（使用 `ssh-add -l` 檢查）。在 Windows 上，可以在「服務」中啟用 OpenSSH Authentication Agent 服務。

### Agent 轉送（`-A`）

使用 `-A` 參數將本機 ssh-agent 轉送到遠端伺服器，允許遠端主機使用本機金鑰進行進一步的 SSH 連線（例如從跳板機執行 `git clone`）：

```bash
# 啟用 agent 轉送
win-sshpass -A -i ~/.ssh/id_ed25519 ssh user@jumphost

# 無密碼/金鑰 — 自動檢測 agent + 轉送
win-sshpass -A ssh user@jumphost
```

!!! info "Agent 轉送需要本機 Agent"
    Agent 轉送（`-A`）需要執行中的本機 ssh-agent 並已載入金鑰。啟用後轉送將自動處理。

## 鍵盤互動式認證

當標準密碼認證不可用時，win-sshpass 會自動回退到鍵盤互動式認證。這確保與基於 PAM 的伺服器（如 Cisco 路由器、部分使用自訂 PAM 設定的 Linux 發行版）相容。無需額外參數 —— 回退是透明的。

## 金鑰生成

win-sshpass 內建了 SSH 金鑰對生成功能，可以在本地生成客戶端金鑰對（私鑰 + 公鑰）。

```bash
# 生成 Ed25519 金鑰（推薦，更快更安全）
win-sshpass keygen

# 生成 RSA 金鑰（4096 位元）
win-sshpass keygen -algo rsa

# 指定輸出路徑
win-sshpass keygen -out ~/.ssh/mykey

# 指定公鑰註釋
win-sshpass keygen -comment "my-laptop"
```

預設輸出到 `~/.ssh/id_ed25519`（Ed25519）或 `~/.ssh/id_rsa`（RSA），公鑰檔案自動新增 `.pub` 後綴。

生成後，將公鑰部署到服務端即可實現無密碼登入（見下文）。

### 部署公鑰實現無密碼登入

產生金鑰後，將公鑰部署到伺服器即可實現無密碼登入：

```bash
# 將公鑰內容讀入變數，再透過 SSH 部署
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'mypassword' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

Shell 在本機展開 `$PUBKEY` 後再將命令傳給 win-sshpass，因此實際的公鑰字串被嵌入到遠端命令中，避免了引號轉義和標準輸入轉送的問題。

部署完成後，即可使用私鑰進行無密碼登入：

```bash
# 無密碼登入
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 無密碼執行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'whoami'

# 無密碼傳輸檔案
win-sshpass -i ~/.ssh/id_ed25519 scp file.txt user@host:/tmp/
```

!!! tip "authorized_keys 權限要求"
    服務端 `~/.ssh` 目錄權限應為 700，`~/.ssh/authorized_keys` 權限應為 600。權限不正確會導致金鑰認證失敗。

## 指定使用者和連接埠

```bash
# 指定使用者名稱（預設 root）
win-sshpass -p 'pass' ssh ubuntu@host

# 指定連接埠（預設 22）
win-sshpass -p 'pass' ssh -p 2222 user@host

# 使用 -u 和 -P 參數
win-sshpass -p 'pass' -h host -u ubuntu -P 2222

# 主機名稱可不帶 user@（使用者名稱預設 root，主機名稱交由系統解析）
win-sshpass -p 'pass' ssh example.com 'uptime'
```

## 執行遠端命令

```bash
# 單條命令
win-sshpass -p 'pass' ssh user@host 'ls -la'

# 多條命令
win-sshpass -p 'pass' ssh user@host 'cd /var/log && ls -la'

# 使用 -c 參數
win-sshpass -p 'pass' -h host -c 'docker ps'
```

## 連線逾時與重試

```bash
# TCP 連線逾時（預設 10 秒）
win-sshpass -p 'pass' -ct 5 ssh user@host

# 操作逾時（預設無限制）
win-sshpass -p 'pass' -t 30 ssh user@host 'long-command'

# 重試次數（預設 3 次）
win-sshpass -p 'pass' -retry 5 ssh user@host
```

逾時機制說明：

- **TCP 連線逾時**（`-ct`）：建立 TCP 連線的逾時時間
- **操作逾時**（`-t`）：整個操作的逾時時間，資料傳輸時會自動重置計時器
- **重試**（`-retry`）：連線失敗後的重試次數，採用指數退避策略（2s、4s、8s、16s，最大 30s）

!!! info "認證失敗不重試"
    如果是認證失敗（密碼錯誤、金鑰無效），不會進行重試，直接回傳錯誤。

## 主機金鑰驗證

預設情況下，win-sshpass 不驗證主機金鑰（等同於 `StrictHostKeyChecking=no`）。

啟用嚴格主機金鑰驗證：

```bash
# 使用 -k 參數
win-sshpass -p 'pass' -k ssh user@host

# 或在設定檔中設定
# strict_host_key: true
```

啟用後，會使用 `~/.ssh/known_hosts` 檔案進行驗證。如果主機不在 known_hosts 中，連線會被拒絕。

## IPv6 支援

win-sshpass 支援 IPv6 位址：

```bash
win-sshpass -p 'pass' ssh user@2001:db8::1
win-sshpass -p 'pass' ssh user@[2001:db8::1]
```

## 代理支援

透過代理伺服器建立 SSH 通道連線：

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

!!! info "支援的代理協定"
    支援 SOCKS5（可選使用者名稱/密碼認證）、SOCKS4、SOCKS4A、HTTP CONNECT 和 HTTPS CONNECT 代理。

## 連接埠轉送

win-sshpass 支援 SSH 連接埠轉送，可透過 SSH 伺服器建立 TCP 隧道連線。本機轉送（`-L`）和遠端轉送（`-R`）均使用標準的 OpenSSH 規範格式。

### 本機轉送（`-L`）

透過 SSH 伺服器將本機連接埠轉送到遠端位址：

```bash
# 格式：-L [繫結位址:]連接埠:主機:主機連接埠
# 將 localhost:8080 → db.internal:3306 透過 SSH 主機轉送
win-sshpass -p 'pass' -L 8080:db.internal:3306 ssh user@jumphost

# 繫結到特定本機位址
win-sshpass -p 'pass' -L 127.0.0.1:8080:db.internal:3306 ssh user@jumphost

# 多個轉送
win-sshpass -p 'pass' -L 8080:db1.internal:3306 -L 8081:db2.internal:3306 ssh user@jumphost

# 僅轉送模式（無命令 — 阻塞直到 Ctrl+C）
win-sshpass -p 'pass' -L 8080:db.internal:3306 -L 9090:redis.internal:6379 ssh user@jumphost
```

### 遠端轉送（`-R`）

將遠端連接埠轉送回本機位址：

```bash
# 格式：-R [繫結位址:]連接埠:主機:主機連接埠
# 將 localhost:8080 暴露在遠端伺服器的 9090 連接埠
win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server

# 允許遠端主機連線
win-sshpass -p 'pass' -R 0.0.0.0:9090:localhost:8080 ssh user@server
```

!!! info "連接埠轉送限制"
    連接埠轉送僅支援 SSH 命令/Shell 模式。無法與 SCP、Rsync 或檔案傳輸（`-local`/`-remote`）操作同時使用。

## 檔案雜湊與校驗

無需 SSH 連線即可計算和校驗本地檔案雜湊：

```bash
# 計算雜湊
win-sshpass hash md5 ./download.iso
win-sshpass hash sha1 ./download.iso
win-sshpass hash sha256 ./download.iso
win-sshpass hash sha512 ./download.iso

# 校驗檔案完整性
win-sshpass verify sha256 d1dc38f6dfb1e4c8e7a1b2c3d4e5f6a7b8c9d0e1f2 ./download.iso
# 輸出: OK

win-sshpass verify sha256 wronghash123... ./download.iso
# 輸出: FAILED
```

支援的演算法：`md5`、`sha1`、`sha256`、`sha512`。

## 下一步

- [互動式 Shell](shell.md) - 不指定命令時的互動模式
- [檔案傳輸](file-transfer.md) - SFTP 上傳下載
- [設定檔](config-file.md) - 管理多台伺服器
