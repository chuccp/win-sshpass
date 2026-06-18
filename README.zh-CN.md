# win-sshpass

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

Windows 版 sshpass 工具，实现类似 Linux sshpass 的功能。

## 功能特性

- 支持密码或私钥认证的 SSH 登录
- 执行远程命令或打开交互式 Shell
- 通过 SFTP 上传/下载文件（带进度条）
- SCP 风格和 Rsync 风格的文件传输
- 配置文件支持，方便管理多台服务器
- 交互式 Shell 使用 raw 终端模式（正确的回显、Ctrl+C、vim/top 全屏程序支持）
- 交互式 Shell 模式下动态调整终端大小
- Git Bash 路径转换检测与自动修复
- 支持 IPv6 地址
- 支持 x64 (amd64) 和 ARM64 架构

## 下载

从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下载最新版本：

| 架构 | Zip | MSI 安装包 |
|------|-----|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. 打开 [Releases](https://github.com/chuccp/win-sshpass/releases) 页面
2. 下载对应架构的 zip 或 MSI 安装包（x64 或 ARM64）
3. 如果使用 MSI 安装包：运行安装程序即可，安装目录会自动添加到系统 PATH 中

> **零依赖**：`win-sshpass.exe` 是一个独立的可执行文件，无需安装 OpenSSH 或任何其他软件。下载后放入 PATH 目录即可直接使用。

## 快速开始

```bash
# 密码登录执行命令
win-sshpass -p 'password' ssh user@example.com 'whoami'

# 私钥登录执行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# 上传文件
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# 下载文件
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## 命令格式

### SSH 登录

```bash
# 密码认证
win-sshpass -p <密码> ssh [user@host] [命令]
win-sshpass -p <密码> ssh -p <端口> user@host '命令'
win-sshpass -p <密码> ssh -o StrictHostKeyChecking=no user@host

# 交互式 Shell（raw 终端模式：正确的回显、Ctrl+C、vim/top 支持）
win-sshpass -p <密码> ssh user@host

# 私钥认证
win-sshpass -i <私钥路径> ssh [user@host] [命令]

# 环境变量密码
SSHPASS=<密码> win-sshpass -e ssh user@host

# 密码文件
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# 配置文件（多行格式）
win-sshpass -f server.config
```

### 文件传输

> **Git Bash 用户**：远程路径需使用 `//` 前缀，例如 `-remote //tmp/file.txt`。详见下方 [Git Bash 注意事项](#git-bash-注意事项)。

```bash
# 上传文件
win-sshpass -h <主机> -p <密码> -local <本地路径> -remote <远程路径>

# 上传多个文件（逗号分隔）
win-sshpass -h <主机> -p <密码> -local "a.txt,b.txt,c.txt" -remote //tmp/

# 上传多个文件（空格分隔，仅适用于不含 / 或 \ 的简单路径）
win-sshpass -h <主机> -p <密码> -local "a.txt b.txt c.txt" -remote //tmp/

# 上传目录（自动递归）
win-sshpass -h <主机> -p <密码> -local <本地目录> -remote <远程目录>

# 下载文件/目录
win-sshpass -h <主机> -p <密码> -d -remote <远程路径> -local <本地路径>
```

### SCP 风格

```bash
# 上传文件
win-sshpass -p <密码> scp <本地文件> user@host:<远程路径>
win-sshpass -p <密码> scp -P <端口> <本地文件> user@host:<远程路径>

# 上传目录
win-sshpass -p <密码> scp -r <本地目录> user@host:<远程路径>

# 下载文件/目录
win-sshpass -p <密码> scp user@host:<远程文件> <本地路径>
```

### Rsync 风格

```bash
# 上传
win-sshpass -p <密码> rsync -avz <本地路径> user@host:<远程路径>

# 下载
win-sshpass -p <密码> rsync -avz user@host:<远程路径> <本地路径>
```

## 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `-p` | 密码 | `-p 'secret123'` |
| `-i` | 私钥路径 | `-i ~/.ssh/id_ed25519` |
| `-f` | 密码文件/配置文件 | `-f pass.txt` |
| `-e` | 从环境变量 SSHPASS 读密码 | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | 主机地址 | `-h example.com` |
| `-u` | 用户名，默认 root | `-u ubuntu` |
| `-P` | 端口，默认 22 | `-P 2222` |
| `-c` | 执行的命令 | `-c 'ls -la'` |
| `-local` | 本地路径（逗号或空格分隔） | `-local "a.txt,b.txt"` |
| `-remote` | 远程路径（上传/下载） | `-remote /tmp/file.txt` |
| `-d` | 下载模式 | `-d` |
| `-k` | 启用严格主机密钥验证 | `-k` |
| `-t` | 总操作超时时间（秒），0 表示不限 | `-t 30` |
| `-ct` | TCP 连接超时时间（秒），默认 10 | `-ct 5` |
| `-v` | 显示版本 | `-v` |
| `-help` | 显示帮助信息 | `-help` |

## 配置文件格式

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # 可选，使用私钥代替密码
# timeout: 0              # 可选，总操作超时时间（秒），0 表示不限
# connect_timeout: 10     # 可选，TCP 连接超时时间（秒）
# strict_host_key: false  # 可选，启用严格主机密钥验证
```

使用方式：
```bash
win-sshpass -f server.config -c 'ls -la'
win-sshpass -f server.config 'ls -la'
```

## 完整示例

```bash
# 1. 密码登录执行命令
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. 私钥登录执行 sudo 命令
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. 上传整个目录到服务器
win-sshpass -h server.com -p 'mypass' -local ./dist -remote //var/www/html

# 4. 下载服务器日志目录
win-sshpass -h server.com -p 'mypass' -d -remote //var/log/nginx -local ./logs

# 5. SCP 上传文件
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. 环境变量传递密码（更安全）
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'

# 7. 操作超时（30 秒后自动中断）
win-sshpass -p 'mypass' -t 30 ssh user@server.com 'long-running-command'

# 8. 配置文件 + 位置参数命令
win-sshpass -f server.config 'docker ps'
```

## Git Bash 注意事项

远程路径用 `//` 开头避免路径转换：
```bash
# 错误：/tmp 会被转换为 Windows 路径
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## 编译

```bash
go build -o win-sshpass.exe .
```

## 依赖

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3