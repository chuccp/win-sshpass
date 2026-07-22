# 安装

## 系统要求

- **操作系统**：Windows 10/11（x64、ARM64）、Linux（amd64、arm64）或 macOS（amd64、arm64）
- **零依赖**：无需安装 OpenSSH 或其他软件

## 安装方式

### 方式一：WinGet 安装（推荐）

```bash
winget install chuccp.win-sshpass
```

### 方式二：Scoop 安装

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方式三：下载发行包

从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下载最新版本：

**Windows**

| 架构 | Zip 包 | MSI 安装包 |
|------|--------|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

**Linux**

| 架构 | Tarball |
|------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

**macOS**

| 架构 | PKG 安装包 | Tarball |
|------|-----------|---------|
| **amd64 (Intel)** | `win-sshpass-*-darwin-amd64.pkg` | `win-sshpass-*-darwin-amd64.tar.gz` |
| **arm64 (Apple Silicon)** | `win-sshpass-*-darwin-arm64.pkg` | `win-sshpass-*-darwin-arm64.tar.gz` |

> `.pkg` 安装包会自动将二进制文件安装到 `/usr/local/bin/win-sshpass`。

1. 前往 [Releases](https://github.com/chuccp/win-sshpass/releases) 页面
2. 下载对应平台和架构的安装包
3. **Windows MSI / macOS PKG**：运行安装程序，二进制文件会自动添加到系统 PATH
4. **Windows Zip / Linux tar.gz / macOS tar.gz**：解压后将二进制文件放入 PATH 目录

### 方式四：从源码构建

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass

# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass
```

## 验证安装

```bash
win-sshpass -v
# 输出: win-sshpass version v0.3.2 (windows/amd64)
```

## 依赖说明

win-sshpass 是独立的可执行文件，无外部依赖。构建时使用的 Go 依赖：

| 依赖 | 用途 |
|------|------|
| golang.org/x/crypto/ssh | SSH 协议实现 |
| github.com/pkg/sftp | SFTP 文件传输 |
| github.com/schollz/progressbar/v3 | CLI 进度条（仅 CLI 使用） |
| github.com/ncruces/zenity | 文件对话框（仅 CLI 使用） |

## 下一步

- [快速开始](quick-start.md) - 第一个 SSH 连接
