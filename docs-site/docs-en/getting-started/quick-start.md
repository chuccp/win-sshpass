# Quick Start

This guide will help you get started with win-sshpass in minutes.

## 5 Seconds: Execute Remote Command

```bash
# Login with password and execute command
win-sshpass -p 'mypassword' ssh root@192.168.1.100 'whoami'
# → root
```

## 30 Seconds: File Transfer

```bash
# Upload file
win-sshpass -h 192.168.1.100 -p 'mypassword' -local ./app.jar -remote /opt/app/

# Download file
win-sshpass -h 192.168.1.100 -p 'mypassword' -d -remote /var/log/app.log -local ./logs/
```

## 1 Minute: Interactive Shell

```bash
# Open interactive shell (no command specified)
win-sshpass -p 'mypassword' ssh root@192.168.1.100
```

Once connected, you can:

- Type commands normally with proper echo
- Use `vim`, `top`, `htop`, and other full-screen apps
- Press `Ctrl+C` to interrupt the current command
- Resize your terminal window — the remote terminal adjusts automatically

## 3 Minutes: Using Configuration Files

Create `server.config`:

```yaml
host: 192.168.1.100
username: root
password: mypassword
port: 22
```

Use the config file:

```bash
# Execute command
win-sshpass -f server.config -c 'docker ps'

# Or pass command as positional argument
win-sshpass -f server.config 'docker ps'

# Open interactive shell
win-sshpass -f server.config
```

## 5 Minutes: SCP/Rsync Style Transfer

```bash
# SCP upload
win-sshpass -p 'mypassword' scp ./app.jar user@server:/opt/app/

# SCP upload directory
win-sshpass -p 'mypassword' scp -r ./dist user@server:/var/www/html

# Rsync upload
win-sshpass -p 'mypassword' rsync -avz ./ user@server:/backup/
```

## Common Usage Cheat Sheet

```bash
# Private key login
win-sshpass -i ~/.ssh/id_ed25519 ssh user@server

# Password via environment variable (more secure)
export SSHPASS='mypassword'
win-sshpass -e ssh user@server

# Custom port
win-sshpass -p 'pass' ssh -p 2222 user@server

# Operation timeout (30 seconds)
win-sshpass -p 'pass' -t 30 ssh user@server 'long-running-command'

# Upload multiple files
win-sshpass -h server -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/
```

## Next Steps

- [SSH Connection](../guide/ssh.md) - Deep dive into authentication
- [File Transfer](../guide/file-transfer.md) - Complete SFTP usage
- [Configuration File](../guide/config-file.md) - Manage multiple servers
