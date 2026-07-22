# Go SDK

win-sshpass はコマンドラインツールであるだけでなく、再利用可能な Go ライブラリ（`package sshpass`）でもあります。SSH/SFTP/Shell 機能を自分のアプリケーションに組み込むことができます。

## インストール

```bash
go get github.com/chuccp/win-sshpass
```

## クイックスタート

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

    // NewClient は接続を確立し、すぐに使えるクライアントを返します
    client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // コマンド実行（出力ストリームはデフォルトで os.Stdout/os.Stderr）
    if err := client.Exec("uname -a"); err != nil {
        log.Fatal(err)
    }
}
```

## Client のコアメソッド

### NewClient

SSH 接続を確立し、クライアントを返します：

```go
client, err := sshpass.NewClient(cfg, opts...)
```

- `cfg`: 接続設定（`*sshpass.Config`）
- `opts`: オプションの関数オプション（`...sshpass.Option`）
- `cfg.Timeout > 0` の場合、操作タイマーが設定され、期限切れになると接続が閉じられます

### Exec

単一のリモートコマンドを実行：

```go
err := client.Exec("ls -la")
```

- 出力ストリームは `WithStdout`/`WithStderr` で設定（デフォルト：`os.Stdout`/`os.Stderr`）
- 入力ストリームは `WithStdin` で設定（デフォルト：`os.Stdin`）

### Shell

インタラクティブシェルを開始：

```go
err := client.Shell()
```

- ターミナルを自動検出し、PTY とローモードをサポート
- 動的ターミナルリサイズをサポート
- rz/sz ファイル転送をサポート（`WithFileSelector` の設定が必要）

### SFTP

SFTP サブチャンネルを開く：

```go
sftp, err := client.SFTP()
if err != nil {
    log.Fatal(err)
}
defer sftp.Close()

// ファイルのアップロード
err = sftp.Upload("./local.txt", "/tmp/remote.txt")

// ファイルのダウンロード
err = sftp.Download("/tmp/remote.txt", "./local.txt")

// 基底の *sftp.Client にアクセス（上級者向け）
rawClient := sftp.SFTP()
```

### Close

SSH 接続を閉じる：

```go
err := client.Close()
```

- 冪等操作 — 複数回呼び出しても同じエラーを返します
- タイマーとシグナルハンドラーを自動的に停止

### TimedOut

タイムアウトが原因で失敗したかチェック：

```go
if client.TimedOut() {
    fmt.Println("操作がタイムアウトしました")
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
cfg.Timeout = 30          // 操作タイムアウト（秒）、0 = 無制限
cfg.ConnectTimeout = 10   // TCP 接続タイムアウト（秒）
cfg.Retries = 3           // 接続リトライ回数
```

### 設定ファイルから読み込み

```go
cfg, err := sshpass.LoadConfig("server.config")
```

### 設定ファイルまたはパスワードファイルの読み込み

```go
cfg, pass, err := sshpass.LoadConfigOrPasswordFile("file.txt", "", false)
if cfg != nil {
    // 設定ファイルです
} else {
    // パスワードファイルです。pass にパスワードが含まれています
}
```

## 鍵生成

プログラムで SSH 鍵ペアを生成：

```go
// Ed25519 鍵ペアを生成（推奨）
pair, err := sshpass.GenerateKeyPair(sshpass.KeyEd25519, "user@host")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("秘密鍵:\n%s\n", pair.PrivateKey)
fmt.Printf("公開鍵:\n%s\n", pair.PublicKey)

// RSA 鍵ペアを生成（最小 2048 ビット）
pair, err = sshpass.GenerateRSAKeyPair(4096, "user@host")

// 鍵ペアをファイルに保存
err = sshpass.SaveKeyPair(pair, "~/.ssh/mykey")
// 作成されるファイル: ~/.ssh/mykey (秘密鍵, 0600) と ~/.ssh/mykey.pub (公開鍵)

// デフォルトの鍵パスを取得
path := sshpass.DefaultKeyPath(sshpass.KeyEd25519) // ~/.ssh/id_ed25519

// 公開鍵をリモートサーバーに展開（既存のクライアント接続が必要）
err = sshpass.DeployPublicKey(client, pair.PublicKey)
```

## 関数オプション

関数オプションで Client の動作を設定：

### I/O ストリーム

```go
// 入出力のリダイレクト
client, err := sshpass.NewClient(cfg,
    sshpass.WithStdin(myReader),
    sshpass.WithStdout(myWriter),
    sshpass.WithStderr(myWriter),
)
```

### プログレスコールバック

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithProgress(func(desc string, sent, total int64) {
        fmt.Printf("\r%s %d/%d bytes", desc, sent, total)
    }),
)
```

- `desc`: 転送の説明（例："Uploading file.txt"）
- `sent`: これまでに転送されたバイト数
- `total`: ファイルの合計サイズ
- SDK はレンダリングを行いません — 呼び出し元がプログレスを表示します

### ファイルセレクター

```go
type myFileSelector struct{}

func (s myFileSelector) OpenFile() (string, error) {
    // ファイルオープンダイアログを実装
    return "/path/to/file", nil
}

func (s myFileSelector) SaveFile(defaultName string) (string, error) {
    // ファイル保存ダイアログを実装
    return "/path/to/save", nil
}

client, err := sshpass.NewClient(cfg,
    sshpass.WithFileSelector(myFileSelector{}),
)
```

rz/sz シェル転送のファイル選択に使用されます。SDK はデフォルトの実装を提供しません。

### シグナルハンドラー

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithSignalHandler(),
)
```

接続を閉じる Ctrl+C ハンドラーを登録します。デフォルトではオフで、ホストのシグナル処理を妨げません。

### ブレークポイントレジューム

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithResume(),
)
```

