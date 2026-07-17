# ファイル転送

win-sshpass は複数のファイル転送方法を提供します：直接 SFTP、SCP スタイル、Rsync スタイル、インタラクティブシェルでの rz/sz。

## 直接 SFTP 転送

### ファイルのアップロード

```bash
# 単一ファイルのアップロード
win-sshpass -h host -p 'pass' -local ./file.txt -remote /tmp/file.txt

# 複数ファイルのアップロード（カンマ区切り）
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 複数ファイルのアップロード（スペース区切り、シンプルなパスのみ）
win-sshpass -h host -p 'pass' -local "a.txt b.txt c.txt" -remote /tmp/

# ディレクトリのアップロード（自動再帰）
win-sshpass -h host -p 'pass' -local ./dist -remote /var/www/html
```

### ファイルのダウンロード

```bash
# ファイルのダウンロード
win-sshpass -h host -p 'pass' -d -remote /tmp/file.txt -local ./file.txt

# ディレクトリのダウンロード
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

### プログレスバー

SFTP 転送時には自動的にプログレスバーが表示されます：

```
Uploading app.jar  45% |████████████         |  45MB/100MB  10MB/s
```

## SCP スタイル転送

標準の scp コマンド構文と互換性があります：

```bash
# ファイルのアップロード
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# ディレクトリのアップロード（-r で再帰）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html

# ポートの指定（-P、大文字に注意）
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/

# ファイルのダウンロード
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# ディレクトリのダウンロード
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

## Rsync スタイル転送

rsync コマンド構文と互換性があります：

```bash
# アップロード
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/

# ダウンロード
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/

# ポートの指定
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

## シェル内 rz/sz 転送

インタラクティブシェルでは、`rz` と `sz` コマンドでファイル転送が可能です — **リモートサーバーにソフトウェアをインストールする必要はありません**。

```bash
# まずインタラクティブシェルを開く
win-sshpass -p 'pass' ssh user@host

# リモートシェルで：
rz                              # ファイルをアップロード（ファイルピッカーを開く）
rz /local/path/to/file          # 特定のローカルファイルをアップロード
sz /remote/path/to/file         # ファイルをダウンロード（保存ダイアログを開く）
sz /remote/path/to/file /local  # 特定のローカルパスにダウンロード
```

### 仕組み

リモートシェルが `rz`/`sz: command not found` を報告すると、win-sshpass はエラーを傍受し、代わりに SFTP で転送を実行します。ファイルとディレクトリの両方をサポートし、プログレスバーが表示されます。

!!! info "リモートへのインストール不要"
    rz/sz 転送は SFTP を使用して実装されているため、リモートサーバーに lrzsz パッケージをインストールする必要はありません。

## ブレークポイントレジューム

大容量ファイル転送時、`-resume` フラグで中断された転送を途中から再開できます：

```bash
# 中断されたアップロードを再開
win-sshpass -p 'pass' -h host -local ./bigfile.iso -remote /data/bigfile.iso -resume

# 中断されたダウンロードを再開
win-sshpass -p 'pass' -h host -d -remote /data/bigfile.iso -local ./bigfile.iso -resume
```

!!! info "仕組み"
    `-resume` 使用時、win-sshpass は宛先ファイルの既存を確認します。ファイルが存在しかつソースより小さい場合、最後のバイトから転送を継続します。ファイルが既に完全な場合、転送はスキップされます。`-resume` なしの場合、転送は常に最初から開始されます。

## Git Bash パスの注意

Git Bash 使用時、`/` で始まるリモートパスは自動的に Windows パスに変換されます。`//` プレフィックスを使用して回避：

```bash
# 間違い：/tmp は Windows パスに変換されます
win-sshpass ... -remote /tmp/file.txt

# 正しい：二重スラッシュを使用
win-sshpass ... -remote //tmp/file.txt
```

## 次のステップ

- [SCP と Rsync](scp-rsync.md) - SCP/Rsync の詳細な使用方法
- [インタラクティブシェル](shell.md) - シェルモードの全機能
