# Interactive Shell

When no remote command is specified, win-sshpass opens an interactive shell with raw terminal mode, providing a near-native SSH experience.

## Basic Usage

```bash
# With password
win-sshpass -p 'password' ssh user@host

# With private key
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# With config file
win-sshpass -f server.config
```

## Raw Terminal Mode Features

### Proper Echo

Typed characters are displayed correctly without double echo. This is achieved by setting the local terminal to raw mode.

### Signal Forwarding

- **Ctrl+C** — interrupts the current remote process
- **Ctrl+Z** — suspends the current remote process

### Full-Screen App Support

The following applications work correctly:

- `vim` / `nvim` — text editors
- `top` / `htop` — system monitors
- `nano` — text editor
- `less` / `more` — pagers
- `mc` (Midnight Commander) — file manager

### Dynamic Terminal Resizing

When you resize your local terminal window, the remote terminal automatically matches the new size. This is implemented via SSH `window-change` requests.

### Tab Completion

Remote shell tab completion works as expected.

## File Transfer (rz/sz)

In interactive shell, you can use `rz` and `sz` commands for file transfer:

```bash
# In remote shell:
rz                              # Upload file (opens file picker)
rz /local/path/to/file          # Upload specific local file
sz /remote/path/to/file         # Download file (opens save dialog)
sz /remote/path/to/file /local  # Download to specific local path
```

!!! info "How It Works"
    When the remote shell reports `rz`/`sz: command not found`, win-sshpass intercepts the error and performs the transfer over SFTP. No need to install lrzsz on the remote server.

### Custom File Selector

By default, rz/sz uses system file dialogs (via zenity). To customize, inject your own implementation via the Go SDK's `WithFileSelector` option.

## Differences from Standard SSH

| Feature | Standard SSH | win-sshpass |
|---------|-------------|-------------|
| Password auth | Requires ssh-agent | Native support |
| Windows support | Requires OpenSSH install | Standalone executable |
| rz/sz transfer | Requires remote lrzsz | Built-in SFTP fallback |
| Progress bar | None | SFTP transfers |

## Next Steps

- [File Transfer](file-transfer.md) - Direct SFTP transfer
- [Go SDK](../advanced/sdk.md) - Use shell programmatically
