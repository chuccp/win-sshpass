# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 sshpass ツール。Linux の sshpass と同様の機能を提供します。

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
- x64 (amd64) および ARM64 アーキテクチャ対応

## ダウンロード

[GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から最新版をダウンロード：

| アーキテクチャ | Zip | MSI インストーラー |
|----------------|-----|-------------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. [Releases](https://github.com/chuccp/win-sshpass/releases) ページを開く
2. お使いのアーキテクチャ（x64 または ARM64）の zip または MSI インストーラーをダウンロード
3. MSI インストーラーを使用する場合：インストーラーを実行するだけで、インストール先がシステム PATH に自動的に追加されます

> **依存関係ゼロ**：`win-sshpass.exe` はスタンドアロンの実行ファイルです。OpenSSH やその他のソフトウェアのインストールは不要です。ダウンロードして PATH に配置すればすぐに使用できます。

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
| `-v` | バージョン表示 | `-v` |
| `-help` | ヘルプメッセージを表示 | `-help` |

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
```

## Git Bash の注意事項

リモートパスは `//` で始めてパス変換を回避してください：
```bash
# 誤り: /tmp が Windows パスに変換される
win-sshpass ... -remote /tmp/file.txt

# 正しい: ダブルスラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## ビルド

```bash
go build -o win-sshpass.exe .
```

## 依存関係

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3

## 関連プロジェクト

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — Go Web フレームワーク。依存性注入、ジェネリクスベースのデータアクセス、本番対応ユーティリティを統合。HTTP ルーティング、キャッシュ、ロギング、Redis などを内蔵。