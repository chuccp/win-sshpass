# SCP 与 Rsync

win-sshpass 兼容标准 scp 和 rsync 命令语法，底层通过 SFTP 实现文件传输。

## SCP 风格传输

### 基本语法

```bash
win-sshpass -p <password> scp [选项] <源> <目标>
```

### 上传文件

```bash
# 上传单个文件
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# 上传到指定文件名
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/newname.txt

# 上传目录（-r 递归）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html
```

### 下载文件

```bash
# 下载文件
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# 下载目录
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

### 指定端口

scp 使用大写 `-P` 指定端口（与 ssh 的小写 `-p` 不同）：

```bash
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/
```

### 支持的选项

| 选项 | 说明 |
|------|------|
| `-r` | 递归复制目录 |
| `-P <port>` | 指定端口 |
| `-i <key>` | 指定私钥 |
| `-q` | 静默模式 |
| `-C` | 压缩（已由 SFTP 处理） |
| `-v` | 详细输出 |

## Rsync 风格传输

### 基本语法

```bash
win-sshpass -p <password> rsync [选项] <源> <目标>
```

### 上传

```bash
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 下载

```bash
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/
```

### 指定端口

rsync 使用 `--port=` 指定端口：

```bash
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

### 支持的选项

| 选项 | 说明 |
|------|------|
| `-a` | 归档模式 |
| `-v` | 详细输出 |
| `-z` | 压缩传输 |
| `--port=N` | 指定端口 |
| `-e ssh` | 指定远程 Shell（会被忽略） |

## SCP vs Rsync vs SFTP

| 方式 | 语法兼容 | 适用场景 |
|------|----------|----------|
| SCP | 标准 scp 语法 | 简单的文件复制 |
| Rsync | 标准 rsync 语法 | 增量同步（注意：当前实现为全量传输） |
| SFTP | `-local` / `-remote` 参数 | 灵活的文件传输，支持多文件 |

!!! note "注意"
    win-sshpass 的 rsync 实现底层使用 SFTP，不支持 rsync 的增量同步算法。如果需要真正的增量同步，请在远程服务器上安装 rsync 并通过 SSH 直接使用。

## 下一步

- [文件传输](file-transfer.md) - SFTP 直接传输的更多用法
- [配置文件](config-file.md) - 使用配置文件简化命令
