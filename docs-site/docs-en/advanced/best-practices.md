# Best Practices

## Security Tips

### 1. Avoid Passing Passwords on the Command Line

```bash
# Not recommended: password appears in command history
win-sshpass -p 'mypassword' ssh user@host

# Recommended: use environment variable
export SSHPASS='mypassword'
win-sshpass -e ssh user@host

# Recommended: use password file
win-sshpass -f pass.txt ssh user@host

# Recommended: use config file
win-sshpass -f server.config ssh user@host
```

### 2. Use Private Key Authentication

Private key authentication is more secure than password authentication:

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

### 3. Protect Config File Permissions

```bash
# Linux/macOS
chmod 600 server.config

# Windows (PowerShell)
$acl = Get-Acl server.config
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule($env:USERNAME, "FullControl", "Allow")
$acl.AddAccessRule($rule)
Set-Acl server.config $acl
```

### 4. Enable Host Key Verification

In production environments, it's recommended to enable strict host key verification:

```bash
win-sshpass -k -f server.config ssh user@host
```

Or in config file:

```yaml
strict_host_key: true
```

## Efficiency Tips

### 1. Use Config Files for Server Management

Create config files for frequently used servers to avoid repeating parameters:

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

### 2. Batch Operations

Combine with shell scripts for batch operations:

```bash
#!/bin/bash
for host in web1 web2 web3; do
    win-sshpass -f ~/.ssh/$host.config 'sudo systemctl restart nginx' &
done
wait
```

### 3. Use SSH-Style Syntax

For users familiar with SSH, a more natural syntax is available:

```bash
# Standard SSH syntax
win-sshpass -p 'pass' ssh user@host 'command'

# SCP syntax
win-sshpass -p 'pass' scp file.txt user@host:/tmp/

# Rsync syntax
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### 4. Set Reasonable Timeouts

```bash
# Quick commands: short timeout
win-sshpass -p 'pass' -ct 5 -t 10 ssh user@host 'echo ok'

# Long operations: long timeout or no timeout
win-sshpass -p 'pass' -t 300 ssh user@host 'backup.sh'
```

## Troubleshooting

### Connection Failures

```bash
# Increase retry count
win-sshpass -p 'pass' -retry 5 ssh user@host

# Increase connection timeout
win-sshpass -p 'pass' -ct 30 ssh user@host
```

### Authentication Failures

- Verify the password is correct
- Verify the private key path is correct
- Check if the remote server allows password/key authentication
- Note: encrypted private keys are not supported

### Git Bash Path Issues

```bash
# Wrong: /tmp will be converted by Git Bash
win-sshpass ... -remote /tmp/file.txt

# Correct: use double slashes
win-sshpass ... -remote //tmp/file.txt
```

## Next Steps

- [Go SDK](sdk.md) - Use programmatically
- [Changelog](../changelog.md) - Version history
