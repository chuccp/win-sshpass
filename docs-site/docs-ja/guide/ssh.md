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

## 次のステップ

- [インタラクティブシェル](shell.md) - コマンド未指定時のインタラクティブモード
- [ファイル転送](file-transfer.md) - SFTP アップロード/ダウンロード
- [設定ファイル](config-file.md) - 複数サーバーの管理
