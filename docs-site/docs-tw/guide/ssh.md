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

## 指定使用者和連接埠

```bash
# 指定使用者名稱（預設 root）
win-sshpass -p 'pass' ssh ubuntu@host

# 指定連接埠（預設 22）
win-sshpass -p 'pass' ssh -p 2222 user@host

# 使用 -u 和 -P 參數
win-sshpass -p 'pass' -h host -u ubuntu -P 2222
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

## 下一步

- [互動式 Shell](shell.md) - 不指定命令時的互動模式
- [檔案傳輸](file-transfer.md) - SFTP 上傳下載
- [設定檔](config-file.md) - 管理多台伺服器
