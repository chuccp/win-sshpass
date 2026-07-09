# 安裝

## 系統需求

- **作業系統**：Windows 10/11（x64 或 ARM64）
- **零依賴**：無需安裝 OpenSSH 或其他軟體

## 安裝方式

### 方式一：Scoop 安裝（推薦）

```bash
scoop bucket add chuccp https://github.com/chuccp/scoop-bucket
scoop install win-sshpass
```

### 方式二：下載發行包

從 [GitHub Releases](https://github.com/chuccp/win-sshpass/releases) 下載最新版本：

| 架構 | Zip 包 | MSI 安裝包 |
|------|--------|------------|
| **x64 (amd64)** | `win-sshpass-*-amd64.zip` | `win-sshpass-*-amd64.msi` |
| **ARM64** | `win-sshpass-*-arm64.zip` | `win-sshpass-*-arm64.msi` |

1. 前往 [Releases](https://github.com/chuccp/win-sshpass/releases) 頁面
2. 下載對應架構的 zip 或 MSI 檔案
3. 如果使用 MSI：執行安裝程式，會自動將安裝目錄新增到系統 PATH

### 方式三：從原始碼建構

```bash
git clone https://github.com/chuccp/win-sshpass.git
cd win-sshpass
go build -o win-sshpass.exe ./cmd/sshpass
```

## 驗證安裝

```bash
win-sshpass -v
# 輸出: sshpass version v0.3.2 (Windows)
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
