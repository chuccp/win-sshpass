# win-sshpass

[![GitHub release](https://badgen.net/github/release/chuccp/win-sshpass/include-prereleases)](https://github.com/chuccp/win-sshpass/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/chuccp/win-sshpass)](https://goreportcard.com/report/github.com/chuccp/win-sshpass)
[![License](https://badgen.net/badge/License/Apache%202.0/blue)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chuccp/win-sshpass)](go.mod)
[![Downloads](https://img.shields.io/github/downloads/chuccp/win-sshpass/total)](https://github.com/chuccp/win-sshpass/releases)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [日本語](README.ja.md)

A Windows implementation of sshpass, providing similar functionality to the Linux sshpass tool.

> 💡 **Like this project?** Give it a ⭐ Star — it helps others discover the tool!

## Features

- SSH login with password or private key authentication
- Execute remote commands or open interactive shell
- File upload/download via SFTP (with progress bar)
- SCP-style and Rsync-style file transfer
- Config file support for managing multiple servers
- Interactive shell with raw terminal mode (proper echo, Ctrl+C, and full-screen app support)
- Dynamic terminal resizing in interactive shell mode
- Git Bash path conversion detection and auto-fix
- IPv6 address support
- Support for both x64 (amd64) and ARM64 architectures

## Download

Download the latest release from [GitHub Releases](https://github.com/chuccp/win-sshpass/releases):

| Architecture | Zip | MSI Installer |
|--------------|-----|---------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. Go to [Releases](https://github.com/chuccp/win-sshpass/releases) page
2. Download the zip or MSI for your architecture (x64 or ARM64)
3. If using MSI: run the installer — it will add the install directory to your system PATH automatically

> **Zero dependencies**: `win-sshpass.exe` is a standalone binary. No need to install OpenSSH or any other software. Download it, put it in your PATH, and you're ready to go.

### Install via Scoop

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

## Quick Start

```bash
# Password login and execute command
win-sshpass -p 'password' ssh user@example.com 'whoami'

# Private key login and execute command
win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com 'hostname'

# Upload file
win-sshpass -h example.com -p 'password' -local file.txt -remote /tmp/file.txt

# Download file
win-sshpass -h example.com -p 'password' -d -remote /tmp/file.txt -local ./file.txt
```

## Interactive Shell

When no command is specified, `win-sshpass` opens an interactive shell with **raw terminal mode**:

```bash
win-sshpass -p 'password' ssh user@host
```

**Raw terminal mode** features:

- **Proper echo** — typed characters are displayed correctly (no double echo)
- **Ctrl+C / Ctrl+Z** — signals are forwarded to the remote process
- **Full-screen apps** — vim, top, htop, nano, etc. work correctly
- **Dynamic terminal resizing** — the remote terminal automatically matches your local window size
- **Tab completion** — remote shell tab completion works as expected

### File Transfer in Interactive Shell

While connected, use `rz` / `sz` commands to transfer files (no need to install anything on the remote server):

```bash
# Upload file to remote current directory (opens file picker)
rz

# Upload specific local file
rz /local/path/to/file

# Download remote file (opens save dialog)
sz /remote/path/to/file

# Download remote file to specific local path
sz /remote/path/to/file /local/save/path
```

> **How it works**: When the remote shell reports `rz`/`sz: command not found`, the tool intercepts it and performs the transfer over SFTP instead. Files and directories are both supported, with progress bars.

## Command Format

### SSH Login

```bash
# Password authentication
win-sshpass -p <password> ssh [user@host] [command]
win-sshpass -p <password> ssh -p <port> user@host 'command'
win-sshpass -p <password> ssh -o StrictHostKeyChecking=no user@host

# Interactive shell (raw terminal mode: proper echo, Ctrl+C, vim/top support)
win-sshpass -p <password> ssh user@host

# Private key authentication
win-sshpass -i <private_key_path> ssh [user@host] [command]

# Password from environment variable
SSHPASS=<password> win-sshpass -e ssh user@host

# Password from file
echo 'password' > pass.txt
win-sshpass -f pass.txt ssh user@host

# Configuration file (multi-line format)
win-sshpass -f server.config
```

### File Transfer

> **Git Bash users**: Use `//` prefix for remote paths, e.g. `-remote //tmp/file.txt`. See [Git Bash Notes](#git-bash-notes) below.

```bash
# Upload file
win-sshpass -h <host> -p <password> -local <local_path> -remote <remote_path>

# Upload multiple files (comma-separated)
win-sshpass -h <host> -p <password> -local "a.txt,b.txt,c.txt" -remote //tmp/

# Upload multiple files (space-separated, only for simple paths without / or \)
win-sshpass -h <host> -p <password> -local "a.txt b.txt c.txt" -remote //tmp/

# Upload directory (auto-recursive)
win-sshpass -h <host> -p <password> -local <local_dir> -remote <remote_dir>

# Download file/directory
win-sshpass -h <host> -p <password> -d -remote <remote_path> -local <local_path>
```

### SCP Style

```bash
# Upload file
win-sshpass -p <password> scp <local_file> user@host:<remote_path>
win-sshpass -p <password> scp -P <port> <local_file> user@host:<remote_path>

# Upload directory
win-sshpass -p <password> scp -r <local_dir> user@host:<remote_path>

# Download file/directory
win-sshpass -p <password> scp user@host:<remote_file> <local_path>
```

### Rsync Style

```bash
# Upload
win-sshpass -p <password> rsync -avz <local_path> user@host:<remote_path>

# Download
win-sshpass -p <password> rsync -avz user@host:<remote_path> <local_path>
```

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
| `-remote` | Remote path (upload/download) | `-remote /tmp/file.txt` |
| `-d` | Download mode | `-d` |
| `-k` | Enable strict host key verification | `-k` |
| `-t` | Total operation timeout in seconds (0 = no limit) | `-t 30` |
| `-ct` | TCP connection timeout in seconds (default: 10) | `-ct 5` |
| `-retry` | Total connection attempts (default: 3) | `-retry 5` |
| `-v` | Show version | `-v` |
| `-help` | Show help message | `-help` |

## Configuration File Format

```yaml
host: example.com
username: root
password: your_password
port: 22
# key: ~/.ssh/id_ed25519  # optional, use private key instead of password
# timeout: 0              # optional, total operation timeout in seconds (0 = no limit)
# connect_timeout: 10     # optional, TCP connection timeout in seconds
# strict_host_key: false  # optional, enable strict host key verification
```

Usage:
```bash
win-sshpass -f server.config -c 'ls -la'
win-sshpass -f server.config 'ls -la'
```

## Complete Examples

```bash
# 1. Password login and execute command
win-sshpass -p 'mypass' ssh root@192.168.1.100 'docker ps'

# 2. Private key login and execute sudo command
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server.com 'sudo systemctl restart nginx'

# 3. Upload entire directory to server
win-sshpass -h server.com -p 'mypass' -local ./dist -remote //var/www/html

# 4. Download server log directory
win-sshpass -h server.com -p 'mypass' -d -remote //var/log/nginx -local ./logs

# 5. SCP upload file
win-sshpass -p 'mypass' scp ./app.jar user@server.com:/opt/app/

# 6. Password via environment variable (more secure)
export SSHPASS='mypass'
win-sshpass -e ssh user@server.com 'whoami'

# 7. Operation timeout (abort after 30 seconds)
win-sshpass -p 'mypass' -t 30 ssh user@server.com 'long-running-command'

# 8. Config file with positional command
win-sshpass -f server.config 'docker ps'
```

## Git Bash Notes

Use `//` prefix for remote paths to avoid path conversion:
```bash
# Wrong: /tmp will be converted to Windows path
win-sshpass ... -remote /tmp/file.txt

# Correct: use double slashes
win-sshpass ... -remote //tmp/file.txt
```

## Use as a Go SDK

`win-sshpass` is also a reusable Go library (`package sshpass`). Import it to
embed SSH/SFTP/shell capabilities in your own application:

```bash
go get github.com/chuccp/win-sshpass
```

```go
package main

import (
	"log"

	sshpass "github.com/chuccp/win-sshpass"
)

func main() {
	cfg := sshpass.NewConfig()
	cfg.Host = "example.com"
	cfg.User = "root"
	cfg.Password = "secret"

	// NewClient dials and returns a ready-to-use client.
	client, err := sshpass.NewClient(cfg, sshpass.WithSignalHandler())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Execute a command (output streams to os.Stdout/os.Stderr by default).
	if err := client.Exec("uname -a"); err != nil {
		log.Fatal(err)
	}

	// Upload a file over SFTP.
	sftp, err := client.SFTP()
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()
	if err := sftp.Upload("./local.txt", "/tmp/remote.txt"); err != nil {
		log.Fatal(err)
	}
}
```

### Customization options

Behavior is configured through functional options passed to `NewClient`:

| Option | Purpose |
|--------|---------|
| `WithStdin(r)` / `WithStdout(w)` / `WithStderr(w)` | Redirect I/O streams (default `os.Stdin`/`os.Stdout`/`os.Stderr`). |
| `WithProgress(fn)` | Set a `ProgressFunc` callback that receives `(description string, sent, total int64)` during SFTP transfers. The SDK performs no rendering — callers display progress however they like. Defaults to none (headless-friendly). |
| `WithFileSelector(s)` | Set the `FileSelector` used by the rz/sz shell-transfer fallback. The SDK ships no default implementation; without one, rz/sz prompts for a path on stdin. |
| `WithSignalHandler()` | Register a Ctrl+C handler that closes the connection. Off by default so the library never interferes with host signal handling. |

The SDK intentionally bundles **no UI code** (no progress bar, no file dialog).
Those concerns live in the CLI package (`cmd/sshpass/ui.go`), which wires a
progressbar-based `ProgressFunc` and a zenity-based `FileSelector` into the
client. Library users provide their own.

Lower-level helpers are also exported for advanced use: `Dial`, `NewConfig`,
`LoadConfig`, `LoadConfigOrPasswordFile`, `ParseSSHArgs`, `ParseSCPArgs`,
`ParseRsyncArgs`, `DetectCommandType`, `RunSCP`, `RunRsync`, `CleanRemotePath`,
`SplitPaths`, `ParseUserHostPath`, `ExitCodeFromError`.

## Build

```bash
go build -o win-sshpass.exe ./cmd/sshpass
```

## Dependencies

- Go 1.23+
- golang.org/x/crypto/ssh
- github.com/pkg/sftp
- github.com/schollz/progressbar/v3

## Related Project

[**go-web-frame**](https://github.com/chuccp/go-web-frame) — A Go web framework with dependency injection, generics-based data access, and production-ready utilities. Ships with HTTP routing, caching, logging, Redis, and more out of the box.