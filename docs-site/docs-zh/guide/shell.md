# 交互式 Shell

当不指定远程命令时，win-sshpass 会打开一个交互式 Shell，采用原始终端模式，提供接近原生 SSH 的体验。

## 基本使用

```bash
# 使用密码
win-sshpass -p 'password' ssh user@host

# 使用私钥
win-sshpass -i ~/.ssh/id_ed25519 ssh user@host

# 使用配置文件
win-sshpass -f server.config
```

## 原始终端模式特性

### 正确的回显

输入的字符会正确显示，不会出现双重回显问题。这得益于本地终端被设置为 raw 模式。

### 信号转发

- **Ctrl+C** — 中断当前远程进程
- **Ctrl+Z** — 挂起当前远程进程

### 全屏应用支持

以下应用可以正常使用：

- `vim` / `nvim` — 文本编辑器
- `top` / `htop` — 系统监控
- `nano` — 文本编辑器
- `less` / `more` — 分页查看
- ` Midnight Commander (mc)` — 文件管理器

### 动态终端调整

当你调整本地终端窗口大小时，远程终端会自动匹配新的尺寸。这通过 SSH 的 `window-change` 请求实现。

### Tab 补全

远程 Shell 的 Tab 补全功能正常工作。

## 文件传输（rz/sz）

在交互式 Shell 中，可以直接使用 `rz` 和 `sz` 命令进行文件传输：

```bash
# 在远程 Shell 中：
rz                              # 上传文件（打开文件选择器）
rz /local/path/to/file          # 上传指定本地文件
sz /remote/path/to/file         # 下载文件（打开保存对话框）
sz /remote/path/to/file /local  # 下载到指定本地路径
```

!!! info "工作原理"
    当远程 Shell 报告 `rz`/`sz: command not found` 时，win-sshpass 会拦截该错误，并通过 SFTP 执行传输。无需在远程服务器上安装 lrzsz。

### 自定义文件选择器

默认情况下，rz/sz 使用系统文件对话框（通过 zenity 实现）。如果需要自定义，可以通过 Go SDK 的 `WithFileSelector` 选项注入。

## 与标准 SSH 的区别

| 特性 | 标准 SSH | win-sshpass |
|------|----------|-------------|
| 密码认证 | 需要 ssh-agent | 原生支持 |
| Windows 支持 | 需要安装 OpenSSH | 独立可执行文件 |
| rz/sz 传输 | 需要远程安装 lrzsz | 内置 SFTP 回退 |
| 进度条 | 无 | SFTP 传输时显示 |

## 下一步

- [文件传输](file-transfer.md) - SFTP 直接传输
- [Go SDK](../advanced/sdk.md) - 以编程方式使用 Shell
