---
hide:
  - navigation
  - toc
---

# win-sshpass

> A cross-platform implementation of sshpass (Windows, Linux & macOS): password/key SSH login, interactive shell, SFTP/SCP/Rsync file transfer, SOCKS5/SOCKS4/HTTP proxy tunneling, breakpoint resume, file hash/verify, key generation, and a reusable Go SDK.

[Quick Start](getting-started/quick-start.md){ .md-button .md-button--primary }
[Installation](getting-started/installation.md){ .md-button }
[Source :simple-github:](https://github.com/chuccp/win-sshpass){ .md-button }

---

## :material-star: Features

- **:material-console: SSH & Shell** — Password or private-key authentication, remote command execution, and interactive shell with raw terminal mode (proper echo, Ctrl+C, vim/top support, dynamic resize).
- **:material-file-sync: File Transfer** — SFTP upload/download with progress bars, SCP-style and Rsync-style transfer syntax, multiple file support, recursive directory upload.
- **:material-cloud-download: Shell rz/sz** — Use `rz`/`sz` commands directly in interactive shell — no software needs to be installed on the remote server; transfers go over SFTP.
- **:material-shield-key: Key Generation** — Built-in SSH key pair generation (Ed25519 and RSA), no ssh-keygen needed. Deploy public keys via SSH to enable password-less login.
- **:material-lan-connect: Proxy Tunneling** — Tunnel SSH connections through SOCKS5 (with auth), SOCKS4, SOCKS4A, HTTP CONNECT, and HTTPS CONNECT proxies.
- **:material-reload: Breakpoint Resume** — Resume interrupted SFTP uploads/downloads from where they left off with the `-resume` flag.
- **:material-fingerprint: File Hash & Verify** — Compute and verify local file checksums (MD5, SHA-1, SHA-256, SHA-512) — no SSH connection needed.
- **:material-package-variant: Reusable Go SDK** — Import `package sshpass` to embed SSH/SFTP/shell in your own app with injectable I/O, progress callbacks, and zero UI dependencies.

---

## :material-rocket-launch: 30-Second Quick Start

```bash
# Install via WinGet
winget install chuccp.win-sshpass

# Or via Scoop
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

# Generate SSH key pair
win-sshpass keygen
```

---

## :material-compass: Quick Navigation

| | | |
|---|---|---|
| [:material-download: Installation](getting-started/installation.md) | [:material-rocket-launch: Quick Start](getting-started/quick-start.md) | [:material-console: SSH Connection](guide/ssh.md) |
| [:material-file-sync: File Transfer](guide/file-transfer.md) | [:material-monitor: Interactive Shell](guide/shell.md) | [:material-folder-multiple: SCP & Rsync](guide/scp-rsync.md) |
| [:material-cog: Configuration File](guide/config-file.md) | [:material-code-braces: Go SDK](advanced/sdk.md) | [:material-security: Best Practices](advanced/best-practices.md) |
| [:material-history: Changelog](changelog.md) | | |

---

## :material-layers: Dependencies

| Dependency | Purpose |
|---|---|
| `golang.org/x/crypto/ssh` | SSH protocol implementation |
| `github.com/pkg/sftp` | SFTP file transfer |
| `github.com/schollz/progressbar/v3` | CLI progress bar (CLI only) |
| `github.com/ncruces/zenity` | File dialogs for rz/sz (CLI only, optional) |

win-sshpass is a **standalone executable** — no external runtime dependencies. Download and run.

---

## :simple-github: Community

- [GitHub Repository](https://github.com/chuccp/win-sshpass)
- [Issue Tracker](https://github.com/chuccp/win-sshpass/issues)
- [Releases](https://github.com/chuccp/win-sshpass/releases)
- [Changelog](changelog.md)
