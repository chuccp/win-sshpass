# SCP 與 Rsync

win-sshpass 相容標準 scp 和 rsync 命令語法，底層透過 SFTP 實作檔案傳輸。

## SCP 風格傳輸

### 基本語法

```bash
win-sshpass -p <password> scp [選項] <來源> <目標>
```

### 上傳檔案

```bash
# 上傳單一檔案
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# 上傳到指定檔名
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/newname.txt

# 上傳目錄（-r 遞迴）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html
```

### 下載檔案

```bash
# 下載檔案
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# 下載目錄
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

### 指定連接埠

scp 使用大寫 `-P` 指定連接埠（與 ssh 的小寫 `-p` 不同）：

```bash
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/
```

### 支援的選項

| 選項 | 說明 |
|------|------|
| `-r` | 遞迴複製目錄 |
| `-P <port>` | 指定連接埠 |
| `-i <key>` | 指定私鑰 |
| `-q` | 靜默模式 |
| `-C` | 壓縮（已由 SFTP 處理） |
| `-v` | 詳細輸出 |

## Rsync 風格傳輸

### 基本語法

```bash
win-sshpass -p <password> rsync [選項] <來源> <目標>
```

### 上傳

```bash
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 下載

```bash
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/
```

### 指定連接埠

rsync 使用 `--port=` 指定連接埠：

```bash
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

### 支援的選項

| 選項 | 說明 |
|------|------|
| `-a` | 封存模式 |
| `-v` | 詳細輸出 |
| `-z` | 傳輸時壓縮 |
| `--port=N` | 指定連接埠 |
| `-e ssh` | 指定遠端 Shell（會被忽略） |

## SCP vs Rsync vs SFTP

| 方式 | 語法相容 | 適用場景 |
|------|----------|----------|
| SCP | 標準 scp 語法 | 簡單的檔案複製 |
| Rsync | 標準 rsync 語法 | 增量同步（注意：當前實作為全量傳輸） |
| SFTP | `-local` / `-remote` 參數 | 靈活的檔案傳輸，支援多檔案 |

!!! note "注意"
    win-sshpass 的 rsync 實作底層使用 SFTP，不支援 rsync 的增量同步演算法。如果需要真正的增量同步，請在遠端伺服器上安裝 rsync 並透過 SSH 直接使用。

## 下一步

- [檔案傳輸](file-transfer.md) - SFTP 直接傳輸的更多用法
- [設定檔](config-file.md) - 使用設定檔簡化命令
