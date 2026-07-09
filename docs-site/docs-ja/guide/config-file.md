# 設定ファイル

win-sshpass は設定ファイルをサポートし、複数サーバーの接続情報を管理できます。毎回長いコマンドライン引数を入力する必要がなくなります。

## 設定ファイルのフォーマット

設定ファイルはシンプルな `key: value` フォーマットを使用します：

```yaml
host: example.com
username: root
password: your_password
port: 22
```

### サポートされるフィールド

| フィールド | 説明 | デフォルト |
|-----------|------|----------|
| `host` | ホストアドレス（必須） | - |
| `username` / `user` | ユーザー名 | root |
| `password` | パスワード | - |
| `port` | ポート | 22 |
| `key` / `keypath` | 秘密鍵ファイルのパス | - |
| `timeout` | 操作タイムアウト（秒）、0 = 無制限 | 0 |
| `connect_timeout` | TCP 接続タイムアウト（秒） | 10 |
| `retry` / `retries` | 接続リトライ回数 | 3 |
| `strict_host_key` | 厳密なホストキー検証を有効化 | false |

### 例

```yaml
host: 192.168.1.100
username: ubuntu
password: mypassword
port: 22
# key: ~/.ssh/id_ed25519  # 秘密鍵認証を使用（パスワードの代替）
# timeout: 0              # 操作タイムアウト（秒）、0 = 無制限
# connect_timeout: 10     # TCP 接続タイムアウト（秒）
# retry: 3                # 接続リトライ回数
# strict_host_key: false  # 厳密なホストキー検証を有効化
```

## 設定ファイルの使用

```bash
# コマンド実行
win-sshpass -f server.config -c 'docker ps'

# またはコマンドを位置引数として渡す
win-sshpass -f server.config 'docker ps'

# インタラクティブシェルを開く
win-sshpass -f server.config

# 設定ファイル + SSH スタイル引数
win-sshpass -f server.config ssh user@host 'ls'
```

## パスワードファイル

ファイルが設定フォーマットではない場合（`host:` などのキーを含まない場合）、win-sshpass はパスワードファイル（1行テキスト）として扱います：

```bash
# pass.txt の内容：mypassword
win-sshpass -f pass.txt ssh user@host
```

## 設定の優先順位

設定は以下の優先順位でマージされます（高い方が低い方を上書き）：

1. **コマンドライン引数**（最高優先度）
2. **設定ファイル**
3. **デフォルト値**（最低優先度）

例：

```bash
# 設定ファイルでは port: 22 だが、コマンドラインで -P 2222 を指定
win-sshpass -f server.config -P 2222 ssh user@host
# ポート 2222 を使用
```

## 複数サーバーの管理

各サーバーごとに個別の設定ファイルを作成：

```
~/.ssh/
├── web-server.config
├── db-server.config
└── staging.config
```

```bash
# Web サーバー
win-sshpass -f ~/.ssh/web-server.config 'nginx -t'

# データベースサーバー
win-sshpass -f ~/.ssh/db-server.config 'systemctl status mysql'

# ステージング環境
win-sshpass -f ~/.ssh/staging.config 'docker ps'
```

## セキュリティのヒント

!!! warning "パスワードのセキュリティ"
    設定ファイルには平文のパスワードが含まれています。ファイル権限に注意してください：

    - 設定ファイルを安全なディレクトリ（例：`~/.ssh/`）に保存
    - 適切なファイル権限を設定（現在のユーザーのみ読み取り可能）
    - パスワードの代わりに秘密鍵認証の使用を検討

```bash
# Linux/macOS
chmod 600 server.config

# Windows（icacls を使用）
icacls server.config /inheritance:r /grant:r %USERNAME%:R
```

## 次のステップ

- [SSH 接続](ssh.md) - その他の認証方法
- [ベストプラクティス](../advanced/best-practices.md) - セキュリティと効率のヒント
