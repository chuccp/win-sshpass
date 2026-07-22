# 快速開始

本指南將幫助你在幾分鐘內上手 win-sshpass。

## 5 秒：執行遠端命令

```bash
# 密碼登入並執行命令
win-sshpass -p 'mypassword' ssh root@192.168.1.100 'whoami'
# → root
```

## 30 秒：檔案傳輸

```bash
# 上傳檔案
win-sshpass -h 192.168.1.100 -p 'mypassword' -local ./app.jar -remote /opt/app/

# 下載檔案
win-sshpass -h 192.168.1.100 -p 'mypassword' -d -remote /var/log/app.log -local ./logs/
```

## 1 分鐘：互動式 Shell

```bash
# 開啟互動式 Shell（不指定命令即可）
win-sshpass -p 'mypassword' ssh root@192.168.1.100
```

連線後你可以：

- 正常輸入命令，回顯正確
- 使用 `vim`、`top`、`htop` 等全螢幕應用
- 按 `Ctrl+C` 中斷當前命令
- 終端視窗大小變化時，遠端終端自動調整

## 3 分鐘：使用設定檔

建立 `server.config`：

```yaml
host: 192.168.1.100
username: root
password: mypassword
port: 22
```

使用設定檔：

```bash
# 執行命令
win-sshpass -f server.config -c 'docker ps'

# 也可以省略 -c，直接將命令作為參數
win-sshpass -f server.config 'docker ps'

# 開啟互動式 Shell
win-sshpass -f server.config
```

## 5 分鐘：SCP/Rsync 風格傳輸

```bash
# SCP 上傳
win-sshpass -p 'mypassword' scp ./app.jar user@server:/opt/app/

# SCP 上傳目錄
win-sshpass -p 'mypassword' scp -r ./dist user@server:/var/www/html

# Rsync 上傳
win-sshpass -p 'mypassword' rsync -avz ./ user@server:/backup/
```

## 常見用法速查

```bash
# 私鑰登入
win-sshpass -i ~/.ssh/id_ed25519 ssh user@server

# 環境變數傳密碼（更安全）
export SSHPASS='mypassword'
win-sshpass -e ssh user@server

# 指定連接埠
win-sshpass -p 'pass' ssh -p 2222 user@server

# 操作逾時（30秒）
win-sshpass -p 'pass' -t 30 ssh user@server 'long-running-command'

# 上傳多個檔案
win-sshpass -h server -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 產生 SSH 金鑰對
win-sshpass keygen

# 產生 RSA 金鑰，指定輸出路徑
win-sshpass keygen -algo rsa -out ~/.ssh/mykey -comment "my-servers"

# 使用產生的金鑰登入
win-sshpass -i ~/.ssh/id_ed25519 ssh user@server

# 本地檔案雜湊與校驗
win-sshpass hash sha256 ./download.iso
win-sshpass verify sha256 d1dc38f6dfb... ./download.iso
```

## 下一步

- [SSH 連線](../guide/ssh.md) - 深入了解認證方式
- [檔案傳輸](../guide/file-transfer.md) - SFTP 完整用法
- [設定檔](../guide/config-file.md) - 管理多台伺服器
