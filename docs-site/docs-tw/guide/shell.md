# 互動式 Shell

當不指定遠端命令時，win-sshpass 會開啟一個互動式 Shell，採用原始終端模式，提供接近原生 SSH 的體驗。

## 基本使用

```bash
# 使用密碼
win-sshpass -p 'password' ssh user@host

# 使用私鑰
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 使用設定檔
win-sshpass -f server.config
```

## 原始終端模式特性

### 正確的回顯

輸入的字元會正確顯示，不會出現雙重回顯問題。這得益於本地終端被設定為 raw 模式。

### 訊號轉發

- **Ctrl+C** — 中斷當前遠端程序
- **Ctrl+Z** — 暫停當前遠端程序

### 全螢幕應用支援

以下應用可以正常使用：

- `vim` / `nvim` — 文字編輯器
- `top` / `htop` — 系統監控
- `nano` — 文字編輯器
- `less` / `more` — 分頁檢視
- `mc` (Midnight Commander) — 檔案管理器

### 動態終端調整

當你調整本地終端視窗大小時，遠端終端會自動匹配新的尺寸。這透過 SSH 的 `window-change` 請求實現。

### Tab 補全

遠端 Shell 的 Tab 補全功能正常工作。

## 檔案傳輸（rz/sz）

在互動式 Shell 中，可以直接使用 `rz` 和 `sz` 命令進行檔案傳輸：

```bash
# 在遠端 Shell 中：
rz                              # 上傳檔案（開啟檔案選擇器）
rz /local/path/to/file          # 上傳指定本地檔案
sz /remote/path/to/file         # 下載檔案（開啟儲存對話框）
sz /remote/path/to/file /local  # 下載到指定本地路徑
```

!!! info "工作原理"
    當遠端 Shell 報告 `rz`/`sz: command not found` 時，win-sshpass 會攔截該錯誤，並透過 SFTP 執行傳輸。無需在遠端伺服器上安裝 lrzsz。

### 自訂檔案選擇器

預設情況下，rz/sz 使用系統檔案對話框（透過 zenity 實作）。如果需要自訂，可以透過 Go SDK 的 `WithFileSelector` 選項注入。

## 與標準 SSH 的區別

| 特性 | 標準 SSH | win-sshpass |
|------|----------|-------------|
| 密碼認證 | 需要 ssh-agent | 原生支援 |
| Windows 支援 | 需要安裝 OpenSSH | 獨立可執行檔 |
| rz/sz 傳輸 | 需要遠端安裝 lrzsz | 內建 SFTP 回退 |
| 進度條 | 無 | SFTP 傳輸時顯示 |

## 下一步

- [檔案傳輸](file-transfer.md) - SFTP 直接傳輸
- [Go SDK](../advanced/sdk.md) - 以程式設計方式使用 Shell
