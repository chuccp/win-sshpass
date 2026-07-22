# Go SDK

win-sshpass 不僅是一個命令列工具，也是一個可重複使用的 Go 函式庫（`package sshpass`）。你可以將 SSH/SFTP/Shell 功能嵌入到自己的應用中。

## 安裝

```bash
go get github.com/chuccp/win-sshpass
```

## 快速開始

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

    // NewClient 建立 SSH 連線並回傳可用的用戶端
    client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 執行命令（輸出串流預設為 os.Stdout/os.Stderr）
    if err := client.Exec("uname -a"); err != nil {
        log.Fatal(err)
    }
}
```

## Client 核心方法

### NewClient

建立 SSH 連線並回傳用戶端：

```go
client, err := sshpass.NewClient(cfg, opts...)
```

- `cfg`：連線設定（`*sshpass.Config`）
- `opts`：可選的函式選項（`...sshpass.Option`）
- 如果 `cfg.Timeout > 0`，會啟動操作計時器，逾時後自動關閉連線

### Exec

執行單條遠端命令：

```go
err := client.Exec("ls -la")
```

- 輸出串流透過 `WithStdout`/`WithStderr` 設定（預設 `os.Stdout`/`os.Stderr`）
- 輸入串流透過 `WithStdin` 設定（預設 `os.Stdin`）

### Shell

啟動互動式 Shell：

```go
err := client.Shell()
```

- 自動偵測終端，支援 PTY 和 raw 模式
- 支援動態終端調整大小
- 支援 rz/sz 檔案傳輸（需要設定 `WithFileSelector`）

### SFTP

開啟 SFTP 子通道：

```go
sftp, err := client.SFTP()
if err != nil {
    log.Fatal(err)
}
defer sftp.Close()

// 上傳檔案
err = sftp.Upload("./local.txt", "/tmp/remote.txt")

// 下載檔案
err = sftp.Download("/tmp/remote.txt", "./local.txt")

// 存取底層 *sftp.Client（進階用法）
rawClient := sftp.SFTP()
```

### Close

關閉 SSH 連線：

```go
err := client.Close()
```

- 冪等操作，多次呼叫回傳相同的錯誤
- 自動停止計時器和訊號處理器

### TimedOut

檢查是否因逾時而失敗：

```go
if client.TimedOut() {
    fmt.Println("操作逾時")
}
```

## 設定（Config）

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.Port = "22"
cfg.KeyPath = "~/.ssh/id_ed25519"
cfg.StrictHostKey = true
cfg.Timeout = 30          // 操作逾時（秒），0 = 無限制
cfg.ConnectTimeout = 10   // TCP 連線逾時（秒）
cfg.Retries = 3           // 連線重試次數
```

### 從設定檔載入

```go
cfg, err := sshpass.LoadConfig("server.config")
```

### 載入設定檔或密碼檔案

```go
cfg, pass, err := sshpass.LoadConfigOrPasswordFile("file.txt", "", false)
if cfg != nil {
    // 是設定檔
} else {
    // 是密碼檔案，pass 包含密碼
}
```

## 金鑰產生

以程式化方式產生 SSH 金鑰對：

```go
// 產生 Ed25519 金鑰對（推薦）
pair, err := sshpass.GenerateKeyPair(sshpass.KeyEd25519, "user@host")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("私鑰：\n%s\n", pair.PrivateKey)
fmt.Printf("公鑰：\n%s\n", pair.PublicKey)

// 產生 RSA 金鑰對（至少 2048 位元）
pair, err = sshpass.GenerateRSAKeyPair(4096, "user@host")

// 將金鑰對儲存到檔案
err = sshpass.SaveKeyPair(pair, "~/.ssh/mykey")
// 建立：~/.ssh/mykey（私鑰，0600）和 ~/.ssh/mykey.pub（公鑰）

// 取得預設金鑰路徑
path := sshpass.DefaultKeyPath(sshpass.KeyEd25519) // ~/.ssh/id_ed25519

// 將公鑰部署到遠端伺服器（需要已有的用戶端連線）
err = sshpass.DeployPublicKey(client, pair.PublicKey)
```

## 函式選項

透過函式選項設定 User 的行為：

### I/O 串流

```go
// 重導向輸入/輸出
client, err := sshpass.NewClient(cfg,
    sshpass.WithStdin(myReader),
    sshpass.WithStdout(myWriter),
    sshpass.WithStderr(myWriter),
)
```

