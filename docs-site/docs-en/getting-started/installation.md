# Installation

## Requirements

- **OS**: Windows 10/11 (x64, ARM64), Linux (amd64, arm64), or macOS (amd64, arm64)
- **Zero dependencies**: No need to install OpenSSH or any other software

## Install

### Option 1: WinGet (Recommended)

```bash
winget install chuccp.win-sshpass
```

### Option 2: Scoop

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### Option 3: Download Release

Download the latest release from [GitHub Releases](https://github.com/chuccp/win-sshpass/releases):

**Windows**

| Architecture | Zip | MSI Installer |
|--------------|-----|---------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

**Linux**

| Architecture | Tarball |
|--------------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

**macOS**

| Architecture | PKG Installer | Tarball |
|--------------|---------------|---------|
| **amd64 (Intel)** | `win-sshpass-*-darwin-amd64.pkg` | `win-sshpass-*-darwin-amd64.tar.gz` |
| **arm64 (Apple Silicon)** | `win-sshpass-*-darwin-arm64.pkg` | `win-sshpass-*-darwin-arm64.tar.gz` |

> The `.pkg` installer places the binary at `/usr/local/bin/win-sshpass` automatically.

1. Go to [Releases](https://github.com/chuccp/win-sshpass/releases) page
2. Download the package for your platform and architecture
3. **Windows MSI / macOS PKG**: run the installer — it will add the binary to your system PATH automatically
4. **Windows Zip / Linux tar.gz / macOS tar.gz**: extract and place the binary in your PATH

### Option 4: Build from Source

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass

# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass
```

## Verify Installation

```bash
win-sshpass -v
# Output: win-sshpass version v0.3.2 (windows/amd64)
```

## Dependencies

win-sshpass is a standalone executable with no external runtime dependencies. Go dependencies used during build:

| Dependency | Purpose |
|------------|---------|
| golang.org/x/crypto/ssh | SSH protocol implementation |
| github.com/pkg/sftp | SFTP file transfer |
| github.com/schollz/progressbar/v3 | CLI progress bar (CLI only) |
| github.com/ncruces/zenity | File dialogs (CLI only) |

## Next Steps

- [Quick Start](quick-start.md) - Your first SSH connection
