# Web 控制台与 MCP

`win-sshpass -web` 会在浏览器中启动本地 SSH 工作区，支持命令行参数自动
连接、浏览器连接表单、SFTP 文件管理、真实终端和 MCP 工具。

## 启动

携带连接参数启动并自动登录：

```bash
win-sshpass -web -h 127.0.0.1 -u root -P 22 -p '密码'
```

打开 <http://127.0.0.1:6326>，页面会自动连接并进入工作区。

如果希望在浏览器中手动填写连接信息：

```bash
win-sshpass -web
```

修改监听地址：

```bash
win-sshpass -web -web-addr 127.0.0.1:6326
```

默认仅监听回环地址。绑定到其他网卡会向该网络暴露连接页面。

## 工作区

### SFTP 文件管理器

左侧是可交互的 SFTP 浏览器：

- 输入远端路径或进入上级目录
- 双击进入目录
- 上传、下载文件
- 新建目录
- 删除文件或空目录
- 将本机文件、目录拖入左侧上传
- 将远端文件拖出浏览器下载
- 在文件管理器底部查看文件名、序号、进度条和总体百分比

拖入目录时会递归创建远端目录。文件操作受 SSH 账号权限限制。

### SSH 终端

右侧是 xterm.js 渲染的真实 SSH PTY：

- 连续交互输入和 ANSI 输出
- Shell 历史、Tab 补全和全屏终端程序
- 自动同步终端行列尺寸
- 提供清屏和重新连接

它不是逐条命令 API，而是通过 WebSocket 在 xterm.js 与远端 Shell 之间传输
终端字节流。

### 可拖动布局

拖动中央分隔线可调整文件管理器和终端宽度。分隔线获得焦点后可使用左右方向键
调整，宽度会保存在浏览器中。

### 主题

主题选择器支持：

- **跟随系统** — 根据 `prefers-color-scheme` 自动选择，默认模式
- **浅色**
- **深色**

xterm 终端配色会与页面主题同步。

## 会话安全

浏览器获得随机 `HttpOnly`、`SameSite=Strict` Cookie。Cookie 不包含密码、
用户名或主机，SSH Client 和凭据只保留在本地服务进程内。会话在 30 分钟无操作
后过期。

HTTP API 要求 JSON 内容并拒绝跨源操作。

## MCP

Web 服务默认提供 Streamable HTTP MCP 端点：

```text
http://127.0.0.1:6326/mcp
```

MCP 会复用 `-web` 启动命令中的 SSH 连接参数。

### Codex 配置

```toml
[mcp_servers.win-sshpass]
url = "http://127.0.0.1:6326/mcp"
```

初始化响应会返回 `Mcp-Session-Id`，后续工具请求需要携带该请求头。可以使用
`DELETE /mcp` 结束会话。

### 工具

| 工具 | 功能 |
|---|---|
| `ssh_exec` | 执行 SSH 命令，返回 stdout、stderr 和退出码 |
| `sftp_list` | 列出远端目录 |
| `sftp_read` | 读取远端 UTF-8 文本文件 |
| `sftp_write` | 使用文本或 base64 创建/覆盖文件 |
| `sftp_mkdir` | 递归创建远端目录 |
| `sftp_delete` | 删除文件或空目录 |

调用示例：

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "ssh_exec",
    "arguments": {
      "command": "docker ps"
    }
  }
}
```

## 反向代理

如果在反向代理终止 TLS，需要转发 `/api/terminal` 的 WebSocket Upgrade，并
保留 `Mcp-Session-Id` 请求头。除非确实需要远程访问，否则保持默认回环绑定。

## 常见问题

- **页面停在 Connecting** — 先用普通 CLI 验证主机、端口、密码、私钥和代理。
- **终端断开** — 点击“重新连接终端”，并检查服务端 SSH/代理错误。
- **文件操作失败** — 检查 SSH 账号权限以及服务器是否启用 SFTP。
- **MCP tools/list 失败** — 先完成 initialize，并保留 `Mcp-Session-Id`。
