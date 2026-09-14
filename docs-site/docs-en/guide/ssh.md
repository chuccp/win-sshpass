# SSH Connection

win-sshpass supports multiple SSH authentication methods for different scenarios.

## Password Authentication

### Direct Password

```bash
win-sshpass -p 'mypassword' ssh user@host
win-sshpass -p 'mypassword' ssh user@host 'whoami'
```

### Password from File

Create a text file containing only the password (single line):

```bash
echo 'mypassword' > pass.txt
win-sshpass -f pass.txt ssh user@host
```

### Password from Environment Variable

```bash
export SSHPASS='mypassword'
win-sshpass -e ssh user@host
```

Or in Windows CMD:

```cmd
set SSHPASS=mypassword
win-sshpass -e ssh user@host
```

!!! tip "Security Tip"
    Using environment variables or config files is more secure than passing passwords on the command line, as the password won't appear in command history.

### Empty Password

For servers whose account has no password set, pass an explicitly empty password.
Omitting `-p` is not the same thing — that falls back to ssh-agent instead:

```bash
win-sshpass -p '' ssh root@192.168.1.100 'hostname'
```

!!! note "Quoting"
    Use `''` in PowerShell and Git Bash; in Windows CMD use `""`.

## Private Key Authentication

```bash
# Using Ed25519 key
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# Using RSA key
win-sshpass -i ~/.ssh/id_rsa ssh user@host

# Execute remote command
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'uname -a'
```

!!! note "Note"
    win-sshpass does not support encrypted (passphrase-protected) private keys. If your key is passphrase-protected, decrypt it first or use ssh-agent.

## SSH Agent Authentication

win-sshpass can automatically use your local ssh-agent (e.g., OpenSSH agent, Pageant on Windows) for authentication. When neither `-p` (password) nor `-i` (key path) is specified, ssh-agent auto-detection is enabled by default:

```bash
# Auto-detect and use ssh-agent (no -p or -i needed)
win-sshpass ssh user@host 'whoami'

# Also works with file transfer and SCP/Rsync
win-sshpass -h host -local file.txt -remote /tmp/file.txt
win-sshpass scp file.txt user@host:/tmp/
win-sshpass rsync -avz ./ user@host:/backup/
```

!!! tip "ssh-agent Setup"
    Make sure your ssh-agent is running and has keys loaded (`ssh-add -l` to check). On Windows, the OpenSSH Authentication Agent service can be enabled in Services.

### Agent Forwarding (`-A`)

Use the `-A` flag to forward your local ssh-agent to the remote server, allowing the remote host to use your local keys for further SSH connections (e.g., `git clone` via SSH from a jump host):

```bash
# Enable agent forwarding
win-sshpass -A -i ~/.ssh/id_ed25519 ssh user@jumphost

# Without password/key — auto-detect agent + forward
win-sshpass -A ssh user@jumphost
```

!!! info "Agent Forwarding Requires Local Agent"
    Agent forwarding (`-A`) requires a running local ssh-agent with keys loaded. The forwarding is handled automatically once enabled.

## Keyboard-Interactive Authentication

win-sshpass automatically falls back to keyboard-interactive authentication when standard password authentication is not available. This ensures compatibility with PAM-based servers (e.g., Cisco routers, some Linux distributions with custom PAM configurations). No additional flags are required — the fallback is transparent.

## Key Generation

win-sshpass has built-in SSH key pair generation. It creates a client-side key pair (private key + public key) locally.

```bash
# Generate Ed25519 key (recommended — faster and more secure)
win-sshpass keygen

# Generate RSA key (4096-bit)
win-sshpass keygen -algo rsa

# Specify output path
win-sshpass keygen -out ~/.ssh/mykey

# Specify public key comment
win-sshpass keygen -comment "my-laptop"
```

Keys are saved to `~/.ssh/id_ed25519` (Ed25519) or `~/.ssh/id_rsa` (RSA) by default. The public key file is automatically given a `.pub` suffix.

After generation, deploy the public key to the server to enable password-less login (see below).

### Deploying the Public Key

After generation, deploy the public key to the server to enable password-less login:

```bash
# Read the public key content into a variable, then deploy via SSH
PUBKEY=$(cat ~/.ssh/id_ed25519.pub)
win-sshpass -p 'mypassword' ssh user@host "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '$PUBKEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
```

The shell expands `$PUBKEY` locally before passing the command to win-sshpass, so the actual public key string is embedded in the remote command. This avoids quoting and stdin-forwarding issues.

After deployment, you can log in without a password using the private key:

```bash
# Password-less login
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# Password-less command execution
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host 'whoami'

# Password-less file transfer
win-sshpass -i ~/.ssh/id_ed25519 scp file.txt user@host:/tmp/
```

!!! tip "authorized_keys Permissions"
    The server's `~/.ssh` directory should be mode 700 and `~/.ssh/authorized_keys` should be mode 600. Incorrect permissions will cause key authentication to fail.

## Specifying User and Port

```bash
# Specify username (default: root)
win-sshpass -p 'pass' ssh ubuntu@host

# Specify port (default: 22)
win-sshpass -p 'pass' ssh -p 2222 user@host

# Using -u and -P flags
win-sshpass -p 'pass' -h host -u ubuntu -P 2222

# Host without user@ — username defaults to root, the name is resolved by the OS
win-sshpass -p 'pass' ssh example.com 'uptime'
```

