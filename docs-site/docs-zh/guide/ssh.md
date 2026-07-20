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

### 手动部署公钥实现无密码登录

出于安全考虑，win-sshpass **不会自动连接服务器部署公钥**。请手动将公钥部署到服务端：

```bash
# 方法一：用 win-sshpass 手动追加公钥到 authorized_keys
win-sshpass -p 'mypassword' ssh user@host 'mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo "$(cat)" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys' < ~/.ssh/id_ed25519.pub

# 方法二：手动复制公钥内容，登录服务器后追加
cat ~/.ssh/id_ed25519.pub
# 然后登录服务器，将公钥追加到 ~/.ssh/authorized_keys
```

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
