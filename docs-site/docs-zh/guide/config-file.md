# 配置文件

win-sshpass 支持配置文件来管理多台服务器的连接信息，避免每次输入冗长的命令行参数。

## 配置文件格式

配置文件使用简单的 `key: value` 格式：

```yaml
host: example.com
username: root
password: your_password
port: 22
```

### 支持的字段

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `host` | 主机地址（必填） | - |
| `username` / `user` | 用户名 | root |
| `password` | 密码 | - |
| `port` | 端口 | 22 |
| `key` / `keypath` | 私钥文件路径 | - |
| `timeout` | 操作超时（秒），0 表示无限制 | 0 |
| `connect_timeout` | TCP 连接超时（秒） | 10 |
| `retry` / `retries` | 连接重试次数 | 3 |
| `strict_host_key` | 启用严格主机密钥验证 | false |

### 示例

```yaml
host: 192.168.1.100
username: ubuntu
password: mypassword
port: 22
# key: ~/.ssh/id_ed25519  # 使用私钥认证（与 password 二选一）
# timeout: 0              # 操作超时（秒），0 = 无限制
# connect_timeout: 10     # TCP 连接超时（秒）
# retry: 3                # 连接重试次数
# strict_host_key: false  # 是否启用严格主机密钥验证
```

## 使用配置文件

```bash
# 执行命令
win-sshpass -f server.config -c 'docker ps'

# 也可以省略 -c，直接将命令作为参数
win-sshpass -f server.config 'docker ps'

# 打开交互式 Shell
win-sshpass -f server.config

# 配置文件 + SSH 风格参数
win-sshpass -f server.config ssh user@host 'ls'
```

## 密码文件

如果文件不是配置文件格式（不包含 `host:` 等键），win-sshpass 会将其视为密码文件（单行文本）：

```bash
# pass.txt 内容：mypassword
win-sshpass -f pass.txt ssh user@host
```

## 配置优先级

配置按以下优先级合并（高优先级覆盖低优先级）：

1. **命令行参数**（最高优先级）
2. **配置文件**
3. **默认值**（最低优先级）

例如：

```bash
# 配置文件中 port: 22，但命令行指定 -P 2222
win-sshpass -f server.config -P 2222 ssh user@host
# 实际使用端口 2222
```

## 多服务器管理

为每台服务器创建独立的配置文件：

```
~/.ssh/
├── web-server.config
├── db-server.config
└── staging.config
```

```bash
# Web 服务器
win-sshpass -f ~/.ssh/web-server.config 'nginx -t'

# 数据库服务器
win-sshpass -f ~/.ssh/db-server.config 'systemctl status mysql'

# 测试环境
win-sshpass -f ~/.ssh/staging.config 'docker ps'
```

## 安全建议

!!! warning "密码安全"
    配置文件中包含明文密码，请注意文件权限：

    - 将配置文件放在安全目录（如 `~/.ssh/`）
    - 设置适当的文件权限（仅当前用户可读）
    - 考虑使用私钥认证代替密码

```bash
# Linux/macOS
chmod 600 server.config

# Windows（使用 icacls）
icacls server.config /inheritance:r /grant:r %USERNAME%:R
```

## 下一步

- [SSH 连接](ssh.md) - 更多认证方式
- [最佳实践](../advanced/best-practices.md) - 安全与效率建议
