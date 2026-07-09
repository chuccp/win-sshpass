# 設定檔

win-sshpass 支援設定檔來管理多台伺服器的連線資訊，避免每次輸入冗長的命令列參數。

## 設定檔格式

設定檔使用簡單的 `key: value` 格式：

```yaml
host: example.com
username: root
password: your_password
port: 22
```

### 支援的欄位

| 欄位 | 說明 | 預設值 |
|------|------|--------|
| `host` | 主機位址（必填） | - |
| `username` / `user` | 使用者名稱 | root |
| `password` | 密碼 | - |
| `port` | 連接埠 | 22 |
| `key` / `keypath` | 私鑰檔案路徑 | - |
| `timeout` | 操作逾時（秒），0 表示無限制 | 0 |
| `connect_timeout` | TCP 連線逾時（秒） | 10 |
| `retry` / `retries` | 連線重試次數 | 3 |
| `strict_host_key` | 啟用嚴格主機金鑰驗證 | false |

### 範例

```yaml
host: 192.168.1.100
username: ubuntu
password: mypassword
port: 22
# key: ~/.ssh/id_ed25519  # 使用私鑰認證（與 password 二選一）
# timeout: 0              # 操作逾時（秒），0 = 無限制
# connect_timeout: 10     # TCP 連線逾時（秒）
# retry: 3                # 連線重試次數
# strict_host_key: false  # 是否啟用嚴格主機金鑰驗證
```

## 使用設定檔

```bash
# 執行命令
win-sshpass -f server.config -c 'docker ps'

# 也可以省略 -c，直接將命令作為參數
win-sshpass -f server.config 'docker ps'

# 開啟互動式 Shell
win-sshpass -f server.config

# 設定檔 + SSH 風格參數
win-sshpass -f server.config ssh user@host 'ls'
```

## 密碼檔案

如果檔案不是設定檔格式（不包含 `host:` 等鍵），win-sshpass 會將其視為密碼檔案（單行文字）：

```bash
# pass.txt 內容：mypassword
win-sshpass -f pass.txt ssh user@host
```

## 設定優先順序

設定按以下優先順序合併（高優先順序覆蓋低優先順序）：

1. **命令列參數**（最高優先順序）
2. **設定檔**
3. **預設值**（最低優先順序）

例如：

```bash
# 設定檔中 port: 22，但命令列指定 -P 2222
win-sshpass -f server.config -P 2222 ssh user@host
# 實際使用連接埠 2222
```

## 多伺服器管理

為每台伺服器建立獨立的設定檔：

```
~/.ssh/
├── web-server.config
├── db-server.config
└── staging.config
```

```bash
# Web 伺服器
win-sshpass -f ~/.ssh/web-server.config 'nginx -t'

# 資料庫伺服器
win-sshpass -f ~/.ssh/db-server.config 'systemctl status mysql'

# 測試環境
win-sshpass -f ~/.ssh/staging.config 'docker ps'
```

## 安全建議

!!! warning "密碼安全"
    設定檔中包含明文密碼，請注意檔案權限：

    - 將設定檔放在安全目錄（如 `~/.ssh/`）
    - 設定適當的檔案權限（僅當前使用者可讀取）
    - 考慮使用私鑰認證代替密碼

```bash
# Linux/macOS
chmod 600 server.config

# Windows（使用 icacls）
icacls server.config /inheritance:r /grant:r %USERNAME%:R
```

## 下一步

- [SSH 連線](ssh.md) - 更多認證方式
- [最佳實踐](../advanced/best-practices.md) - 安全與效率建議
