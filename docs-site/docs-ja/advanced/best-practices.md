# ベストプラクティス

## セキュリティのヒント

### 1. コマンドラインでパスワードを渡すのを避ける

```bash
# 非推奨：パスワードがコマンド履歴に残る
win-sshpass -p 'mypassword' ssh user@host

# 推奨：環境変数を使用
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推奨：パスワードファイルを使用
win-sshpass -f pass.txt ssh user@host

# 推奨：設定ファイルを使用
win-sshpass -f server.config ssh user@host
```

### 2. 秘密鍵認証を使用する

秘密鍵認証はパスワード認証よりも安全です：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 設定ファイルの権限を保護する

```bash
# Linux/macOS
chmod 600 server.config

# Windows（PowerShell）
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. ホストキー検証を有効にする

本番環境では、厳密なホストキー検証を有効にすることをお勧めします：

```bash
win-sshpass -k -f server.config ssh user@host
```

または設定ファイルで：

```yaml
strict_host_key: true
```

## 効率化のヒント

### 1. 設定ファイルでサーバーを管理する

よく使うサーバーの設定ファイルを作成し、パラメータの繰り返し入力を避ける：

```bash
# ~/.ssh/prod-web.config
host: web.example.com
username: deploy
key: ~/.ssh/id_ed25519

# ~/.ssh/prod-db.config
host: db.example.com
username: admin
key: ~/.ssh/id_ed25519
```

### 2. バッチ操作

シェルスクリプトと組み合わせてバッチ操作を実行：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. SSH スタイル構文を使用する

SSH に慣れたユーザーには、より自然な構文が利用可能です：

```bash
# 標準 SSH 構文
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 構文
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 構文
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 適切なタイムアウトを設定する

```bash
# 短時間のコマンド：短いタイムアウト
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 長時間の操作：長いタイムアウトまたはタイムアウトなし
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

## トラブルシューティング

### 接続失敗

```bash
# リトライ回数を増やす
win-sshpass -p 'pass' -retry 5 ssh user@host

# 接続タイムアウトを増やす
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 認証失敗

- パスワードが正しいか確認
- 秘密鍵のパスが正しいか確認
- リモートサーバーがパスワード/鍵認証を許可しているか確認
- 注意：暗号化された秘密鍵はサポートされていません

### Git Bash のパス問題

```bash
# 間違い：/tmp は Git Bash によって変換されます
win-sshpass ... -remote /tmp/file.txt

# 正しい：二重スラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## 次のステップ

- [Go SDK](sdk.md) - プログラムから使用
- [変更履歴](../changelog.md) - バージョン履歴
