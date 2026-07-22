# Configuration File

win-sshpass supports configuration files to manage connection info for multiple servers, avoiding lengthy command-line arguments every time.

## Config File Format

Config files use a simple `key: value` format:

```yaml
host: example.com
username: root
password: your_password
port: 22
```

### Supported Fields

| Field | Description | Default |
|-------|-------------|---------|
| `host` | Host address (required) | - |
| `username` / `user` | Username | root |
| `password` | Password | - |
| `port` | Port | 22 |
| `key` / `keypath` | Private key file path | - |
| `timeout` | Operation timeout in seconds, 0 = no limit | 0 |
| `connect_timeout` | TCP connection timeout in seconds | 10 |
| `retry` / `retries` | Connection retry count | 3 |
| `strict_host_key` | Enable strict host key verification | false |
| `proxy` / `proxy_url` | Proxy URL for SSH tunnel | - |

### Example

```yaml
host: 192.168.1.100
username: ubuntu
password: mypassword
port: 22
# key: ~/.ssh/id_ed25519  # Use private key auth (alternative to password)
# timeout: 0              # Operation timeout (seconds), 0 = no limit
# connect_timeout: 10     # TCP connection timeout (seconds)
# retry: 3                # Connection retry count
# strict_host_key: false  # Enable strict host key verification
# proxy: socks5://127.0.0.1:1080  # Proxy URL (socks5/socks4/http/https)
```

## Using Config Files

```bash
# Execute command
win-sshpass -f server.config -c 'docker ps'

# Or pass command as positional argument
win-sshpass -f server.config 'docker ps'

# Open interactive shell
win-sshpass -f server.config

# Config file + SSH-style arguments
win-sshpass -f server.config ssh user@host 'ls'
```

## Password File

If the file is not in config format (doesn't contain `host:` etc.), win-sshpass treats it as a password file (single-line text):

```bash
# pass.txt contains: mypassword
win-sshpass -f pass.txt ssh user@host
```

## Configuration Priority

Configuration is merged in the following priority order (higher overrides lower):

1. **Command-line arguments** (highest priority)
2. **Configuration file**
3. **Default values** (lowest priority)

Example:

```bash
# Config file has port: 22, but command line specifies -P 2222
win-sshpass -f server.config -P 2222 ssh user@host
# Uses port 2222
```

## Multi-Server Management

Create separate config files for each server:

```
~/.ssh/
├── web-server.config
├── db-server.config
└── staging.config
```

```bash
# Web server
win-sshpass -f ~/.ssh/web-server.config 'nginx -t'

# Database server
win-sshpass -f ~/.ssh/db-server.config 'systemctl status mysql'

# Staging environment
win-sshpass -f ~/.ssh/staging.config 'docker ps'
```

## Security Tips

!!! warning "Password Security"
    Config files contain plaintext passwords. Take care with file permissions:

    - Store config files in a secure directory (e.g., `~/.ssh/`)
    - Set appropriate file permissions (readable only by current user)
    - Consider using private key authentication instead of passwords

```bash
# Linux/macOS
chmod 600 server.config

# Windows (using icacls)
icacls server.config /inheritance:r /grant:r %USERNAME%:R
```

## Next Steps

- [SSH Connection](ssh.md) - More authentication methods
- [Best Practices](../advanced/best-practices.md) - Security and efficiency tips
