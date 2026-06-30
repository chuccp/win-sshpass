# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build

```bash
go build -o win-sshpass.exe ./cmd/sshpass
```

## Architecture

The project is a reusable Go SDK (`package sshpass`) plus a CLI entry point.

- `cmd/sshpass/main.go` - CLI entry point: flag parsing, config merging, command dispatch. Built into `win-sshpass.exe`.
- `client.go` - `Client` object: `NewClient`, `Exec`, `Shell`, `SFTP`, `Close`, `TimedOut`.
- `options.go` - Functional options (`WithStdin`, `WithStdout`, `WithStderr`, `WithProgress`, `WithFileSelector`, `WithSignalHandler`) and the `ProgressFunc`/`FileSelector` abstractions. The SDK ships **no UI implementations**; CLI-side adapters (progressbar, zenity) live in `cmd/sshpass/ui.go`.
- `config.go` - `Config` struct, `NewConfig`, `LoadConfig`, `LoadConfigOrPasswordFile`, merge/validate/normalize methods.
- `ssh.go` - `Dial` (exported, alias `SSHClient`), retry/timeout/known_hosts, `runShell`/`executeCommand` (use injected I/O).
- `sftp.go` - `SFTPClient` with `Upload`/`Download`, timeout-aware readers/writers, progress reporting.
- `scp.go` - `RunSCP`/`RunRsync` over a `*Client`.
- `args.go` - `ParseSSHArgs`/`ParseSCPArgs`/`ParseRsyncArgs`/`DetectCommandType`.
- `shell_transfer.go` - rz/sz monitor using `FileSelector` and injected I/O.
- `util.go` - path helpers, `ParseUserHostPath`, `CleanRemotePath` (returns error), `SplitPaths`, `ExitCodeFromError`, `setupOperationTimeout`.
- `version.go` - exported `Version`.

The library avoids process-level side effects: no `os.Exit`, no global signal
registration (unless `WithSignalHandler` is used), and all I/O streams are
injectable via options.

## Dependencies

- `golang.org/x/crypto/ssh` - SSH protocol
- `github.com/pkg/sftp` - SFTP file transfer

## Release

Push a `v*` tag to trigger GitHub Actions workflow that builds `sshpass.exe` and creates a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Git Bash Path Note

Remote paths starting with `/` get converted by Git Bash. Use `//` prefix (e.g., `//root/file.txt`) to avoid this.