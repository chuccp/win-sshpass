# Installation

## Requirements

- **OS**: Windows 10/11 (x64 or ARM64)
- **Zero dependencies**: No need to install OpenSSH or any other software

## Install

### Option 1: Scoop (Recommended)

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### Option 2: Download Release

Download the latest release from [GitHub Releases](https://github.com/chuccp/win-sshpass/releases):

| Architecture | Zip | MSI Installer |
|--------------|-----|---------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. Go to [Releases](https://github.com/chuccp/win-sshpass/releases) page
2. Download the zip or MSI for your architecture (x64 or ARM64)
3. If using MSI: run the installer — it will add the install directory to your system PATH automatically

### Option 3: Build from Source

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass
go build -o win-sshpass.exe ./cmd/sshpass
```

## Verify Installation

```bash
win-sshpass -v
# Output: sshpass version v0.3.2 (Windows)
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
