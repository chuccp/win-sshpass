# SSH 连接

win-sshpass 支持多种 SSH 认证方式，满足不同场景需求。

## 密码认证

### 直接指定密码

```bash
win-sshpass -p 'mypassword' ssh user@host
win-sshpass -p 'mypassword' ssh user@host 'whoami'
```

### 从文件读取密码

创建一个只包含密码的文本文件（单行）：

```bash
echo 'mypassword' > pass.txt
win-sshpass -f pass.txt ssh user@host
```

### 从环境变量读取密码

```bash
export SSHPASS='mypassword'
win-sshpass -e ssh user@host
```

或在 Windows CMD 中：

```cmd
set SSHPASS=mypassword
win-sshpass -e ssh user@host
```

!!! tip "安全性建议"
    使用环境变量或配置文件比在命令行中直接传递密码更安全，因为命令历史中不会记录密码。

### 空密码

当服务端账号没有设置密码时，显式传入空密码。省略 `-p` 含义不同——那会回退到 ssh-agent：

```bash
win-sshpass -p '' ssh root@192.168.1.100 'hostname'
```

!!! note "引号写法"
    PowerShell 和 Git Bash 用 `''`；Windows CMD 用 `""`。

## 私钥认证

```bash
# 使用 Ed25519 密钥
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 使用 RSA 密钥
win-sshpass -i ~/.ssh/id_rsa ssh user@host

# 执行远程命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'uname -a'
```

!!! note "注意"
    win-sshpass 不支持加密（有密码保护）的私钥。如果私钥有密码保护，请先解密或使用 ssh-agent。

## SSH Agent 认证

win-sshpass 可以自动使用本地 ssh-agent（如 OpenSSH agent、Windows 上的 Pageant）进行认证。当未指定 `-p`（密码）或 `-i`（密钥路径）时，默认启用 ssh-agent 自动检测：

```bash
# 自动检测并使用 ssh-agent（无需 -p 或 -i）
win-sshpass ssh user@host 'whoami'

# 同样适用于文件传输和 SCP/Rsync
win-sshpass -h host -local file.txt -remote /tmp/file.txt
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
```

!!! tip "ssh-agent 设置"
    确保 ssh-agent 正在运行并已加载密钥（使用 `ssh-add -l` 检查）。在 Windows 上，可以在"服务"中启用 OpenSSH Authentication Agent 服务。

### Agent 转发（`-A`）

使用 `-A` 参数将本地 ssh-agent 转发到远程服务器，允许远程主机使用本地密钥进行进一步的 SSH 连接（例如从跳板机执行 `git clone`）：

```bash
# 启用 agent 转发
win-sshpass -A -i ~/.ssh/id_ed25519 ssh user@jumphost

# 无密码/密钥 — 自动检测 agent + 转发
win-sshpass -A ssh user@jumphost
```

!!! info "Agent 转发需要本地 Agent"
    Agent 转发（`-A`）需要运行中的本地 ssh-agent 并已加载密钥。启用后转发将自动处理。

## 键盘交互式认证

当标准密码认证不可用时，win-sshpass 会自动回退到键盘交互式认证。这确保与基于 PAM 的服务器（如 Cisco 路由器、部分使用自定义 PAM 配置的 Linux 发行版）兼容。无需额外参数 —— 回退是透明的。

## 密钥生成

win-sshpass 内置了 SSH 密钥对生成功能，可以在本地生成客户端密钥对（私钥 + 公钥）。

```bash
# 生成 Ed25519 密钥（推荐，更快更安全）
win-sshpass keygen

# 生成 RSA 密钥（4096 位）
win-sshpass keygen -algo rsa

# 指定输出路径
win-sshpass keygen -out ~/.ssh/mykey

# 指定公钥注释
win-sshpass keygen -comment "my-laptop"
```

默认输出到 `~/.ssh/id_ed25519`（Ed25519）或 `~/.ssh/id_rsa`（RSA），公钥文件自动添加 `.pub` 后缀。

生成后，将公钥部署到服务端即可实现无密码登录（见下文）。

### 部署公钥实现无密码登录

生成密钥后，将公钥部署到服务器即可实现无密码登录：

```bash
# 将公钥内容读入变量，再通过 SSH 部署
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'mypassword' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

Shell 在本地展开 `$PUBKEY` 后再将命令传给 win-sshpass，因此实际的公钥字符串被嵌入到远程命令中，避免了引号转义和标准输入转发的问题。

部署完成后，即可使用私钥进行无密码登录：

```bash
# 无密码登录
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 无密码执行命令
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'whoami'

# 无密码传输文件
win-sshpass -i ~/.ssh/id_ed25519 scp file.txt user@host:/tmp/
```

!!! tip "authorized_keys 权限要求"
    服务端 `~/.ssh` 目录权限应为 700，`~/.ssh/authorized_keys` 权限应为 600。权限不正确会导致密钥认证失败。

## 指定用户和端口

```bash
# 指定用户名（默认 root）
win-sshpass -p 'pass' ssh ubuntu@host

# 指定端口（默认 22）
win-sshpass -p 'pass' ssh -p 2222 user@host

# 使用 -u 和 -P 参数
win-sshpass -p 'pass' -h host -u ubuntu -P 2222

# 主机名不带 user@（用户名默认 root，主机名交给系统解析）
win-sshpass -p 'pass' ssh example.com 'uptime'
```

## 执行远程命令

```bash
# 单条命令
win-sshpass -p 'pass' ssh user@host 'ls -la'

