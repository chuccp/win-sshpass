# 更新日誌

## v0.3.2

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
