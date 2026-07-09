# 安装

## 系统要求

- **操作系统**：Windows 10/11（x64 或 ARM64）
- **零依赖**：无需安装 OpenSSH 或其他软件

## 安装方式

### 方式一：Scoop 安装（推荐）

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方式二：下载发行包

从 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下载最新版本：

| 架构 | Zip 包 | MSI 安装包 |
|------|--------|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. 前往 [Releases](https://github.com/chuccp/win-sshpass/releases) 页面
2. 下载对应架构的 zip 或 MSI 文件
3. 如果使用 MSI：运行安装程序，会自动将安装目录添加到系统 PATH

### 方式三：从源码构建

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass
go build -o win-sshpass.exe ./cmd/sshpass
```

## 验证安装

```bash
win-sshpass -v
# 输出: sshpass version v0.3.2 (Windows)
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