SFTP 転送のブレークポイントレジュームを有効にします。中断されたアップロード/ダウンロードは中断した箇所から再開します。

### プロキシ設定

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // または http://, https://, socks4://
```

`ProxyURL` を設定すると、SSH 接続は指定されたプロキシを経由してトンネリングされます。対応プロトコル: SOCKS5 (認証オプション付き)、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

## 低レベル API

### Dial

SSH クライアント接続を直接作成：

```go
sshClient, err := sshpass.Dial(cfg)
```

より低レベルの制御が必要なシナリオ用に `*ssh.Client` を返します。

### 引数の解析

```go
// SSH 引数の解析
config, cmd := sshpass.ParseSSHArgs([]string{"ssh", "user@host", "ls"})

// SCP 引数の解析
config, args := sshpass.ParseSCPArgs([]string{"scp", "file.txt", "user@host:/tmp/"})

// Rsync 引数の解析
config, args := sshpass.ParseRsyncArgs([]string{"rsync", "-avz", "./", "user@host:/backup/"})

// コマンドタイプの検出
cmdType := sshpass.DetectCommandType(args)
```

### ユーティリティ関数

```go
// SCP 転送の実行
err := sshpass.RunSCP(client, args)

// Rsync 転送の実行
err := sshpass.RunRsync(client, args)

// リモートパスのクリーンアップ（Git Bash パス変換の処理）
path, err := sshpass.CleanRemotePath("//tmp/file.txt")

// user@host:path フォーマットの解析
user, host, path := sshpass.ParseUserHostPath("user@host:/tmp/file.txt")

// パスの分割（カンマまたはスペース区切り）
paths, err := sshpass.SplitPaths("a.txt,b.txt,c.txt", "local")

// エラーから終了コードを抽出
code, ok := sshpass.ExitCodeFromError(err)
```

## 完全な例

### バッチコマンド実行

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
            log.Printf("[%s] 接続失敗: %v", host, err)
            continue
        }

        fmt.Printf("[%s] コマンド実行中...\n", host)
        if err := client.Exec("uptime"); err != nil {
            log.Printf("[%s] コマンド失敗: %v", host, err)
        }
        client.Close()
    }
}
```

### プログレス付きファイルアップロード

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
    fmt.Println("\nアップロード完了!")
}
```

## 次のステップ

- [ベストプラクティス](best-practices.md) - セキュリティと効率のヒント
