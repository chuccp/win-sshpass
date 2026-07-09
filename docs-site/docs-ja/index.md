# win-sshpass ユーザーガイド

> Windows 版 sshpass：パスワード/キー SSH ログイン、インタラクティブシェル、SFTP ファイル転送、SCP/Rsync スタイル転送、そして再利用可能な Go SDK。

## win-sshpass とは？

win-sshpass は Linux の sshpass ツールの Windows 実装です。OpenSSH や他のソフトウェアをインストールする必要がない独立した実行ファイルです。ダウンロードしてすぐに使えます。

**コマンドラインツール**と **Go SDK** の2つの使用方法をサポートしています：

- **パスワード/キー認証**：パスワード、秘密鍵、環境変数、設定ファイルなど複数の認証方法をサポート。
- **インタラクティブシェル**：ローミングターミナルモード、vim、top、htop などのフルスクリーンアプリ、動的ターミナルリサイズをサポート。
- **SFTP ファイル転送**：ファイルとディレクトリのアップロード/ダウンロード、プログレスバー付き。
- **SCP/Rsync スタイル**：scp および rsync コマンド構文と互換性あり。
- **再利用可能な Go SDK**：Go ライブラリとしてインポートし、自分のアプリケーションに SSH/SFTP/Shell 機能を組み込み可能。I/O ストリームとコールバックの注入をサポート。

## 30秒クイックスタート

```bash
# Scoop でインストール
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# パスワードでログインしてコマンド実行
win-sshpass -p 'password' ssh user@example.com 'whoami'

# ファイルをアップロード
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# インタラクティブシェルを開く
win-sshpass -p 'password' ssh user@host
```

## 主な機能

### 1. 複数の認証方法

**パスワード認証**：直接指定、ファイルから、環境変数から。

```bash
# パスワードを直接指定
win-sshpass -p 'secret' ssh user@host

# ファイルから読み取り
win-sshpass -f pass.txt ssh user@host

# 環境変数から読み取り
SSHPASS='secret' win-sshpass -e ssh user@host
```

**秘密鍵認証**：Ed25519、RSA などの鍵形式をサポート。

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

**設定ファイル**：複数サーバーの接続情報を管理。

```bash
win-sshpass -f server.config -c 'docker ps'
```

### 2. インタラクティブシェル（ローミングターミナルモード）

コマンドを指定しない場合、win-sshpass はインタラクティブシェルを開きます：

- **正しいエコー** — 入力した文字が正しく表示されます（二重エコーなし）
- **Ctrl+C / Ctrl+Z** — シグナルがリモートプロセスに転送されます
- **フルスクリーンアプリ** — vim、top、htop、nano が正しく動作します
- **動的ターミナルリサイズ** — リモートターミナルがローカルウィンドウサイズに自動調整
- **Tab 補完** — リモートシェルの Tab 補完が正常に動作

```bash
win-sshpass -p 'password' ssh user@host
```

### 3. ファイル転送

**SFTP**：ファイルとディレクトリのアップロード/ダウンロード、プログレスバー付き。

```bash
# ファイルをアップロード
win-sshpass -h host -p 'pass' -local file.txt -remote /tmp/file.txt

# 複数ファイルをアップロード
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# ディレクトリをダウンロード
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

**SCP スタイル**：scp コマンド構文と互換性あり。

```bash
win-sshpass -p 'pass' scp ./app.jar user@server:/opt/app/
win-sshpass -p 'pass' scp -r ./dist user@server:/var/www/html
```

**Rsync スタイル**：rsync コマンド構文と互換性あり。

```bash
win-sshpass -p 'pass' rsync -avz ./ user@server:/backup/
```

**シェル内 rz/sz**：インタラクティブシェルで直接 rz/sz コマンドを使用。

```bash
# リモートシェルで：
rz                              # ファイルをアップロード（ファイルピッカーを開く）
sz /remote/path/to/file         # ファイルをダウンロード
```

### 4. 再利用可能な Go SDK

win-sshpass は Go ライブラリ（`package sshpass`）でもあり、自分のアプリケーションに組み込むことができます：

```go
import sshpass "github.com/chuccp/win-sshpass"

cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
defer client.Close()

// コマンド実行
client.Exec("uname -a")

// SFTP 転送
sftp, _ := client.SFTP()
sftp.Upload("./local.txt", "/tmp/remote.txt")
```

SDK は **UI コードを含みません**（プログレスバー、ファイルダイアログなし）。動作は関数オプションで設定します：

| オプション | 用途 |
|-----------|------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | I/O ストリームのリダイレクト |
| `WithProgress(fn)` | 転送プログレスコールバックの設定 |
| `WithFileSelector(s)` | rz/sz ファイルセレクターの設定 |
| `WithSignalHandler()` | Ctrl+C シグナルハンドラーの登録 |

## パラメータ一覧

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| `-p` | パスワード | `-p 'secret123'` |
| `-i` | 秘密鍵のパス | `-i ~/.ssh/id_ed25519` |
| `-f` | パスワードファイル / 設定ファイル | `-f pass.txt` |
| `-e` | SSHPASS 環境変数からパスワードを読み取り | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | ホストアドレス | `-h example.com` |
| `-u` | ユーザー名（デフォルト: root） | `-u ubuntu` |
| `-P` | ポート（デフォルト: 22） | `-P 2222` |
| `-c` | 実行するコマンド | `-c 'ls -la'` |
| `-local` | ローカルパス（カンマまたはスペース区切り） | `-local "a.txt,b.txt"` |
| `-remote` | リモートパス | `-remote /tmp/file.txt` |
| `-d` | ダウンロードモード | `-d` |
| `-k` | 厳密なホストキー検証を有効化 | `-k` |
| `-t` | 操作タイムアウト（秒）、0 = 無制限 | `-t 30` |
| `-ct` | TCP 接続タイムアウト（秒）、デフォルト: 10 | `-ct 5` |
| `-retry` | 接続リトライ回数（デフォルト: 3） | `-retry 5` |
| `-v` | バージョン表示 | `-v` |
| `-help` | ヘルプ表示 | `-help` |

## クイックリンク

### はじめに

- [インストール](getting-started/installation.md) - ダウンロードとインストール方法
- [クイックスタート](getting-started/quick-start.md) - 最初の接続

### ユーザーガイド

- [SSH 接続](guide/ssh.md) - パスワード、キー、環境変数認証
- [ファイル転送](guide/file-transfer.md) - SFTP アップロード/ダウンロード
- [インタラクティブシェル](guide/shell.md) - ローミングターミナルモードと rz/sz
- [SCP と Rsync](guide/scp-rsync.md) - scp/rsync 互換構文
- [設定ファイル](guide/config-file.md) - 複数サーバーの管理

### 上級編とリファレンス

- [Go SDK](advanced/sdk.md) - Go ライブラリとして使用
- [ベストプラクティス](advanced/best-practices.md) - セキュリティと効率のヒント
- [変更履歴](changelog.md)

## コミュニティ

- [GitHub](https://github.com/chuccp/win-sshpass)
- [課題トラッカー](https://github.com/chuccp/win-sshpass/issues)
