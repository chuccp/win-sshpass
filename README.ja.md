## 関連プロジェクト

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — 認証の問題を簡単に解決——ルート宣言に必要な権限、Filterで一括チェック、handlerはクリーンに。ジェネリクスModelでフルスタックCRUD——structを定義するだけで、作成・読取・更新・削除がすぐ使える。軽量で、必要なコンポーネントだけオンデマンドでインストール。コード生成不要、CIツール不要、今最も精巧なGoウェブフルスタックフレームワーク。

# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版および Linux 版 sshpass ツール。Linux の sshpass と同様の機能を提供します。

> 💡 **このプロジェクトが役に立ったら、⭐ Star をお願いします！** より多くの人にこのツールを見つけてもらえます。

## 機能

- パスワードまたは秘密鍵認証による SSH ログイン
- リモートコマンドの実行またはインタラクティブシェル
- SFTP によるファイルのアップロード/ダウンロード（プログレスバー付き）
- SCP スタイルおよび Rsync スタイルのファイル転送
- 複数サーバー管理用の設定ファイルサポート
- インタラクティブシェルの raw ターミナルモード（正しいエコー、Ctrl+C、vim/top フルスクリーンアプリ対応）
- インタラクティブシェルモードでの動的ターミナルサイズ変更
- Git Bash パス変換の検出と自動修正
- IPv6 アドレス対応
- Windows（x64、ARM64）および Linux（amd64、arm64）対応
- **再利用可能な Go SDK** — ライブラリとしてインポート（`package sshpass`）して、SSH/SFTP/Shell 機能を独自アプリに組み込み可能。I/O ストリームと進行コールバックの注入に対応
- **プロキシ対応** — SOCKS5/SOCKS4/HTTP/HTTPS プロキシ経由で SSH 接続をトンネル
- **ブレークポイントレジューム** — 中断された SFTP ファイル転送を途中から再開
- **ファイルハッシュと検証** — ローカルファイルのハッシュ計算と検証（MD5、SHA-1、SHA-256、SHA-512）
- **鍵生成** — 内蔵SSH鍵ペア生成（Ed25519 と RSA）

## ダウンロード

[GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から最新版をダウンロード：

### Windows

| アーキテクチャ | Zip | MSI インストーラー |
|----------------|-----|-------------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

### Linux

| アーキテクチャ | Tarball |
|----------------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

1. [Releases](https://github.com/chuccp/win-sshpass/releases) ページを開く
2. お使いのプラットフォームとアーキテクチャに合ったパッケージをダウンロード
3. **Windows MSI**：インストーラーを実行するだけで、インストール先がシステム PATH に自動的に追加されます
4. **Windows Zip / Linux tar.gz**：展開してバイナリを PATH に配置

> **依存関係ゼロ**：`win-sshpass.exe` はスタンドアロンの実行ファイルです。OpenSSH やその他のソフトウェアのインストールは不要です。ダウンロードして PATH に配置すればすぐに使用できます。

### Scoop でインストール

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### WinGet でインストール

```bash
winget install chuccp.win-sshpass
```

## クイックスタート

```bash
# パスワード認証でコマンド実行
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 秘密鍵認証でコマンド実行
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# ファイルをアップロード
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# ファイルをダウンロード
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## インタラクティブシェル

コマンドを指定しない場合、`win-sshpass` は **raw ターミナルモード** でインタラクティブシェルを開きます：

```bash
win-sshpass -p 'password' ssh user@host
```

**Raw ターミナルモード** の機能：

- **正しいエコーバック** — 入力した文字が正しく表示される（二重エコーなし）
- **Ctrl+C / Ctrl+Z** — シグナルがリモートプロセスに正しく転送される
- **フルスクリーンアプリ** — vim、top、htop、nano などが正常に動作
- **動的ターミナルサイズ変更** — リモートターミナルがローカルウィンドウサイズに自動調整
- **Tab 補完** — リモートシェルの Tab 補完が正常に機能

### インタラクティブシェルでのファイル転送

接続中に `rz` / `sz` コマンドでファイル転送が可能（リモートサーバーに何もインストールする必要なし）：

```bash
# リモートの現在のディレクトリにファイルをアップロード（ファイル選択ダイアログが開く）
rz

# 特定のローカルファイルをアップロード
rz /ローカル/ファイル/パス

# リモートファイルをダウンロード（保存ダイアログが開く）
sz /リモート/ファイル/パス

# リモートファイルを特定のローカルパスにダウンロード
sz /リモート/ファイル/パス /ローカル/保存/パス
```

> **仕組み**：リモートシェルが `rz`/`sz: command not found` を返した時、ツールがそれを検出して SFTP で転送を実行します。ファイルとディレクトリの両方に対応し、プログレスバー付きです。

## コマンド形式

### SSH ログイン

```bash
# パスワード認証
win-sshpass -p <パスワード> ssh [user@host] [コマンド]
win-sshpass -p <パスワード> ssh -p <ポート> user@host 'コマンド'
win-sshpass -p <パスワード> ssh -o StrictHostKeyChecking=no user@host

# インタラクティブシェル（raw ターミナルモード：正しいエコー、Ctrl+C、vim/top 対応）
win-sshpass -p <パスワード> ssh user@host

# 秘密鍵認証
win-sshpass -i <秘密鍵パス> ssh [user@host] [コマンド]

# 環境変数からパスワード読み込み
SSHPASS=<パスワード> win-sshpass -e ssh user@host

# パスワードファイル
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 設定ファイル（複数行形式）
win-sshpass -f server.config
```

### ファイル転送

> **Git Bash ユーザー**: リモートパスには `//` プレフィックスを使用してください（例: `-remote //tmp/file.txt）。詳細は下記の [Git Bash の注意事項](#git-bash-の注意事項) を参照してください。

```bash
# ファイルをアップロード
win-sshpass -h <ホスト> -p <パスワード> -local <ローカルパス> -remote <リモートパス>

# 複数のファイルをアップロード（カンマ区切り）
win-sshpass -h <ホスト> -p <パスワード> -local "a.txt,b.txt,c.txt" -remote //tmp/

# 複数のファイルをアップロード（スペース区切り、/ や \ を含まない単純なパスのみ）
win-sshpass -h <ホスト> -p <パスワード> -local "a.txt b.txt c.txt" -remote //tmp/

# ディレクトリをアップロード（自動再帰）
win-sshpass -h <ホスト> -p <パスワード> -local <ローカルディレクトリ> -remote <リモートディレクトリ>

# ファイル/ディレクトリをダウンロード
win-sshpass -h <ホスト> -p <パスワード> -d -remote <リモートパス> -local <ローカルパス>
```

### SCP スタイル

```bash
# ファイルをアップロード
win-sshpass -p <パスワード> scp <ローカルファイル> user@host:<リモートパス>
win-sshpass -p <パスワード> scp -P <ポート> <ローカルファイル> user@host:<リモートパス>

# ディレクトリをアップロード
win-sshpass -p <パスワード> scp -r <ローカルディレクトリ> user@host:<リモートパス>

# ファイル/ディレクトリをダウンロード
win-sshpass -p <パスワード> scp user@host:<リモートファイル> <ローカルパス>
```

### Rsync スタイル

```bash
# アップロード
win-sshpass -p <パスワード> rsync -avz <ローカルパス> user@host:<リモートパス>

# ダウンロード
win-sshpass -p <パスワード> rsync -avz user@host:<リモートパス> <ローカルパス>
```

## パラメータ

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `-p` | パスワード | `-p 'secret123'` |
| `-i` | 秘密鍵パス | `-i ~/.ssh/id_ed25519` |
| `-f` | パスワードファイル/設定ファイル | `-f pass.txt` |
| `-e` | 環境変数 SSHPASS からパスワード読み込み | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | ホストアドレス | `-h example.com` |
| `-u` | ユーザー名、デフォルト: root | `-u ubuntu` |
| `-P` | ポート、デフォルト: 22 | `-P 2222` |
| `-c` | 実行するコマンド | `-c 'ls -la'` |
| `-local` | ローカルパス（カンマまたはスペース区切り） | `-local "a.txt,b.txt"` |
| `-remote` | リモートパス（アップロード/ダウンロード） | `-remote /tmp/file.txt` |
| `-d` | ダウンロードモード | `-d` |
| `-k` | 厳密なホスト鍵検証を有効化 | `-k` |
| `-t` | 総操作タイムアウト（秒）、0 = 無制限 | `-t 30` |
| `-ct` | TCP 接続タイムアウト（秒）、デフォルト: 10 | `-ct 5` |
| `-retry` | 総接続試行回数（デフォルト：3） | `-retry 5` |
| `-resume` | 中断されたファイル転送をブレークポイントから再開 | `-resume` |
| `-proxy` | プロキシ URL（socks5/socks4/http/https） | `-proxy socks5://127.0.0.1:1080` |
| `-v` | バージョン表示 | `-v` |
| `-help` | ヘルプメッセージを表示 | `-help` |

### Keygen フラグ

| パラメータ | 説明 | デフォルト |
|-----------|------|------------|
| `-algo` | 鍵アルゴリズム（`ed25519` または `rsa`） | `ed25519` |
| `-comment` | 生成された公開鍵のコメント | `user@host` |
| `-out` | 秘密鍵の出力パス | `~/.ssh/id_ed25519` または `~/.ssh/id_rsa` |

## ハッシュと検証

SSH 接続なしでローカルファイルのハッシュを計算・検証：

```bash
# ハッシュを計算
win-sshpass hash md5 ./file.iso
win-sshpass hash sha256 ./file.iso

# 期待値と照合
win-sshpass verify sha256 d1dc38f6df... ./file.iso
# 出力: OK  (または: FAILED)
```

対応アルゴリズム: `md5`、`sha1`、`sha256`、`sha512`。

## 鍵生成

ssh-keygen なしで SSH 鍵ペアをローカルに生成：

```bash
# Ed25519鍵を生成（推奨 — 高速でより安全）
win-sshpass keygen

# RSA 4096ビット鍵を生成
win-sshpass keygen -algo rsa

# 出力パスを指定
win-sshpass keygen -out ~/.ssh/mykey

# コメントを追加
win-sshpass keygen -comment "my-laptop"
```

デフォルトでは、鍵は `~/.ssh/id_ed25519`（Ed25519）または `~/.ssh/id_rsa`（RSA）に保存されます。公開鍵ファイルには自動的に `.pub` 拡張子が付与されます。

**公開鍵をデプロイしてパスワード不要のログインを有効にする：**

```bash
# 公開鍵の内容を変数に読み込み、SSH経由でデプロイ
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'password' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

# その後、秘密鍵でログイン
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

## 設定ファイル形式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # オプション、パスワードの代わりに秘密鍵を使用
# timeout: 0              # オプション、総操作タイムアウト（秒）、0 = 無制限
# connect_timeout: 10     # オプション、TCP 接続タイムアウト（秒）
# strict_host_key: false  # オプション、厳密なホスト鍵検証を有効化
# proxy: socks5://user:pass@127.0.0.1:1080  # オプション、プロキシ URL（socks5/socks4/http/https）
```

使用方法：
```bash
win-sshpass -f server.config -c 'ls -la'
win-sshpass -f server.config 'ls -la'
```

## 完全な例

```bash
# 1. パスワード認証でコマンド実行
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 秘密鍵認証で sudo コマンド実行
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. ディレクトリ全体をサーバーにアップロード
win-sshpass -h server.com -p 'mypass' -local ./dist -remote //var/www/html

# 4. サーバーのログディレクトリをダウンロード
win-sshpass -h server.com -p 'mypass' -d -remote //var/log/nginx -local ./logs

# 5. SCP でファイルをアップロード
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 環境変数でパスワードを渡す（より安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'

# 7. 操作タイムアウト（30 秒後に自動中断）
win-sshpass -p 'mypass' -t 30 ssh user@server.com 'long-running-command'

# 8. 設定ファイル + 位置引数コマンド
win-sshpass -f server.config 'docker ps'

# 9. 中断されたアップロードを再開
win-sshpass -p 'mypass' -h server.com -local ./bigfile.iso -remote //data/bigfile.iso -resume

# 10. ファイルハッシュを計算
win-sshpass hash sha256 ./download.iso

# 11. ファイルの整合性を検証
win-sshpass verify sha256 d1dc38f6dfb1e4c8... ./download.iso

# 12. SSH鍵ペアを生成
win-sshpass keygen

# 13. カスタムパスとコメントで RSA 鍵を生成
win-sshpass keygen -algo rsa -out ~/.ssh/mykey -comment "my-server"

# 14. 生成した秘密鍵でログイン（公開鍵をサーバーにデプロイ後）
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

## プロキシ対応

プロキシサーバー経由で SSH 接続をトンネルできます。対応プロトコル：SOCKS5、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT。

```bash
# SOCKS5 プロキシ
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@host

# SOCKS5（認証付き）
win-sshpass -p 'pass' -proxy socks5://proxyuser:proxypass@127.0.0.1:1080 ssh user@host

# SOCKS4 プロキシ
win-sshpass -p 'pass' -proxy socks4://192.168.1.1:1080 ssh user@host

# HTTP CONNECT プロキシ
win-sshpass -p 'pass' -proxy http://proxy.local:8080 ssh user@host

# HTTPS CONNECT プロキシ（認証付き）
win-sshpass -p 'pass' -proxy https://user:pass@proxy.local:8443 ssh user@host

# プロキシ + ファイル転送
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 -h host -local ./file.txt -remote /tmp/file.txt

# プロキシ + SCP
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 scp ./app.jar user@host:/opt/app/

# 設定ファイルでプロキシを指定
# proxy: socks5://user:pass@127.0.0.1:1080
```

## Git Bash の注意事項

リモートパスは `//` で始めてパス変換を回避してください：
```bash
# 誤り: /tmp が Windows パスに変換される
win-sshpass ... -remote /tmp/file.txt

# 正しい: ダブルスラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## Go SDK としての利用

`win-sshpass` は再利用可能な Go ライブラリ（`package sshpass`）でもあります。インポートして、SSH/SFTP/Shell 機能を独自アプリに組み込めます：

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

	// NewClient がダイヤルし、すぐに使えるクライアントを返します。
	client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// コマンドを実行（出力はデフォルトで os.Stdout/os.Stderr に流れます）。
	if err := client.Exec("uname -a"); err != nil {
		log.Fatal(err)
	}

	// SFTP でファイルをアップロード。
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

### カスタマイズオプション

`NewClient` に渡す関数型オプションで動作を設定します：

| オプション | 用途 |
|-----------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | I/O ストリームのリダイレクト（デフォルト `os.Stdin`/`os.Stdout`/`os.Stderr`）。 |
| `WithProgress(fn)` | SFTP 転送中に `(description string, sent, total int64)` を受け取る `ProgressFunc` コールバックを設定。SDK 自身はレンダリングを行わず、進行の表示方法は呼び出し側に委ねられます。デフォルトは未設定（ヘッドレス環境向け）。 |
| `WithFileSelector(s)` | rz/sz ファイル転送フォールバックで使用する `FileSelector` を設定。SDK はデフォルト実装を提供せず、未設定時は rz/sz が stdin からパスを読み取ります。 |
| `WithSignalHandler()` | Ctrl+C ハンドラを登録して接続を閉じる。デフォルトでは登録されず、ライブラリはホストプロセスのシグナル処理に干渉しません。 |

SDK は意図的に **UI コードを一切内蔵していません**（プログレスバーやファイルダイアログなし）。
これらは CLI パッケージ（`cmd/sshpass/ui.go`）に実装され、progressbar ベースの
`ProgressFunc` と zenity ベースの `FileSelector` をクライアントに組み込んでいます。
ライブラリ利用者は独自に提供する必要があります。

SSH 接続をプロキシ経由でトンネルするには、`NewClient` を呼び出す前に `Config.ProxyURL` を設定します：

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.ProxyURL = "socks5://user:pass@127.0.0.1:1080" // または http://、https://、socks4://

client, err := sshpass.NewClient(cfg)
```

低レベルヘルパー関数も高度な用途向けにエクスポートされています：`Dial`、`NewConfig`、
`LoadConfig`、`LoadConfigOrPasswordFile`、`ParseSSHArgs`、`ParseSCPArgs`、`ParseRsyncArgs`、
`DetectCommandType`、`RunSCP`、`RunRsync`、`CleanRemotePath`、`SplitPaths`、
`ParseUserHostPath`、`ExitCodeFromError`。

## ビルド

```bash
# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass

# クロスコンパイル
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=windows GOARCH=amd64               go build -o win-sshpass.exe ./cmd/sshpass
GOOS=darwin  GOARCH=arm64               go build -o win-sshpass ./cmd/sshpass
```

## 依存関係

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3（CLI プログレスバーのみ）
- github.com/ncruces/zenity（CLI ファイルダイアログのみ）

