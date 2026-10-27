# Web Console & MCP

`win-sshpass -web` starts a local SSH workspace in the browser. A host supplied
with `-h` or `-f` is required. The workspace includes SFTP file management, a
real terminal, and MCP tool access.

## Start

Start with connection details for automatic login:

```bash
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p 'password'
```

Open <http://127.0.0.1:6326>. The page connects automatically and opens the
workspace.

Running bare `win-sshpass -web` is rejected. Use a config file when connecting
with saved settings:

```bash
win-sshpass -f server.config -web
```

Change the listener deliberately:

```bash
win-sshpass -web -h example.com -u root -p 'password' -web-addr 127.0.0.1:6326
```

The default listener is loopback-only. Binding to another interface exposes
the Web workspace to that network.

## Workspace

### SFTP file manager

The left side is an interactive SFTP browser:

- enter a path or move to the parent directory
- open directories by double-clicking
- upload and download files
- create directories
- delete files and empty directories
- drag local files or folders onto the pane to upload
- drag remote files out of the browser to download
- follow aggregate upload progress with filename and percentage

Directory drops recursively create remote folders. File actions run with the
permissions of the SSH account.

### SSH terminal

The right side is a real SSH PTY rendered by xterm.js. It supports:

- interactive input and ANSI output
- shell history, tab completion, and full-screen terminal programs
- automatic terminal resize
- reconnect and clear-terminal actions

This is not a sequence of command API requests. A WebSocket carries terminal
bytes between xterm.js and the remote shell.

### Resizable layout

Drag the center divider to resize the file manager and terminal. When the
divider has keyboard focus, use Left and Right arrows. The width is stored in
the browser.

### Themes

The theme selector supports:

- **System** — follows `prefers-color-scheme`; default
- **Light**
- **Dark**

The xterm color palette changes with the page theme.

## Session security

The browser receives a random session ID in an `HttpOnly`,
`SameSite=Strict` cookie. The cookie contains no password, username, or host.
The SSH client and credentials remain in the local server process. Sessions
expire after 30 minutes of inactivity.

The HTTP APIs require JSON content and reject cross-origin operations.

## MCP

The same server exposes a Streamable HTTP MCP endpoint:

```text
http://127.0.0.1:6326/mcp
```

MCP is enabled by default. Start the server with connection flags so MCP can
create its SSH session.

### Codex configuration

```toml
[mcp_servers.win-sshpass]
url = "http://127.0.0.1:6326/mcp"
```

The initialize response includes a `Mcp-Session-Id` header. Send that header
with later tool requests. Use `DELETE /mcp` to end the session.

### Tools

| Tool | Purpose |
|---|---|
| `ssh_exec` | Run an SSH command; return stdout, stderr, and exit code |
| `sftp_list` | List a remote directory |
| `sftp_read` | Read UTF-8 text from a remote file |
| `sftp_write` | Create/overwrite using text or base64 content |
| `sftp_mkdir` | Create a remote directory and parents |
| `sftp_delete` | Delete a file or empty directory |

Example tool call:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "ssh_exec",
    "arguments": {
      "command": "docker ps"
    }
  }
}
```

## Reverse proxy

If TLS terminates at a reverse proxy, ensure it forwards WebSocket upgrades for
`/api/terminal` and the `Mcp-Session-Id` header. Keep the default loopback
binding unless remote access is intentional.

## Troubleshooting

- **Page stays on Connecting** — verify the SSH host, port, password, key, and
  proxy with the regular CLI before starting Web mode.
- **Terminal disconnects** — use Reconnect terminal; inspect the server log for
  SSH or proxy errors.
- **File operations fail** — confirm the SSH account can access the path and
  that SFTP is enabled on the server.
- **MCP tools/list fails** — complete MCP initialization and preserve the
  `Mcp-Session-Id` header.
