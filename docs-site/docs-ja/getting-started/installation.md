# インストール

## システム要件

- **OS**: Windows 10/11（x64、ARM64）または Linux（amd64、arm64）
- **依存関係なし**: OpenSSH や他のソフトウェアをインストールする必要はありません

## インストール方法

### 方法 1: WinGet（推奨）

```bash
winget install chuccp.win-sshpass
```

### 方法 2: Scoop

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方法 3: リリースをダウンロード

[GitHub Releases](https://github.com/chuccp/win-sshpass/releases) から最新バージョンをダウンロード：

**Windows**

| アーキテクチャ | Zip | MSI インストーラー |
|---------------|-----|-------------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

**Linux**

| アーキテクチャ | Tarball |
|---------------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

1. [Releases](https://github.com/chuccp/win-sshpass/releases) ページに移動
2. プラットフォームとアーキテクチャに合ったパッケージをダウンロード
3. **Windows MSI**：インストーラーを実行すると、システム PATH に自動追加
4. **Windows Zip / Linux tar.gz**：展開してバイナリを PATH に配置

### 方法 4: ソースからビルド

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass

# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass
```

## インストールの確認

```bash
win-sshpass -v
# 出力: win-sshpass version v0.3.2 (windows/amd64)
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