### 進度回呼

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithProgress(func(desc string, sent, total int64) {
        fmt.Printf("\r%s %d/%d bytes", desc, sent, total)
    }),
)
```

- `desc`：傳輸說明（如 "Uploading file.txt"）
- `sent`：已傳輸位元組數
- `total`：檔案總大小
- SDK 不做任何渲染，呼叫者自行顯示進度

### 檔案選擇器

```go
type myFileSelector struct{}

func (s myFileSelector) OpenFile() (string, error) {
    // 實作檔案開啟對話框
    return "/path/to/file", nil
}

func (s myFileSelector) SaveFile(defaultName string) (string, error) {
    // 實作檔案儲存對話框
    return "/path/to/save", nil
}

client, err := sshpass.NewClient(cfg,
    sshpass.WithFileSelector(myFileSelector{}),
)
```

用於 rz/sz Shell 傳輸的檔案選擇。SDK 不提供預設實作。

### 訊號處理器

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithSignalHandler(),
)
```

註冊 Ctrl+C 處理器，按下時關閉連線。預設不註冊，以免干擾宿主程序的訊號處理。

### 斷點續傳

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithResume(),
)
```

啟用 SFTP 傳輸的斷點續傳功能 —— 中斷的上傳/下載將從斷點處繼續。

### 代理配置

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // 或 http://、https://、socks4://
```

設定 `ProxyURL` 後，SSH 連線將透過指定的代理伺服器進行通道傳輸。支援的協定：SOCKS5（可選認證）、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

## 底層 API

### Dial

直接建立 SSH 用戶端連線：

```go
sshClient, err := sshpass.Dial(cfg)
```

回傳 `*ssh.Client`，適用於需要更底層控制的場景。

### 參數解析

```go
// 解析 SSH 參數
config, cmd := sshpass.ParseSSHArgs([]string{"ssh", "user@host", "ls"})

// 解析 SCP 參數
config, args := sshpass.ParseSCPArgs([]string{"scp", "file.txt", "user@host:/tmp/"})

// 解析 Rsync 參數
config, args := sshpass.ParseRsyncArgs([]string{"rsync", "-avz", "./", "user@host:/backup/"})

// 偵測命令類型
cmdType := sshpass.DetectCommandType(args)
```

### 工具函式

```go
// 執行 SCP 傳輸
err := sshpass.RunSCP(client, args)

// 執行 Rsync 傳輸
err := sshpass.RunRsync(client, args)

// 清理遠端路徑（處理 Git Bash 路徑轉換）
path, err := sshpass.CleanRemotePath("//tmp/file.txt")

// 解析 user@host:path 格式
user, host, path := sshpass.ParseUserHostPath("user@host:/tmp/file.txt")

// 分割路徑（逗號或空格分隔）
paths, err := sshpass.SplitPaths("a.txt,b.txt,c.txt", "local")

// 從錯誤中提取結束碼
code, ok := sshpass.ExitCodeFromError(err)
```

## 完整範例

### 批次執行命令

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    hosts := []string{"192.168.1.101", "192.168.1.102", "192.168.1.103"}

    for _, host := range hosts {
        cfg := sshpass.NewConfig()
        cfg.Host = host
        cfg.User = "root"
        cfg.Password = "secret"

        client, err := sshpass.NewClient(cfg)
        if err != nil {
            log.Printf("[%s] 連線失敗: %v", host, err)
            continue
        }

        fmt.Printf("[%s] 執行命令...\n", host)
        if err := client.Exec("uptime"); err != nil {
            log.Printf("[%s] 執行失敗: %v", host, err)
        }
        client.Close()
    }
}
```

### 帶進度的檔案上傳

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    cfg := sshpass.NewConfig()
    cfg.Host = "example.com"
    cfg.User = "root"
    cfg.Password = "secret"

    client, err := sshpass.NewClient(cfg,
        sshpass.WithProgress(func(desc string, sent, total int64) {
            pct := sent * 100 / total
            fmt.Printf("\r%s %d%%", desc, pct)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    sftp, err := client.SFTP()
    if err != nil {
        log.Fatal(err)
    }
    defer sftp.Close()

    if err := sftp.Upload("./large-file.zip", "/tmp/large-file.zip"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("\n上傳完成!")
}
```

## 下一步

- [最佳實踐](best-practices.md) - 安全與效率建議
