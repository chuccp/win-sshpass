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

## 下一步

- [交互式 Shell](shell.md) - 不指定命令时的交互模式
- [文件传输](file-transfer.md) - SFTP 上传下载
- [配置文件](config-file.md) - 管理多台服务器
