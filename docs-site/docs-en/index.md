# win-sshpass User Guide

> A cross-platform implementation of sshpass (Windows & Linux): password/key SSH login, interactive shell, SFTP file transfer, SCP/Rsync-style transfer, proxy tunneling, breakpoint-resume, file hash/verify, and a reusable Go SDK.

## What is win-sshpass?

win-sshpass is a cross-platform implementation of the Linux sshpass tool. It runs on Windows and Linux as a standalone executable with no dependencies — no need to install OpenSSH or any other software. Download it and you're ready to go.

It supports both **command-line tool** and **Go SDK** usage:

- **Password/Key authentication**: Supports password, private key, environment variable, and config file authentication.
- **Interactive shell**: Raw terminal mode with proper echo, Ctrl+C, full-screen apps (vim, top), and dynamic terminal resizing.
- **SFTP file transfer**: Upload/download files and directories with progress bars.
- **SCP/Rsync style**: Compatible with scp and rsync command syntax.
- **Proxy support**: Tunnel SSH connections through SOCKS5/SOCKS4/HTTP/HTTPS proxies.
- **Breakpoint resume**: Resume interrupted SFTP file transfers from where they left off.
- **File hash & verify**: Compute and verify local file hashes (MD5, SHA-1, SHA-256, SHA-512).
- **Reusable Go SDK**: Import as a Go library to embed SSH/SFTP/shell capabilities in your own application with injectable I/O streams and progress callbacks.

## 30-Second Quick Start

```bash
# Install via WinGet
winget install chuccp.win-sshpass

# Install via Scoop
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

```bash
# Login with password and execute command
win-sshpass -p 'password' ssh user@example.com 'whoami'

# Upload a file
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# Open interactive shell
win-sshpass -p 'password' ssh user@host
```

## Core Features

### 1. Multiple Authentication Methods

**Password authentication**: Direct password, from file, or from environment variable.

```bash
# Direct password
win-sshpass -p 'secret' ssh user@host

# From file
win-sshpass -f pass.txt ssh user@host

# From environment variable
SSHPASS='secret' win-sshpass -e ssh user@host
```

**Private key authentication**: Supports Ed25519, RSA, and other key formats.

```bash
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host
```

**Configuration file**: Manage connection info for multiple servers.

```bash
win-sshpass -f server.config -c 'docker ps'
```

### 2. Interactive Shell (Raw Terminal Mode)

When no command is specified, win-sshpass opens an interactive shell with:

- **Proper echo** — typed characters are displayed correctly (no double echo)
- **Ctrl+C / Ctrl+Z** — signals are forwarded to the remote process
- **Full-screen apps** — vim, top, htop, nano work correctly
- **Dynamic terminal resizing** — the remote terminal automatically matches your local window size
- **Tab completion** — remote shell tab completion works as expected

```bash
win-sshpass -p 'password' ssh user@host
```

### 3. File Transfer

**SFTP**: Upload/download files and directories with progress bars.

```bash
# Upload file
win-sshpass -h host -p 'pass' -local file.txt -remote /tmp/file.txt

# Upload multiple files
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# Download directory
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

**SCP style**: Compatible with scp command syntax.

```bash
win-sshpass -p 'pass' scp ./app.jar user@server:/opt/app/
win-sshpass -p 'pass' scp -r ./dist user@server:/var/www/html
```

**Rsync style**: Compatible with rsync command syntax.

```bash
win-sshpass -p 'pass' rsync -avz ./ user@server:/backup/
```

**Shell rz/sz**: Use rz/sz commands directly in interactive shell.

```bash
# In remote shell:
rz                              # Upload file (opens file picker)
sz /remote/path/to/file         # Download file
```

### 4. Reusable Go SDK

win-sshpass is also a Go library (`package sshpass`) that you can embed in your own application:

```go
import sshpass "github.com/chuccp/win-sshpass"

cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"

client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
defer client.Close()

// Execute command
client.Exec("uname -a")

// SFTP transfer
sftp, _ := client.SFTP()
sftp.Upload("./local.txt", "/tmp/remote.txt")
```

The SDK ships **no UI code** (no progress bar, no file dialog). Behavior is configured through functional options:

| Option | Purpose |
|--------|---------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | Redirect I/O streams |
| `WithProgress(fn)` | Set transfer progress callback |
| `WithFileSelector(s)` | Set rz/sz file selector |
| `WithResume()` | Enable breakpoint-resume for file transfers |
| `WithSignalHandler()` | Register Ctrl+C signal handler |

## Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `-p` | Password | `-p 'secret123'` |
| `-i` | Private key path | `-i ~/.ssh/id_ed25519` |
| `-f` | Password file / config file | `-f pass.txt` |
| `-e` | Read password from SSHPASS env var | `SSHPASS='pass' win-sshpass -e ssh ...` |
| `-h` | Host address | `-h example.com` |
| `-u` | Username, default: root | `-u ubuntu` |
| `-P` | Port, default: 22 | `-P 2222` |
| `-c` | Command to execute | `-c 'ls -la'` |
| `-local` | Local path(s) (comma or space separated) | `-local "a.txt,b.txt"` |
| `-remote` | Remote path | `-remote /tmp/file.txt` |
| `-d` | Download mode | `-d` |
| `-k` | Enable strict host key verification | `-k` |
| `-t` | Total operation timeout in seconds (0 = no limit) | `-t 30` |
| `-ct` | TCP connection timeout in seconds (default: 10) | `-ct 5` |
| `-retry` | Total connection attempts (default: 3) | `-retry 5` |
| `-resume` | Resume interrupted file transfer from breakpoint | `-resume` |
| `-proxy` | Proxy URL (socks5/socks4/http/https) | `-proxy socks5://127.0.0.1:1080` |
| `-v` | Show version | `-v` |
| `-help` | Show help | `-help` |

## Quick Links

### Getting Started

- [Installation](getting-started/installation.md) - Download and installation
- [Quick Start](getting-started/quick-start.md) - Your first connection

### User Guide

- [SSH Connection](guide/ssh.md) - Password, key, and environment variable authentication
- [File Transfer](guide/file-transfer.md) - SFTP upload/download
- [Interactive Shell](guide/shell.md) - Raw terminal mode and rz/sz
- [SCP & Rsync](guide/scp-rsync.md) - Compatible scp/rsync syntax
- [Configuration File](guide/config-file.md) - Manage multiple servers

### Advanced and Reference

- [Go SDK](advanced/sdk.md) - Use as a Go library
- [Best Practices](advanced/best-practices.md) - Security and efficiency tips
- [Changelog](changelog.md)

## Community

- [GitHub](https://github.com/chuccp/win-sshpass)
- [Issue Tracker](https://github.com/chuccp/win-sshpass/issues)