# 多条命令
win-sshpass -p 'pass' ssh user@host 'cd /var/log && ls -la'

# 使用 -c 参数
win-sshpass -p 'pass' -h host -c 'docker ps'
```

## 连接超时与重试

```bash
# TCP 连接超时（默认 10 秒）
win-sshpass -p 'pass' -ct 5 ssh user@host

# 操作超时（默认无限制）
win-sshpass -p 'pass' -t 30 ssh user@host 'long-command'

# 重试次数（默认 3 次）
win-sshpass -p 'pass' -retry 5 ssh user@host
```

超时机制说明：

- **TCP 连接超时**（`-ct`）：建立 TCP 连接的超时时间
- **操作超时**（`-t`）：整个操作的超时时间，数据传输时会自动重置计时器
- **重试**（`-retry`）：连接失败后的重试次数，采用指数退避策略（2s, 4s, 8s, 16s，最大 30s）

!!! info "认证失败不重试"
    如果是认证失败（密码错误、密钥无效），不会进行重试，直接返回错误。

## 主机密钥验证

默认情况下，win-sshpass 不验证主机密钥（等同于 `StrictHostKeyChecking=no`）。

启用严格主机密钥验证：

```bash
# 使用 -k 参数
win-sshpass -p 'pass' -k ssh user@host

# 或在配置文件中设置
# strict_host_key: true
```

启用后，会使用 `~/.ssh/known_hosts` 文件进行验证。如果主机不在 known_hosts 中，连接会被拒绝。

## IPv6 支持

win-sshpass 支持 IPv6 地址：

```bash
win-sshpass -p 'pass' ssh user@2001:db8::1
win-sshpass -p 'pass' ssh user@[2001:db8::1]
```

## 代理支持

通过代理服务器建立 SSH 隧道连接：

```bash
# SOCKS5 代理
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@host

# SOCKS5 带认证
win-sshpass -p 'pass' -proxy socks5://proxyuser:proxypass@127.0.0.1:1080 ssh user@host

# SOCKS4 代理
win-sshpass -p 'pass' -proxy socks4://192.168.1.1:1080 ssh user@host

# HTTP CONNECT 代理
win-sshpass -p 'pass' -proxy http://proxy.local:8080 ssh user@host

# HTTPS CONNECT 代理（带认证）
win-sshpass -p 'pass' -proxy https://user:pass@proxy.local:8443 ssh user@host

# 代理 + 文件传输
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 -h host -local ./file.txt -remote /tmp/file.txt

# 代理 + SCP
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 scp ./app.jar user@host:/opt/app/

# 配置文件中设置代理
# proxy: socks5://user:pass@127.0.0.1:1080
```

!!! info "支持的代理协议"
    支持 SOCKS5（可选用户名/密码认证）、SOCKS4、SOCKS4A、HTTP CONNECT 和 HTTPS CONNECT 代理。

## 端口转发

win-sshpass 支持 SSH 端口转发，可通过 SSH 服务器建立 TCP 隧道连接。本地转发（`-L`）和远程转发（`-R`）均使用标准的 OpenSSH 规范格式。

### 本地转发（`-L`）

通过 SSH 服务器将本地端口转发到远程地址：

```bash
# 格式：-L [绑定地址:]端口:主机:主机端口
# 将 localhost:8080 → db.internal:3306 通过 SSH 主机转发
win-sshpass -p 'pass' -L 8080:db.internal:3306 ssh user@jumphost

# 绑定到特定本地地址
win-sshpass -p 'pass' -L 127.0.0.1:8080:db.internal:3306 ssh user@jumphost

# 多个转发
win-sshpass -p 'pass' -L 8080:db1.internal:3306 -L 8081:db2.internal:3306 ssh user@jumphost

# 仅转发模式（无命令 — 阻塞直到 Ctrl+C）
win-sshpass -p 'pass' -L 8080:db.internal:3306 -L 9090:redis.internal:6379 ssh user@jumphost
```

### 远程转发（`-R`）

将远程端口转发回本地地址：

```bash
# 格式：-R [绑定地址:]端口:主机:主机端口
# 将 localhost:8080 暴露在远程服务器的 9090 端口
win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server

# 允许远程主机连接
win-sshpass -p 'pass' -R 0.0.0.0:9090:localhost:8080 ssh user@server
```

!!! info "端口转发限制"
    端口转发仅支持 SSH 命令/Shell 模式。无法与 SCP、Rsync 或文件传输（`-local`/`-remote`）操作同时使用。

## 文件哈希与校验

无需 SSH 连接即可计算和校验本地文件哈希：

```bash
# 计算哈希
win-sshpass hash md5 ./download.iso
win-sshpass hash sha1 ./download.iso
win-sshpass hash sha256 ./download.iso
win-sshpass hash sha512 ./download.iso

# 校验文件完整性
win-sshpass verify sha256 d1dc38f6dfb1e4c8e7a1b2c3d4e5f6a7b8c9d0e1f2 ./download.iso
# 输出: OK

win-sshpass verify sha256 wronghash123... ./download.iso
# 输出: FAILED
```

支持的算法：`md5`、`sha1`、`sha256`、`sha512`。

## 下一步

- [交互式 Shell](shell.md) - 不指定命令时的交互模式
- [文件传输](file-transfer.md) - SFTP 上传下载
- [配置文件](config-file.md) - 管理多台服务器
