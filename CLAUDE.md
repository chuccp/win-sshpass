# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

```bash
# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass

# Cross-compile
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o win-sshpass ./cmd/sshpass
GOOS=windows GOARCH=amd64               go build -o win-sshpass.exe ./cmd/sshpass
GOOS=darwin  GOARCH=arm64               go build -o win-sshpass ./cmd/sshpass
```

## Architecture

The project is a reusable Go SDK (`package sshpass`) plus a CLI entry point.

- `cmd/sshpass/main.go` - CLI entry point: flag parsing, config merging, command dispatch.
- `cmd/sshpass/ui.go` - CLI progress bar adapter (shared, all platforms).
- `cmd/sshpass/ui_windows.go` - rz/sz file dialog via zenity (Windows).
- `cmd/sshpass/ui_darwin.go` - rz/sz file dialog via zenity/osascript (macOS, Finder native).
- `cmd/sshpass/ui_other.go` - rz/sz no-op file selector, falls back to stdin prompt (Linux).
- `client.go` - `Client` object: `NewClient`, `Exec`, `Shell`, `SFTP`, `Close`, `TimedOut`.
- `options.go` - Functional options (`WithStdin`, `WithStdout`, `WithStderr`, `WithProgress`, `WithFileSelector`, `WithSignalHandler`, `WithResume`) and the `ProgressFunc`/`FileSelector` abstractions. The SDK ships **no UI implementations**; CLI-side adapters (progressbar, zenity) live in `cmd/sshpass/ui*.go`.
- `config.go` - `Config` struct, `NewConfig`, `LoadConfig`, `LoadConfigOrPasswordFile`, merge/validate/normalize methods.
- `ssh.go` - `Dial` (exported, alias `SSHClient`), retry/timeout/known_hosts, `runShell`/`executeCommand` (use injected I/O).
- `ssh_resize_unix.go` - terminal resize via SIGWINCH (Linux/macOS).
- `ssh_resize_windows.go` - terminal resize via polling (Windows).
- `sftp.go` - `SFTPClient` with `Upload`/`Download`, timeout-aware readers/writers, progress reporting, breakpoint resume.
- `scp.go` - `RunSCP`/`RunRsync` over a `*Client`.
- `args.go` - `ParseSSHArgs`/`ParseSCPArgs`/`ParseRsyncArgs`/`DetectCommandType`.
- `shell_transfer.go` - rz/sz monitor using `FileSelector` and injected I/O.
- `util.go` - path helpers, `ParseUserHostPath`, `CleanRemotePath` (returns error), `SplitPaths`, `ExitCodeFromError`, `setupOperationTimeout`.
- `proxy.go` - `proxyDial`: SOCKS5 (via golang.org/x/net/proxy), SOCKS4/SOCKS4A (inline), and HTTP/HTTPS CONNECT proxy tunneling. Used by `dialAndHandshake` when `Config.ProxyURL` is set.
- `hash.go` - `HashFile`/`VerifyFile`: local file hash computation and verification (MD5, SHA-1, SHA-256, SHA-512).
- `keygen.go` - `GenerateKeyPair`, `GenerateRSAKeyPair`, `SaveKeyPair`, `DeployPublicKey`, `DefaultKeyPath`: SSH key pair generation (Ed25519, RSA).
- `version.go` - exported `Version`.

The library avoids process-level side effects: no `os.Exit`, no global signal
registration (unless `WithSignalHandler` is used), and all I/O streams are
injectable via options.

## Docker Testing

A local Docker-based integration test suite is in `docker-test/`:

```bash
# Start the SSH test server
cd docker-test
docker compose up -d ssh-server

# Build & run all integration tests (71 tests covering all features)
cd ..
go build -o win-sshpass.exe ./cmd/sshpass
./docker-test/test_all.sh
```

The Docker image pre-deploys a test Ed25519 key pair (`docker-test/test_key` /
`docker-test/test_key.pub`) into `testuser` and `root` authorized_keys.

The test script uses `127.0.0.1:10809` as the SOCKS5 proxy endpoint
(configurable via `SOCKS5_PROXY` env var).

## Platform Support

- **Windows**: full support including zenity file dialogs for rz/sz.
- **macOS**: full support including native Finder file dialogs for rz/sz (via zenity/osascript).
- **Linux**: full support; rz/sz falls back to stdin path input (no GUI dialog).
- Terminal resize uses SIGWINCH on Unix, polling on Windows.
- SDK (`package sshpass`) is fully cross-platform; platform-specific code is
  isolated behind build tags in `cmd/sshpass/` and `ssh_resize_*.go`.

## Dependencies

- `golang.org/x/crypto/ssh` - SSH protocol
- `github.com/pkg/sftp` - SFTP file transfer
- `github.com/ncruces/zenity` - file dialogs (Windows/macOS, build-tagged)
- `github.com/schollz/progressbar/v3` - CLI progress bar (all platforms)

## Release

Push a `v*` tag to trigger GitHub Actions workflow that builds Windows (zip + MSI),
Linux (tar.gz), and macOS (tar.gz) binaries and creates a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Git Bash Path Note

Remote paths starting with `/` get converted by Git Bash. Use `//` prefix (e.g., `//root/file.txt`) to avoid this.
