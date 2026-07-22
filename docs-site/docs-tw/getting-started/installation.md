# 安裝

## 系統需求

- **作業系統**：Windows 10/11（x64、ARM64）、Linux（amd64、arm64）或 macOS（amd64、arm64）
- **零依賴**：無需安裝 OpenSSH 或其他軟體

## 安裝方式

### 方式一：WinGet 安裝（推薦）

```bash
winget install chuccp.win-sshpass
```

### 方式二：Scoop 安裝

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方式三：下載發行包

從 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下載最新版本：

**Windows**

| 架構 | Zip 包 | MSI 安裝包 |
|------|--------|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

**Linux**

| 架構 | Tarball |
|------|---------|
| **amd64** | `win-sshpass-*-linux-amd64.tar.gz` |
| **arm64** | `win-sshpass-*-linux-arm64.tar.gz` |

**macOS**

| 架構 | PKG 安裝包 | Tarball |
|------|-----------|---------|
| **amd64 (Intel)** | `win-sshpass-*-darwin-amd64.pkg` | `win-sshpass-*-darwin-amd64.tar.gz` |
| **arm64 (Apple Silicon)** | `win-sshpass-*-darwin-arm64.pkg` | `win-sshpass-*-darwin-arm64.tar.gz` |

> `.pkg` 安裝包會自動將二進位檔案安裝到 `/usr/local/bin/win-sshpass`。

1. 前往 [Releases](https://github.com/chuccp/win-sshpass/releases) 頁面
2. 下載對應平台和架構的安裝包
3. **Windows MSI / macOS PKG**：執行安裝程式，二進位檔案會自動新增到系統 PATH
4. **Windows Zip / Linux tar.gz / macOS tar.gz**：解壓後將二進位檔案放入 PATH 目錄

### 方式四：從原始碼建構

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass

# Windows
go build -o win-sshpass.exe ./cmd/sshpass

# Linux / macOS
go build -o win-sshpass ./cmd/sshpass
```

## 驗證安裝

```bash
win-sshpass -v
# 輸出: win-sshpass version v0.3.2 (windows/amd64)
```

## 依賴說明

win-sshpass 是獨立的可執行檔，無外部執行時依賴。建構時使用的 Go 依賴：

| 依賴 | 用途 |
|------|------|
| golang.org/x/crypto/ssh | SSH 協定實作 |
| github.com/pkg/sftp | SFTP 檔案傳輸 |
| github.com/schollz/progressbar/v3 | CLI 進度條（僅 CLI 使用） |
| github.com/ncruces/zenity | 檔案對話框（僅 CLI 使用） |

## 下一步

- [快速開始](quick-start.md) - 第一個 SSH 連線
