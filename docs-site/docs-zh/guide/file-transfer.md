# 文件传输

win-sshpass 提供多种文件传输方式：SFTP 直接传输、SCP 风格、Rsync 风格，以及交互式 Shell 中的 rz/sz。

## SFTP 直接传输

### 上传文件

```bash
# 上传单个文件
win-sshpass -h host -p 'pass' -local ./file.txt -remote /tmp/file.txt

# 上传多个文件（逗号分隔）
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# 上传多个文件（空格分隔，仅适用于简单路径）
win-sshpass -h host -p 'pass' -local "a.txt b.txt c.txt" -remote /tmp/

# 上传目录（自动递归）
win-sshpass -h host -p 'pass' -local ./dist -remote /var/www/html
```

### 下载文件

```bash
# 下载文件
win-sshpass -h host -p 'pass' -d -remote /tmp/file.txt -local ./file.txt

# 下载目录
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

### 进度条

SFTP 传输时会自动显示进度条：

```
Uploading app.jar  45% |████████████         |  45MB/100MB  10MB/s
```

## SCP 风格传输

兼容标准 scp 命令语法：

```bash
# 上传文件
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# 上传目录（-r 递归）
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html

# 指定端口（-P，注意是大写）
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/

# 下载文件
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# 下载目录
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

## Rsync 风格传输

兼容 rsync 命令语法：

```bash
# 上传
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/

# 下载
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/

# 指定端口
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

## Shell 内 rz/sz 传输

在交互式 Shell 中，可以使用 `rz` 和 `sz` 命令进行文件传输，**无需在远程服务器上安装任何软件**。

```bash
# 先打开交互式 Shell
win-sshpass -p 'pass' ssh user@host

# 在远程 Shell 中：
rz                              # 上传文件（打开文件选择器）
rz /local/path/to/file          # 上传指定本地文件
sz /remote/path/to/file         # 下载文件（打开保存对话框）
sz /remote/path/to/file /local  # 下载到指定本地路径
```

### 工作原理

当远程 Shell 报告 `rz`/`sz: command not found` 时，win-sshpass 会拦截该错误，并通过 SFTP 执行传输。文件和目录都支持，并显示进度条。

!!! info "无需远程安装"
    rz/sz 传输基于 SFTP 实现，不需要远程服务器安装 lrzsz 包。

## 断点续传

传输大文件时，可使用 `-resume` 参数从中断处恢复：

```bash
# 恢复中断的上传
win-sshpass -p 'pass' -h host -local ./bigfile.iso -remote /data/bigfile.iso -resume

# 恢复中断的下载
win-sshpass -p 'pass' -h host -d -remote /data/bigfile.iso -local ./bigfile.iso -resume
```

!!! info "工作原理"
    使用 `-resume` 时，win-sshpass 会检查目标文件是否已存在。如果存在且小于源文件，则从最后一个字节处继续传输。如果文件已完整，则跳过传输。不使用 `-resume` 时，传输始终从头开始。

## Git Bash 路径注意

在 Git Bash 中使用时，以 `/` 开头的远程路径会被自动转换为 Windows 路径。使用 `//` 前缀避免此问题：

```bash
# 错误：/tmp 会被转换为 Windows 路径
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## 下一步

- [SCP 与 Rsync](scp-rsync.md) - 更多 SCP/Rsync 用法
- [交互式 Shell](shell.md) - Shell 模式的完整功能
