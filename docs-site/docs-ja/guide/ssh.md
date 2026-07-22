# SSH 接続

win-sshpass は複数の SSH 認証方法をサポートし、さまざまなシナリオに対応します。

## パスワード認証

### パスワードを直接指定

```bash
win-sshpass -p 'mypassword' ssh user@host
win-sshpass -p 'mypassword' ssh user@host 'whoami'
```

### ファイルからパスワードを読み取り

パスワードのみを含むテキストファイル（1行）を作成：

```bash
echo 'mypassword' > pass.txt
win-sshpass -f pass.txt ssh user@host
```

### 環境変数からパスワードを読み取り

```bash
export SSHPASS='mypassword'
win-sshpass -e ssh user@host
```

Windows CMD の場合：

```cmd
set SSHPASS=mypassword
win-sshpass -e ssh user@host
```

!!! tip "セキュリティのヒント"
    環境変数や設定ファイルを使用する方が、コマンドラインで直接パスワードを渡すよりも安全です。コマンド履歴にパスワードが残りません。

## 秘密鍵認証

```bash
# Ed25519 鍵を使用
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# RSA 鍵を使用
win-sshpass -i ~/.ssh/id_rsa ssh user@host

# リモートコマンドを実行
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'uname -a'
```

!!! note "注意"
    win-sshpass は暗号化された（パスフレーズで保護された）秘密鍵をサポートしていません。鍵がパスフレーズで保護されている場合は、先に復号化するか ssh-agent を使用してください。

## 鍵生成

win-sshpass には SSH 鍵ペア生成機能が組み込まれています。クライアント側の鍵ペア（秘密鍵 + 公開鍵）をローカルで生成できます。

```bash
# Ed25519 鍵を生成（推奨 — 高速でより安全）
win-sshpass keygen

# RSA 鍵を生成（4096 ビット）
win-sshpass keygen -algo rsa

# 出力パスを指定
win-sshpass keygen -out ~/.ssh/mykey

# 公開鍵のコメントを指定
win-sshpass keygen -comment "my-laptop"
```

デフォルトで `~/.ssh/id_ed25519`（Ed25519）または `~/.ssh/id_rsa`（RSA）に保存されます。公開鍵ファイルには自動的に `.pub` 接尾辞が付きます。

生成後、公開鍵をサーバーにデプロイすればパスワードレスログインが可能になります（下記参照）。

### 公開鍵のデプロイ

生成後、公開鍵をサーバーにデプロイしてパスワードレスログインを有効にします：

```bash
# 公開鍵の内容を変数に読み込み、SSH経由でデプロイ
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'mypassword' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

シェルが `$PUBKEY` をローカルで展開してから win-sshpass にコマンドを渡すため、実際の公開鍵文字列がリモートコマンドに埋め込まれます。この方法により、引用符や標準入力転送の問題を回避できます。

デプロイ完了後、秘密鍵でパスワードなしログインが可能になります：

```bash
# パスワードレスログイン
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# パスワードなしコマンド実行
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'whoami'

# パスワードなしファイル転送
win-sshpass -i ~/.ssh/id_ed25519 scp file.txt user@host:/tmp/
```

!!! tip "authorized_keys の権限"
    サーバーの `~/.ssh` ディレクトリは権限 700、`~/.ssh/authorized_keys` は権限 600 にする必要があります。権限が正しくないと鍵認証が失敗します。

## ユーザーとポートの指定

```bash
# ユーザー名を指定（デフォルト: root）
win-sshpass -p 'pass' ssh ubuntu@host

# ポートを指定（デフォルト: 22）
win-sshpass -p 'pass' ssh -p 2222 user@host

# -u と -P フラグを使用
win-sshpass -p 'pass' -h host -u ubuntu -P 2222
```

## リモートコマンドの実行

```bash
# 単一コマンド
win-sshpass -p 'pass' ssh user@host 'ls -la'

# 複数コマンド
win-sshpass -p 'pass' ssh user@host 'cd /var/log && ls -la'

# -c フラグを使用
win-sshpass -p 'pass' -h host -c 'docker ps'
```

## 接続タイムアウトとリトライ

```bash
# TCP 接続タイムアウト（デフォルト: 10秒）
win-sshpass -p 'pass' -ct 5 ssh user@host

# 操作タイムアウト（デフォルト: 無制限）
win-sshpass -p 'pass' -t 30 ssh user@host 'long-command'

# リトライ回数（デフォルト: 3回）
win-sshpass -p 'pass' -retry 5 ssh user@host
```

タイムアウトメカニズム：

- **TCP 接続タイムアウト**（`-ct`）：TCP 接続確立のタイムアウト時間
- **操作タイムアウト**（`-t`）：操作全体のタイムアウト時間。データ転送中にタイマーが自動リセット
- **リトライ**（`-retry`）：接続失敗時のリトライ回数。指数バックオフ（2s、4s、8s、16s、最大30s）

!!! info "認証失敗時はリトライしない"
    認証失敗（パスワード間違い、無効な鍵）の場合はリトライせず、エラーを即座に返します。

## ホストキー検証

デフォルトでは、win-sshpass はホストキーを検証しません（`StrictHostKeyChecking=no` と同等）。

厳密なホストキー検証を有効にする：

```bash
# -k フラグを使用
win-sshpass -p 'pass' -k ssh user@host

# または設定ファイルで設定
# strict_host_key: true
```

有効にすると、`~/.ssh/known_hosts` ファイルを使用して検証します。ホストが known_hosts にない場合、接続は拒否されます。

## IPv6 サポート

win-sshpass は IPv6 アドレスをサポートしています：

```bash
win-sshpass -p 'pass' ssh user@2001:db8::1
win-sshpass -p 'pass' ssh user@[2001:db8::1]
```

## プロキシ対応

プロキシサーバー経由で SSH 接続をトンネル：

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

!!! info "対応プロキシプロトコル"
    SOCKS5（オプションのユーザー名/パスワード認証）、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT プロキシに対応しています。

## ファイルハッシュと検証

SSH 接続なしでローカルファイルのハッシュを計算・検証：

```bash
# ハッシュを計算
win-sshpass hash md5 ./download.iso
win-sshpass hash sha1 ./download.iso
win-sshpass hash sha256 ./download.iso
win-sshpass hash sha512 ./download.iso

# ファイルの整合性を検証
win-sshpass verify sha256 d1dc38f6dfb1e4c8e7a1b2c3d4e5f6a7b8c9d0e1f2 ./download.iso
# 出力: OK

win-sshpass verify sha256 wronghash123... ./download.iso
# 出力: FAILED
```

対応アルゴリズム: `md5`、`sha1`、`sha256`、`sha512`。

## 次のステップ

- [インタラクティブシェル](shell.md) - コマンド未指定時のインタラクティブモード
- [ファイル転送](file-transfer.md) - SFTP アップロード/ダウンロード
- [設定ファイル](config-file.md) - 複数サーバーの管理
