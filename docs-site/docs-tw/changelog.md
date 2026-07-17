# 更新日誌

## v0.3.2

- 新增 Linux 支援（amd64、arm64）
- 新增代理支援：SOCKS5/SOCKS4/HTTP/HTTPS 代理通道
- 新增 SFTP 斷點續傳（`-resume` 參數）
- 新增 hash 和 verify 子命令（md5、sha1、sha256、sha512）
- 新增 WinGet 安裝支援
- 改進代理逾時處理
- 初始文件網站

## v0.3.1

- 修復 SFTP 傳輸中的逾時重置問題
- 改進 rz/sz 檔案傳輸的穩定性

## v0.3.0

- 提取 SDK 套件和 CLI 進入點
- 新增 `WithProgress`、`WithFileSelector`、`WithSignalHandler` 選項
- SDK 不再包含 UI 程式碼，CLI 側適配器獨立實作

## v0.2.0

- 新增 SCP 風格傳輸支援
- 新增 Rsync 風格傳輸支援
- 新增設定檔支援
- 新增連線重試和指數退避

## v0.1.0

- 初始版本
- SSH 密碼/金鑰認證
- 互動式 Shell（原始終端模式）
- SFTP 檔案傳輸
- 動態終端調整大小
- rz/sz Shell 內傳輸
