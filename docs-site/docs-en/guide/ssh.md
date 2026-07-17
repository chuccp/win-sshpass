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

## Specifying User and Port

```bash
# Specify username (default: root)
win-sshpass -p 'pass' ssh ubuntu@host

# Specify port (default: 22)
win-sshpass -p 'pass' ssh -p 2222 user@host

# Using -u and -P flags
win-sshpass -p 'pass' -h host -u ubuntu -P 2222
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
