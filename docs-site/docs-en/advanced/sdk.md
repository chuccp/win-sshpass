# Go SDK

win-sshpass is not only a command-line tool but also a reusable Go library (`package sshpass`). You can embed SSH/SFTP/Shell capabilities into your own application.

## Installation

```bash
go get github.com/chuccp/win-sshpass
```

## Quick Start

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
}
```

## Client Core Methods

### NewClient

Establishes an SSH connection and returns a client:

```go
client, err := sshpass.NewClient(cfg, opts...)
```

- `cfg`: Connection configuration (`*sshpass.Config`)
- `opts`: Optional functional options (`...sshpass.Option`)
- If `cfg.Timeout > 0`, an operation timer is armed that closes the connection when the deadline elapses

### Exec

Executes a single remote command:

```go
err := client.Exec("ls -la")
```

- Output streams are configured via `WithStdout`/`WithStderr` (default: `os.Stdout`/`os.Stderr`)
- Input stream is configured via `WithStdin` (default: `os.Stdin`)

### Shell

Starts an interactive shell:

```go
err := client.Shell()
```

- Auto-detects terminal, supports PTY and raw mode
- Supports dynamic terminal resizing
- Supports rz/sz file transfer (requires `WithFileSelector` configuration)

### SFTP

Opens an SFTP sub-channel:

```go
sftp, err := client.SFTP()
if err != nil {
    log.Fatal(err)
}
defer sftp.Close()

// Upload file
err = sftp.Upload("./local.txt", "/tmp/remote.txt")

// Download file
err = sftp.Download("/tmp/remote.txt", "./local.txt")

// Access underlying *sftp.Client (advanced usage)
rawClient := sftp.SFTP()
```

### Close

Closes the SSH connection:

```go
err := client.Close()
```

- Idempotent — multiple calls return the same error
- Automatically stops timer and signal handler

### TimedOut

Check if the failure was due to timeout:

```go
if client.TimedOut() {
    fmt.Println("Operation timed out")
}
```

## Configuration (Config)

```go
cfg := sshpass.NewConfig()
cfg.Host = "example.com"
cfg.User = "root"
cfg.Password = "secret"
cfg.Port = "22"
cfg.KeyPath = "~/.ssh/id_ed25519"
cfg.StrictHostKey = true
cfg.Timeout = 30          // Operation timeout (seconds), 0 = no limit
cfg.ConnectTimeout = 10   // TCP connection timeout (seconds)
cfg.Retries = 3           // Connection retry count
```

### Load from Config File

```go
cfg, err := sshpass.LoadConfig("server.config")
```

### Load Config or Password File

```go
cfg, pass, err := sshpass.LoadConfigOrPasswordFile("file.txt", "", false)
if cfg != nil {
    // It's a config file
} else {
    // It's a password file; pass contains the password
}
```

## Functional Options

Configure Client behavior through functional options:

### I/O Streams

```go
// Redirect input/output
client, err := sshpass.NewClient(cfg,
    sshpass.WithStdin(myReader),
    sshpass.WithStdout(myWriter),
    sshpass.WithStderr(myWriter),
)
```

### Progress Callback

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithProgress(func(desc string, sent, total int64) {
        fmt.Printf("\r%s %d/%d bytes", desc, sent, total)
    }),
)
```

- `desc`: Transfer description (e.g., "Uploading file.txt")
- `sent`: Bytes transferred so far
- `total`: Total file size
- The SDK performs no rendering — callers display progress however they like

### File Selector

```go
type myFileSelector struct{}

func (s myFileSelector) OpenFile() (string, error) {
    // Implement file open dialog
    return "/path/to/file", nil
}

func (s myFileSelector) SaveFile(defaultName string) (string, error) {
    // Implement file save dialog
    return "/path/to/save", nil
}

client, err := sshpass.NewClient(cfg,
    sshpass.WithFileSelector(myFileSelector{}),
)
```

Used for rz/sz shell transfer file selection. The SDK ships no default implementation.

### Signal Handler

```go
client, err := sshpass.NewClient(cfg,
    sshpass.WithSignalHandler(),
)
```

Registers a Ctrl+C handler that closes the connection. Off by default so the library never interferes with host signal handling.

## Low-Level API

### Dial

Create an SSH client connection directly:

```go
sshClient, err := sshpass.Dial(cfg)
```

Returns `*ssh.Client` for scenarios requiring more low-level control.

### Argument Parsing

```go
// Parse SSH arguments
config, cmd := sshpass.ParseSSHArgs([]string{"ssh", "user@host", "ls"})

// Parse SCP arguments
config, args := sshpass.ParseSCPArgs([]string{"scp", "file.txt", "user@host:/tmp/"})

// Parse Rsync arguments
config, args := sshpass.ParseRsyncArgs([]string{"rsync", "-avz", "./", "user@host:/backup/"})

// Detect command type
cmdType := sshpass.DetectCommandType(args)
```

### Utility Functions

```go
// Run SCP transfer
err := sshpass.RunSCP(client, args)

// Run Rsync transfer
err := sshpass.RunRsync(client, args)

// Clean remote path (handle Git Bash path conversion)
path, err := sshpass.CleanRemotePath("//tmp/file.txt")

// Parse user@host:path format
user, host, path := sshpass.ParseUserHostPath("user@host:/tmp/file.txt")

// Split paths (comma or space separated)
paths, err := sshpass.SplitPaths("a.txt,b.txt,c.txt", "local")

// Extract exit code from error
code, ok := sshpass.ExitCodeFromError(err)
```

## Complete Examples

### Batch Command Execution

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    hosts := []string{"192.168.1.101", "192.168.1.102", "192.168.1.103"}

    for _, host := range hosts {
        cfg := sshpass.NewConfig()
        cfg.Host = host
        cfg.User = "root"
        cfg.Password = "secret"

        client, err := sshpass.NewClient(cfg)
        if err != nil {
            log.Printf("[%s] Connection failed: %v", host, err)
            continue
        }

        fmt.Printf("[%s] Executing command...\n", host)
        if err := client.Exec("uptime"); err != nil {
            log.Printf("[%s] Command failed: %v", host, err)
        }
        client.Close()
    }
}
```

### File Upload with Progress

```go
package main

import (
    "fmt"
    "log"

    sshpass "github.com/chuccp/win-sshpass"
)

func main() {
    cfg := sshpass.NewConfig()
    cfg.Host = "example.com"
    cfg.User = "root"
    cfg.Password = "secret"

    client, err := sshpass.NewClient(cfg,
        sshpass.WithProgress(func(desc string, sent, total int64) {
            pct := sent * 100 / total
            fmt.Printf("\r%s %d%%", desc, pct)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    sftp, err := client.SFTP()
    if err != nil {
        log.Fatal(err)
    }
    defer sftp.Close()

    if err := sftp.Upload("./large-file.zip", "/tmp/large-file.zip"); err != nil {
        log.Fatal(err)
    }
    fmt.Println("\nUpload complete!")
}
```

## Next Steps

- [Best Practices](best-practices.md) - Security and efficiency tips
