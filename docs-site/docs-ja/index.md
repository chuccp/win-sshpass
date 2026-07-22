---
hide:
  - navigation
  - toc
---

# win-sshpass

> クロスプラットフォームの sshpass 実装（Windows、Linux、macOS）：パスワード/キー SSH ログイン、インタラクティブシェル、SFTP/SCP/Rsync ファイル転送、SOCKS5/SOCKS4/HTTP プロキシトンネル、ブレークポイントレジューム、ファイルハッシュ検証、鍵生成、再利用可能な Go SDK。

[クイックスタート](getting-started/quick-start.md){ .md-button .md-button--primary }
[インストール](getting-started/installation.md){ .md-button }
[ソース :simple-github:](https://github.com/chuccp/win-sshpass){ .md-button }

---

## :material-star: 機能

- **:material-console: SSH とシェル** — パスワードまたは秘密鍵による認証、リモートコマンド実行、raw ターミナルモードによるインタラクティブシェル（正しいエコー、Ctrl+C、vim/top 対応、動的リサイズ）。
- **:material-file-sync: ファイル転送** — プログレスバー付き SFTP アップロード/ダウンロード、SCP スタイルおよび Rsync スタイルの転送構文、複数ファイル対応、再帰的ディレクトリアップロード。
- **:material-cloud-download: シェル rz/sz** — インタラクティブシェルで `rz`/`sz` コマンドを直接使用可能 — リモートサーバーにソフトウェアのインストールは不要。転送は SFTP 経由。
- **:material-shield-key: 鍵生成** — 組み込みの SSH 鍵ペア生成（Ed25519 および RSA）、ssh-keygen 不要。SSH 経由で公開鍵を配布し、パスワードなしログインを実現。
- **:material-lan-connect: プロキシトンネル** — SOCKS5（認証付き）、SOCKS4、SOCKS4A、HTTP CONNECT、HTTPS CONNECT プロキシ経由で SSH 接続をトンネル。
- **:material-reload: ブレークポイントレジューム** — 中断された SFTP アップロード/ダウンロードを `-resume` フラグで途中から再開。
- **:material-fingerprint: ファイルハッシュ検証** — ローカルファイルのチェックサム（MD5、SHA-1、SHA-256、SHA-512）を計算・検証 — SSH 接続不要。
- **:material-package-variant: 再利用可能な Go SDK** — `package sshpass` をインポートして、SSH/SFTP/シェルを独自のアプリに組み込み可能。注入可能な I/O、プログレスコールバック、UI 依存ゼロ。

---

## :material-rocket-launch: 30 秒クイックスタート

```bash
# WinGet でインストール
winget install chuccp.win-sshpass

# または Scoop でインストール
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# パスワードでログインしてコマンドを実行
win-sshpass -p 'password' ssh user@example.com 'whoami'

# ファイルをアップロード
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# インタラクティブシェルを開く
win-sshpass -p 'password' ssh user@host

# SSH キーペアを生成
win-sshpass keygen
```

---

## :material-compass: クイックナビゲーション

| | | |
|---|---|---|
| [:material-download: インストール](getting-started/installation.md) | [:material-rocket-launch: クイックスタート](getting-started/quick-start.md) | [:material-console: SSH 接続](guide/ssh.md) |
| [:material-file-sync: ファイル転送](guide/file-transfer.md) | [:material-monitor: インタラクティブシェル](guide/shell.md) | [:material-folder-multiple: SCP と Rsync](guide/scp-rsync.md) |
| [:material-cog: 設定ファイル](guide/config-file.md) | [:material-code-braces: Go SDK](advanced/sdk.md) | [:material-security: ベストプラクティス](advanced/best-practices.md) |
| [:material-history: 変更履歴](changelog.md) | | |

---

## :material-layers: 依存関係

| 依存パッケージ | 用途 |
|---|---|
| `golang.org/x/crypto/ssh` | SSH プロトコル実装 |
| `github.com/pkg/sftp` | SFTP ファイル転送 |
| `github.com/schollz/progressbar/v3` | CLI プログレスバー（CLI のみ） |
| `github.com/ncruces/zenity` | rz/sz 用ファイルダイアログ（CLI のみ、オプション） |

win-sshpass は**単一の実行ファイル**です — 外部ランタイム依存はありません。ダウンロードして実行するだけです。

---

## :simple-github: コミュニティ

- [GitHub リポジトリ](https://github.com/chuccp/win-sshpass)
- [Issue トラッカー](https://github.com/chuccp/win-sshpass/issues)
- [リリース](https://github.com/chuccp/win-sshpass/releases)
- [変更履歴](changelog.md)
