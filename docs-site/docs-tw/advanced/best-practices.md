# 最佳實踐

## 安全建議

### 1. 避免在命令列中直接傳遞密碼

```bash
# 不推薦：密碼會出現在命令歷史中
win-sshpass -p 'mypassword' ssh user@host

# 推薦：使用環境變數
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推薦：使用密碼檔案
win-sshpass -f pass.txt ssh user@host

# 推薦：使用設定檔
win-sshpass -f server.config ssh user@host
```

### 2. 使用私鑰認證

私鑰認證比密碼認證更安全：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 保護設定檔權限

```bash
# Linux/macOS
chmod 600 server.config

# Windows（PowerShell）
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. 啟用主機金鑰驗證

在正式環境中，建議啟用嚴格主機金鑰驗證：

```bash
win-sshpass -k -f server.config ssh user@host
```

或在設定檔中：

```yaml
strict_host_key: true
```

## 效率建議

### 1. 使用設定檔管理多台伺服器

為常用伺服器建立設定檔，避免重複輸入參數：

```bash
# ~/.ssh/prod-web.config
host: web.example.com
username: deploy
key: ~/.ssh/id_ed25519

# ~/.ssh/prod-db.config
host: db.example.com
username: admin
key: ~/.ssh/id_ed25519
```

### 2. 批次操作

結合 Shell 指令碼進行批次操作：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. 使用 SSH 風格語法

對於熟悉 SSH 的使用者，可以使用更自然的語法：

```bash
# 標準 SSH 語法
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 語法
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 語法
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 設定合理的逾時

```bash
# 快速命令：短逾時
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 長時間操作：長逾時或無逾時
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

## 故障排除

### 連線失敗

```bash
# 增加重試次數
win-sshpass -p 'pass' -retry 5 ssh user@host

# 增加連線逾時
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 認證失敗

- 檢查密碼是否正確
- 檢查私鑰路徑是否正確
- 檢查遠端伺服器是否允許密碼/金鑰認證
- 注意：不支援加密的私鑰

### Git Bash 路徑問題

```bash
# 錯誤：/tmp 會被 Git Bash 轉換
win-sshpass ... -remote /tmp/file.txt

# 正確：使用雙斜線
win-sshpass ... -remote //tmp/file.txt
```

## 下一步

- [Go SDK](sdk.md) - 以程式設計方式使用
- [更新日誌](../changelog.md) - 版本更新記錄
