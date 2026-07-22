# 更新日誌

## v0.8.1

- 新增 SSH 金鑰對產生子命令（`keygen`）—— 支援 Ed25519 和 RSA，部署公鑰以實現免密碼登入
- 文件網站更新為 go-web-frame 風格，支援 Material 圖示、樣式化按鈕和更豐富的主題
- 新增金鑰產生 SDK 函式：`GenerateKeyPair`、`GenerateRSAKeyPair`、`SaveKeyPair`、`DeployPublicKey`、`DefaultKeyPath`
- 更新 4 種語言的 README 和文件網站，新增 keygen、代理和 hash/verify 章節
- 多項文件修復和改進

## v0.7.2

- 使用最新功能更新所有 README 和文件網站
- 在發布工作流程中啟用 WinGet 自動提交
- 新增 Linux 交叉編譯支援
- 新增代理支援：SOCKS5/SOCKS4/HTTP/HTTPS 代理通道（`-proxy` 參數）
- 新增 SFTP 檔案傳輸斷點續傳（`-resume` 參數）
- 新增 hash 和 verify 子命令（MD5、SHA-1、SHA-256、SHA-512）
- 修復代理逾時處理，附全面測試涵蓋

## v0.7.1

- 新增 MkDocs 文件網站，支援英語、簡體中文、繁體中文和日語翻譯
- 提取 SDK 套件（`package sshpass`）和 CLI 進入點（`cmd/sshpass`）
- 新增 Scoop 安裝說明
- 為所有 README 新增 Star 提醒

## v0.6.4

- 更新發布工作流程

## v0.6.3

- 更新發布工作流程和管線
- 新增互動式 Shell 中的 rz/sz 檔案傳輸支援（附 SFTP 回退）
- 使用最新功能同步所有 README

## v0.6.2

- 為所有 README 新增互動式 Shell 和設定檔位置命令範例
- 修復原始終端模式（正確的回顯、訊號轉發、全螢幕應用支援）
- 修復錯誤結束碼和設定檔命令處理
- 修復指數退避溢位
- 重構以消除程式碼重複並提高穩健性

## v0.6.1

- 新增 ARM64 架構建置支援
- 在文件中新增 ARM64 下載表
- 在設定中新增連接埠號驗證

## v0.5.1

- 新增可設定的連線逾時（`-ct`）和操作逾時（`-t`）
- 多項錯誤修復和改進
- 修復 Git Bash 路徑轉換範例（`//` 前綴）

## v0.4.x

- WiX v7 MSI 安裝程式支援
- macOS DMG/pkg 封裝
- 帶指數退避的連線重試
- 帶進度條的 SFTP 上傳/下載
- SCP 風格和 Rsync 風格命令語法
- 設定檔支援（`-f`）
- 帶原始終端模式的互動式 Shell
- 動態終端大小調整（SIGWINCH / 輪詢）

## v0.3.x

- macOS .pkg 安裝程式和 .tar.gz 封裝
- 管線改進

## v0.2.x

- SCP 和 Rsync 傳輸支援
- 設定檔支援
- 帶指數退避的連線重試
- 建置管線和發布自動化

## v0.1.0

- 初始版本
- SSH 密碼和私密金鑰認證
- 遠端命令執行
- 帶原始終端模式的互動式 Shell
- SFTP 檔案傳輸
- 動態終端大小調整
