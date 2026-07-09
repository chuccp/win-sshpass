# 最佳实践

## 安全建议

### 1. 避免在命令行中直接传递密码

```bash
# 不推荐：密码会出现在命令历史中
win-sshpass -p 'mypassword' ssh user@host

# 推荐：使用环境变量
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# 推荐：使用密码文件
win-sshpass -f pass.txt ssh user@host

# 推荐：使用配置文件
win-sshpass -f server.config ssh user@host
```

### 2. 使用私钥认证

私钥认证比密码认证更安全：

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. 保护配置文件权限

```bash
# Linux/macOS
chmod 600 server.config

# Windows（PowerShell）
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. 启用主机密钥验证

在生产环境中，建议启用严格主机密钥验证：

```bash
win-sshpass -k -f server.config ssh user@host
```

或在配置文件中：

```yaml
strict_host_key: true
```

## 效率建议

### 1. 使用配置文件管理多台服务器

为常用服务器创建配置文件，避免重复输入参数：

```bash
# ~/.ssh/prod-web.config
host: web.example.com
username: deploy
key: ~/.ssh/id_ed25519

# ~/.ssh/prod-db.config
host: db.example.com
username: admin
key: ~/.ssh/id_ed25519
```

### 2. 批量操作

结合 Shell 脚本进行批量操作：

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. 使用 SSH 风格语法

对于熟悉 SSH 的用户，可以使用更自然的语法：

```bash
# 标准 SSH 语法
win-sshpass -p 'pass' ssh user@host 'command'

# SCP 语法
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync 语法
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. 设置合理的超时

```bash
# 快速命令：短超时
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# 长时间操作：长超时或无超时
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

## 故障排查

### 连接失败

```bash
# 增加重试次数
win-sshpass -p 'pass' -retry 5 ssh user@host

# 增加连接超时
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### 认证失败

- 检查密码是否正确
- 检查私钥路径是否正确
- 检查远程服务器是否允许密码/密钥认证
- 注意：不支持加密的私钥

### Git Bash 路径问题

```bash
# 错误：/tmp 会被 Git Bash 转换
win-sshpass ... -remote /tmp/file.txt

# 正确：使用双斜杠
win-sshpass ... -remote //tmp/file.txt
```

## 下一步

- [Go SDK](sdk.md) - 以编程方式使用
- [更新日志](../changelog.md) - 版本更新记录
