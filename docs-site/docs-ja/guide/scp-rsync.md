# SCP と Rsync

win-sshpass は標準の scp および rsync コマンド構文と互換性があり、ファイル転送は内部的に SFTP を使用して実装されています。

## SCP スタイル転送

### 基本構文

```bash
win-sshpass -p <password> scp [オプション] <ソース> <ターゲット>
```

### ファイルのアップロード

```bash
# 単一ファイルのアップロード
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# 特定のファイル名でアップロード
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/newname.txt

# ディレクトリのアップロード（-r で再帰）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html
```

### ファイルのダウンロード

```bash
# ファイルのダウンロード
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# ディレクトリのダウンロード
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

### ポートの指定

scp はポートに大文字の `-P` を使用します（ssh の小文字 `-p` とは異なります）：

```bash
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/
```

### サポートされるオプション

| オプション | 説明 |
|-----------|------|
| `-r` | ディレクトリの再帰的コピー |
| `-P <port>` | ポートの指定 |
| `-i <key>` | 秘密鍵の指定 |
| `-q` | 静寂モード |
| `-C` | 圧縮（SFTP で処理済み） |
| `-v` | 詳細出力 |

## Rsync スタイル転送

### 基本構文

```bash
win-sshpass -p <password> rsync [オプション] <ソース> <ターゲット>
```

### アップロード

```bash
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### ダウンロード

```bash
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/
```

### ポートの指定

rsync は `--port=` を使用してポートを指定します：

```bash
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

### サポートされるオプション

| オプション | 説明 |
|-----------|------|
| `-a` | アーカイブモード |
| `-v` | 詳細出力 |
| `-z` | 転送時の圧縮 |
| `--port=N` | ポートの指定 |
| `-e ssh` | リモートシェルの指定（無視されます） |

## SCP vs Rsync vs SFTP

| 方法 | 構文 | 最適な用途 |
|------|------|-----------|
| SCP | 標準 scp 構文 | シンプルなファイルコピー |
| Rsync | 標準 rsync 構文 | 差分同期（注意：現在の実装は全量転送） |
| SFTP | `-local` / `-remote` フラグ | 柔軟なファイル転送、複数ファイル対応 |

!!! note "注意"
    win-sshpass の rsync 実装は内部的に SFTP を使用しており、rsync の差分同期アルゴリズムをサポートしていません。真の差分同期が必要な場合は、リモートサーバーに rsync をインストールし、SSH 経由で直接使用してください。

## 次のステップ

- [ファイル転送](file-transfer.md) - 直接 SFTP 転送の詳細
- [設定ファイル](config-file.md) - 設定ファイルでコマンドを簡素化
