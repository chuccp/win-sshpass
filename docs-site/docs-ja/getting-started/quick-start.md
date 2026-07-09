# クイックスタート

このガイドでは、数分で win-sshpass を使い始めることができます。

## 5秒：リモートコマンドの実行

```bash
# パスワードでログインしてコマンド実行
win-sshpass -p 'mypassword' ssh root@192.168.1.100 'whoami'
# → root
```

## 30秒：ファイル転送

```bash
# ファイルをアップロード
win-sshpass -h 192.168.1.100 -p 'mypassword' -local ./app.jar -remote /opt/app/

# ファイルをダウンロード
win-sshpass -h 192.168.1.100 -p 'mypassword' -d -remote /var/log/app.log -local ./logs/
```

## 1分：インタラクティブシェル

```bash
# インタラクティブシェルを開く（コマンドを指定しない）
win-sshpass -p 'mypassword' ssh root@192.168.1.100
```

接続後：

- コマンドを正しく入力でき、エコーが正常に動作
- `vim`、`top`、`htop` などのフルスクリーンアプリが使用可能
- `Ctrl+C` で現在のコマンドを中断
- ターミナルウィンドウのサイズ変更時に、リモートターミナルが自動調整

## 3分：設定ファイルの使用

`server.config` を作成：

```yaml
host: 192.168.1.100
username: root
password: mypassword
port: 22
```

設定ファイルを使用：

```bash
# コマンド実行
win-sshpass -f server.config -c 'docker ps'

# またはコマンドを位置引数として渡す
win-sshpass -f server.config 'docker ps'

# インタラクティブシェルを開く
win-sshpass -f server.config
```

## 5分：SCP/Rsync スタイル転送

```bash
# SCP アップロード
win-sshpass -p 'mypassword' scp ./app.jar user@server:/opt/app/

# SCP ディレクトリアップロード
win-sshpass -p 'mypassword' scp -r ./dist user@server:/var/www/html

# Rsync アップロード
win-sshpass -p 'mypassword' rsync -avz ./ user@server:/backup/
```

## よく使うコマンド集

```bash
# 秘密鍵でログイン
win-sshpass -i ~/.ssh/id_ed25519 ssh user@server

# 環境変数でパスワードを渡す（より安全）
export SSHPASS='mypassword'
win-sshpass -e ssh user@server

# ポート指定
win-sshpass -p 'pass' ssh -p 2222 user@server

# 操作タイムアウト（30秒）
win-sshpass -p 'pass' -t 30 ssh user@server 'long-running-command'

# 複数ファイルのアップロード
win-sshpass -h server -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/
```

## 次のステップ

- [SSH 接続](../guide/ssh.md) - 認証方法の詳細
- [ファイル転送](../guide/file-transfer.md) - SFTP の完全な使用方法
- [設定ファイル](../guide/config-file.md) - 複数サーバーの管理