## Executing Remote Commands

```bash
# Single command
win-sshpass -p 'pass' ssh user@host 'ls -la'

# Multiple commands
win-sshpass -p 'pass' ssh user@host 'cd /var/log && ls -la'

# Using -c flag
win-sshpass -p 'pass' -h host -c 'docker ps'
```

## Connection Timeout and Retry

```bash
# TCP connection timeout (default: 10 seconds)
win-sshpass -p 'pass' -ct 5 ssh user@host

# Operation timeout (default: no limit)
win-sshpass -p 'pass' -t 30 ssh user@host 'long-command'

# Retry count (default: 3)
win-sshpass -p 'pass' -retry 5 ssh user@host
```

Timeout mechanism:

- **TCP connection timeout** (`-ct`): Timeout for establishing TCP connection
- **Operation timeout** (`-t`): Total operation timeout; timer resets automatically during data transfer
- **Retry** (`-retry`): Number of connection retry attempts with exponential backoff (2s, 4s, 8s, 16s, capped at 30s)

!!! info "No Retry on Auth Failure"
    Authentication failures (wrong password, invalid key) are not retried — the error is returned immediately.

## Host Key Verification

By default, win-sshpass does not verify host keys (equivalent to `StrictHostKeyChecking=no`).

Enable strict host key verification:

```bash
# Using -k flag
win-sshpass -p 'pass' -k ssh user@host

# Or in config file
# strict_host_key: true
```

When enabled, the `~/.ssh/known_hosts` file is used for verification. If the host is not in known_hosts, the connection is rejected.

## IPv6 Support

win-sshpass supports IPv6 addresses:

```bash
win-sshpass -p 'pass' ssh user@2001:db8::1
win-sshpass -p 'pass' ssh user@[2001:db8::1]
```

## Proxy Support

Tunnel SSH connections through a proxy server:

```bash
# SOCKS5 proxy
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@host

# SOCKS5 with authentication
win-sshpass -p 'pass' -proxy socks5://proxyuser:proxypass@127.0.0.1:1080 ssh user@host

# SOCKS4 proxy
win-sshpass -p 'pass' -proxy socks4://192.168.1.1:1080 ssh user@host

# HTTP CONNECT proxy
win-sshpass -p 'pass' -proxy http://proxy.local:8080 ssh user@host

# HTTPS CONNECT proxy with authentication
win-sshpass -p 'pass' -proxy https://user:pass@proxy.local:8443 ssh user@host

# Proxy with file transfer
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 -h host -local ./file.txt -remote /tmp/file.txt

# Proxy with SCP
win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 scp ./app.jar user@host:/opt/app/

# Proxy via config file
# proxy: socks5://user:pass@127.0.0.1:1080
```

!!! info "Supported Proxy Protocols"
    SOCKS5 (with optional username/password auth), SOCKS4, SOCKS4A, HTTP CONNECT, and HTTPS CONNECT proxies are all supported.

## Port Forwarding

win-sshpass supports SSH port forwarding to tunnel TCP connections through an SSH server. Both local (`-L`) and remote (`-R`) forwarding use the standard OpenSSH spec format.

### Local Forwarding (`-L`)

Forward a local port to a remote address through the SSH server:

```bash
# Format: -L [bind_address:]port:host:hostport
# Forward localhost:8080 → db.internal:3306 via the SSH host
win-sshpass -p 'pass' -L 8080:db.internal:3306 ssh user@jumphost

# Bind to a specific local address
win-sshpass -p 'pass' -L 127.0.0.1:8080:db.internal:3306 ssh user@jumphost

# Multiple forwards
win-sshpass -p 'pass' -L 8080:db1.internal:3306 -L 8081:db2.internal:3306 ssh user@jumphost

# Forward-only mode (no command — blocks until Ctrl+C)
win-sshpass -p 'pass' -L 8080:db.internal:3306 -L 9090:redis.internal:6379 ssh user@jumphost
```

### Remote Forwarding (`-R`)

Forward a remote port back to a local address:

```bash
# Format: -R [bind_address:]port:host:hostport
# Expose localhost:8080 on the remote server's port 9090
win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server

# Allow remote hosts to connect
win-sshpass -p 'pass' -R 0.0.0.0:9090:localhost:8080 ssh user@server
```

!!! info "Port Forwarding Limits"
    Port forwarding is only supported with SSH command/shell mode. It cannot be combined with SCP, Rsync, or file transfer (`-local`/`-remote`) operations.

## File Hash & Verify

Compute and verify checksums of local files without needing an SSH connection:

```bash
# Compute hash
win-sshpass hash md5 ./download.iso
win-sshpass hash sha1 ./download.iso
win-sshpass hash sha256 ./download.iso
win-sshpass hash sha512 ./download.iso

# Verify file integrity against expected hash
win-sshpass verify sha256 d1dc38f6dfb1e4c8e7a1b2c3d4e5f6a7b8c9d0e1f2 ./download.iso
# Output: OK

win-sshpass verify sha256 wronghash123... ./download.iso
# Output: FAILED
```

Supported algorithms: `md5`, `sha1`, `sha256`, `sha512`.

## Next Steps

- [Interactive Shell](shell.md) - Interactive mode when no command is specified
- [File Transfer](file-transfer.md) - SFTP upload/download
- [Configuration File](config-file.md) - Manage multiple servers
