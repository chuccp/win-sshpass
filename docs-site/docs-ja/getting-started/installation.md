# インストール

## システム要件

- **OS**: Windows 10/11（x64 または ARM64）
- **依存関係なし**: OpenSSH や他のソフトウェアをインストールする必要はありません

## インストール方法

### 方法 1: Scoop（推奨）

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方法 2: リリースをダウンロード

[GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から最新バージョンをダウンロード：

| アーキテクチャ | Zip | MSI インストーラー |
|---------------|-----|-------------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. [Releases](https://github.com/chuccp/win-sshpass/releases) ページに移動
2. アーキテクチャに合った zip または MSI をダウンロード
3. MSI の場合：インストーラーを実行すると、インストールディレクトリが自動的にシステム PATH に追加されます

### 方法 3: ソースからビルド

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass
go build -o win-sshpass.exe ./cmd/sshpass
```

## インストールの確認

```bash
win-sshpass -v
# 出力: sshpass version v0.3.2 (Windows)
```

## 依存関係

win-sshpass は外部ランタイム依存関係のない独立した実行ファイルです。ビルド時に使用される Go 依存関係：

| 依存関係 | 用途 |
|---------|------|
| golang.org/x/crypto/ssh | SSH プロトコル実装 |
| github.com/pkg/sftp | SFTP ファイル転送 |
| github.com/schollz/progressbar/v3 | CLI プログレスバー（CLI のみ） |
| github.com/ncruces/zenity | ファイルダイアログ（CLI のみ） |

## 次のステップ

- [クイックスタート](quick-start.md) - 最初の SSH 接続
