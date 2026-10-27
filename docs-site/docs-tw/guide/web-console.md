# Web 控制台與 MCP

`win-sshpass -web` 會在瀏覽器中啟動本機 SSH 工作區，支援命令列自動連線、
瀏覽器連線表單、SFTP 檔案管理、真實終端與 MCP 工具。

## 啟動

攜帶連線參數並自動登入：

```bash
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p '密碼'
```

開啟 <http://127.0.0.1:6326> 後會自動連線並進入工作區。

若要在瀏覽器手動填寫連線資訊：

```bash
win-sshpass -web
```

變更監聽位址：

```bash
win-sshpass -web -web-addr 127.0.0.1:6326
```

預設僅監聽回環位址。綁定到其他網路介面會向該網路公開連線頁面。

## 工作區

### SFTP 檔案管理員

左側為互動式 SFTP 瀏覽器：

- 輸入遠端路徑或前往上層目錄
- 雙擊進入目錄
- 上傳、下載檔案
- 建立目錄
- 刪除檔案或空目錄
- 將本機檔案、目錄拖入左側上傳
- 將遠端檔案拖出瀏覽器下載
- 在底部查看檔名、序號、進度條與總體百分比

拖入目錄時會遞迴建立遠端資料夾。檔案操作受 SSH 帳號權限限制。

### SSH 終端

右側為 xterm.js 呈現的真實 SSH PTY：

- 連續互動輸入與 ANSI 輸出
- Shell 歷史、Tab 補全與全螢幕終端程式
- 自動同步終端尺寸
- 清屏與重新連線

透過 WebSocket 在 xterm.js 與遠端 Shell 間傳輸終端位元組，不是逐條命令 API。

### 可拖曳版面

拖曳中央分隔線可調整檔案管理員與終端寬度。分隔線取得焦點後可使用左右方向鍵
調整，寬度會保存在瀏覽器。

### 主題

- **跟隨系統** — 依 `prefers-color-scheme`，預設模式
- **淺色**
- **深色**

xterm 配色會同步切換。

## 工作階段安全

瀏覽器取得隨機 `HttpOnly`、`SameSite=Strict` Cookie。Cookie 不包含密碼、
使用者名稱或主機，SSH 憑證僅保留於本機服務程序。工作階段閒置 30 分鐘後過期。

HTTP API 要求 JSON，並拒絕跨來源操作。

## MCP

Web 服務預設提供 Streamable HTTP MCP 端點：

```text
http://127.0.0.1:6326/mcp
```

MCP 會沿用 `-web` 啟動命令中的 SSH 連線參數。

### Codex 設定

```toml
[mcp_servers.win-sshpass]
url = "http://127.0.0.1:6326/mcp"
```

初始化回應會提供 `Mcp-Session-Id`，後續工具請求需攜帶此標頭。可使用
`DELETE /mcp` 結束工作階段。

### 工具

| 工具 | 功能 |
|---|---|
| `ssh_exec` | 執行 SSH 命令，回傳 stdout、stderr 與結束碼 |
| `sftp_list` | 列出遠端目錄 |
| `sftp_read` | 讀取遠端 UTF-8 文字檔 |
| `sftp_write` | 以文字或 base64 建立/覆寫檔案 |
| `sftp_mkdir` | 遞迴建立遠端目錄 |
| `sftp_delete` | 刪除檔案或空目錄 |

## 反向代理

若 TLS 由反向代理終止，需轉送 `/api/terminal` 的 WebSocket Upgrade，並保留
`Mcp-Session-Id` 標頭。除非確實需要遠端存取，否則保持預設回環綁定。

## 常見問題

- **頁面停在 Connecting** — 先使用一般 CLI 驗證主機、連接埠、密碼、金鑰與代理。
- **終端中斷** — 使用「重新連線終端」，並檢查服務端 SSH/代理錯誤。
- **檔案操作失敗** — 檢查 SSH 帳號權限及伺服器 SFTP 設定。
- **MCP tools/list 失敗** — 先完成 initialize，並保留 `Mcp-Session-Id`。
