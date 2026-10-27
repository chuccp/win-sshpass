# win-sshpass

[![Release](https://badgen.net/github/release/chuccp/win-sshpass)](https://github.com/chuccp/win-sshpass/releases)
[![License](https://badgen.net/badge/License/Apache%202.0/blue)](LICENSE)
[![Go Version](https://badgen.net/badge/Go/1.26.0/00ADD8?icon=go)](go.mod)
[![Downloads](https://img.shields.io/github/downloads/chuccp/win-sshpass/total)](https://github.com/chuccp/win-sshpass/releases)

[English](README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md)

`win-sshpass` は Windows、Linux、macOS 対応の SSH クライアント兼自動化ツールです。
パスワード/鍵認証、SFTP と SCP/Rsync 形式の転送、ポート転送、ブラウザー SSH
ワークスペース、MCP ツール、再利用可能な Go SDK を提供します。

## 主な機能

- パスワード、秘密鍵、ssh-agent、明示的な空パスワードで SSH 接続。
- リモートコマンド、JSON、再試行、タイムアウト、プロキシ、`known_hosts`。
- 対話型ターミナル、バックグラウンド実行 `--bg`、`rz`/`sz` 転送。
- SFTP アップロード/ダウンロード、ディレクトリ再帰、複数ファイル、再開。
- ローカル/リモートポート転送、SOCKS4/SOCKS5/HTTP/HTTPS プロキシ。
- 全画面 Web SSH：SFTP ファイルマネージャーと xterm.js の実 PTY。
- ドラッグ＆ドロップ、アップロード進捗率、可変分割バー、ライト/ダークテーマ。
- SSH 実行と SFTP 操作用の MCP ツール。
- 自己更新、SSH 鍵生成、ハッシュ、検証。

## インストール

```powershell
winget install chuccp.win-sshpass
```

```powershell
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

または [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から
アーカイブ、MSI、PKG をダウンロードできます。外部ランタイムは不要です。

## クイックスタート

```bash
# コマンドを実行
win-sshpass -p 'password' ssh root@example.com 'whoami'

# 対話型シェルを開く
win-sshpass -p 'password' ssh root@example.com

# アップロードとダウンロード
win-sshpass -h example.com -p 'password' -local ./app.jar -remote /tmp/app.jar
win-sshpass -h example.com -p 'password' -d -remote /tmp/app.log -local ./app.log

# Web SSH コンソール（既定: http://127.0.0.1:6326）
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p 'password'
```

コマンドには `-h` または `-f` で SSH ホストを指定する必要があります。
<http://127.0.0.1:6326> を開くと自動接続してワークスペースを直接表示します。

## Web SSH ワークスペース

- 左側：SFTP ディレクトリ閲覧、アップロード/ダウンロード、ディレクトリ作成、削除、ドラッグ＆ドロップ。
- 右側：xterm.js で表示する実際の SSH PTY。vim/top と Tab 補完に対応。
- 分割バー：ドラッグまたは方向キーで幅を調整し、設定を保存。
- テーマ：既定はシステム設定。ライト/ダークも選択可能。
- アップロード進捗：ファイル名、順序、バー、全体パーセント。

`win-sshpass -web` だけではエラー終了します。コマンドでホスト、必要に応じて
ユーザーと認証情報を指定するか、設定ファイルを使用してください：

```bash
win-sshpass -f server.config -web
```

待受アドレスの変更：

```bash
win-sshpass -web -h example.com -u root -p 'password' -web-addr 127.0.0.1:6326
```

既定ではループバックのみをリッスンします。

## MCP

Web サーバーは既定で MCP Streamable HTTP を公開します：

```text
http://127.0.0.1:6326/mcp
```

Codex 設定例：

```toml
[mcp_servers.win-sshpass]
url = "http://127.0.0.1:6326/mcp"
```

利用できるツール：

- `ssh_exec`
- `sftp_list`
- `sftp_read`
- `sftp_write`
- `sftp_mkdir`
- `sftp_delete`

MCP は `-web` 起動時の SSH 接続を使用し、`Mcp-Session-Id` でセッションを管理します。

## よく使うコマンド

| 操作 | コマンド |
|---|---|
| 秘密鍵 | `win-sshpass -i ~/.ssh/id_ed25519 ssh user@host` |
| 環境変数パスワード | `SSHPASS='pass' win-sshpass -e ssh user@host` |
| パスワードファイル | `win-sshpass -f pass.txt ssh user@host` |
| 設定ファイル | `win-sshpass -f server.config 'docker ps'` |
| ポート指定 | `win-sshpass -p 'pass' ssh -p 2222 user@host` |
| 操作タイムアウト | `win-sshpass -t 30 -p 'pass' ssh user@host 'long-task'` |
| SOCKS5 プロキシ | `win-sshpass -proxy socks5://127.0.0.1:1080 ssh user@host` |
| SCP アップロード | `win-sshpass -p 'pass' scp -P 2222 file.txt user@host:/tmp/` |
| Rsync アップロード | `win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/` |
| 鍵生成 | `win-sshpass keygen -algo ed25519` |
| ハッシュ / 検証 | `win-sshpass hash sha256 file.iso` |
| 更新確認 | `win-sshpass update -check` |

## 設定ファイル

```yaml
host: 192.168.1.10
username: root
password: secret
port: 22
timeout: 0
strict_host_key: false
```

```bash
win-sshpass -f server.config 'docker ps'
```

詳細は[設定ファイルガイド](https://chuccp.github.io/win-sshpass/ja/guide/config-file/)を参照してください。

## Go SDK

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

if err := client.Exec("uname -a"); err != nil {
    log.Fatal(err)
}
```

```bash
go get github.com/chuccp/win-sshpass
```

[Go SDK ガイド](https://chuccp.github.io/win-sshpass/ja/advanced/sdk/)も参照してください。

## オンラインドキュメント

- [English](https://chuccp.github.io/win-sshpass/en/)
- [简体中文](https://chuccp.github.io/win-sshpass/zh/)
- [繁體中文](https://chuccp.github.io/win-sshpass/tw/)
- [日本語](https://chuccp.github.io/win-sshpass/ja/)
- [ドキュメントソース](https://github.com/chuccp/win-sshpass/tree/main/docs-site)

## 開発

```bash
go test -race ./...
go vet ./...
go build -o win-sshpass.exe ./cmd/sshpass
```

クロスコンパイル：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
```

## License

Apache-2.0。[LICENSE](LICENSE) を参照してください。
