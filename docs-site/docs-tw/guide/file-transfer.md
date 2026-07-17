# 檔案傳輸

win-sshpass 提供多種檔案傳輸方式：SFTP 直接傳輸、SCP 風格、Rsync 風格，以及互動式 Shell 中的 rz/sz。

## SFTP 直接傳輸

### 上傳檔案

```bash
# 上傳單一檔案
win-sshpass -h host -p 'pass' -local ./file.txt -remote /tmp/file.txt

# 上傳多個檔案（逗號分隔）
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 上傳多個檔案（空格分隔，僅適用於簡單路徑）
win-sshpass -h host -p 'pass' -local "a.txt b.txt c.txt" -remote /tmp/

# 上傳目錄（自動遞迴）
win-sshpass -h host -p 'pass' -local ./dist -remote /var/www/html
```

### 下載檔案

```bash
# 下載檔案
win-sshpass -h host -p 'pass' -d -remote /tmp/file.txt -local ./file.txt

# 下載目錄
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

### 進度條

SFTP 傳輸時會自動顯示進度條：

```
Uploading app.jar  45% |████████████         |  45MB/100MB  10MB/s
```

## SCP 風格傳輸

相容標準 scp 命令語法：

```bash
# 上傳檔案
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# 上傳目錄（-r 遞迴）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html

# 指定連接埠（-P，注意是大寫）
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/

# 下載檔案
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# 下載目錄
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

## Rsync 風格傳輸

相容 rsync 命令語法：

```bash
# 上傳
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/

# 下載
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/

# 指定連接埠
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

## Shell 內 rz/sz 傳輸

在互動式 Shell 中，可以使用 `rz` 和 `sz` 命令進行檔案傳輸，**無需在遠端伺服器上安裝任何軟體**。

```bash
# 先開啟互動式 Shell
win-sshpass -p 'pass' ssh user@host

# 在遠端 Shell 中：
rz                              # 上傳檔案（開啟檔案選擇器）
rz /local/path/to/file          # 上傳指定本地檔案
sz /remote/path/to/file         # 下載檔案（開啟儲存對話框）
sz /remote/path/to/file /local  # 下載到指定本地路徑
```

### 工作原理

當遠端 Shell 報告 `rz`/`sz: command not found` 時，win-sshpass 會攔截該錯誤，並透過 SFTP 執行傳輸。檔案和目錄都支援，並顯示進度條。

!!! info "無需遠端安裝"
    rz/sz 傳輸基於 SFTP 實作，不需要遠端伺服器安裝 lrzsz 套件。

## 斷點續傳

傳輸大檔案時，可使用 `-resume` 參數從中斷處恢復：

```bash
# 恢復中斷的上傳
win-sshpass -p 'pass' -h host -local ./bigfile.iso -remote /data/bigfile.iso -resume

# 恢復中斷的下載
win-sshpass -p 'pass' -h host -d -remote /data/bigfile.iso -local ./bigfile.iso -resume
```

!!! info "工作原理"
    使用 `-resume` 時，win-sshpass 會檢查目標檔案是否已存在。如果存在且小於來源檔案，則從最後一個位元組處繼續傳輸。如果檔案已完整，則跳過傳輸。不使用 `-resume` 時，傳輸始終從頭開始。

## Git Bash 路徑注意

在 Git Bash 中使用時，以 `/` 開頭的遠端路徑會被自動轉換為 Windows 路徑。使用 `//` 前綴避免此問題：

```bash
# 錯誤：/tmp 會被轉換為 Windows 路徑
win-sshpass ... -remote /tmp/file.txt

# 正確：使用雙斜線
win-sshpass ... -remote //tmp/file.txt
```

## 下一步

- [SCP 與 Rsync](scp-rsync.md) - 更多 SCP/Rsync 用法
- [互動式 Shell](shell.md) - Shell 模式的完整功能
