# win-sshpass 使用说明

---

我开发机是 Windows，服务器是 Linux。平时要 SSH 上去看日志、传文件、重启服务，输密码麻烦，写到脚本里又会卡在密码提示。

win-sshpass 可以绕过密码提示。命令里带 -p 参数直接传密码：

```bash
win-sshpass -p 'mima123' ssh root@192.168.1.100 'uptime'
```

> **[截图位置：一条命令从输入到输出的终端截图]**

下载地址：https://github.com/chuccp/win-sshpass



---

密码除了 -p 直接传，还可以从文件读、从环境变量读、或者用私钥：

```bash
# 密码文件
echo 'mima123' > pass.txt
win-sshpass -f pass.txt ssh root@server 'df -h'

# 环境变量
export SSHPASS='mima123'
win-sshpass -e ssh root@server 'df -h'

# 私钥
win-sshpass -i ~/.ssh/id_ed25519 ssh ubuntu@server 'hostname'
```

> **[截图位置：四种方式并排展示]**

---

多台服务器可以写配置文件，每台一个：

```
host: 192.168.1.10
username: root
password: qwe123
port: 22
timeout: 30
connect_timeout: 5
retry: 3
```

然后用 -f 指定配置文件：

```bash
win-sshpass -f web01.config -c 'systemctl status nginx'
win-sshpass -f db01.config  -c 'df -h /data'
```

> **[截图位置：目录里几个 config 文件]**

---

传文件有三种方式。SFTP 直连、SCP 风格、Rsync 风格。

```bash
# SFTP 直连，上传
win-sshpass -h server -p 'mima' -local app.jar -remote //opt/app/

# 上传目录
win-sshpass -h server -p 'mima' -local ./dist -remote //var/www/html

# 下载，加 -d
win-sshpass -h server -p 'mima' -d -remote //var/log/nginx.log -local ./

# 多个文件用逗号隔开
win-sshpass -h server -p 'mima' -local "a.txt,b.txt,c.txt" -remote //tmp/

# SCP
win-sshpass -p 'mima' scp ./app.jar root@server:/opt/
win-sshpass -p 'mima' scp -r ./dist root@server:/var/www/

# Rsync
win-sshpass -p 'mima' rsync -avz ./dist/ root@server:/var/www/
```

传输有进度条。

> **[截图位置：进度条截屏]**

---

不指定命令就是交互 Shell：

```bash
win-sshpass -p 'mima' ssh root@server
```

进去后跟普通 ssh 一样。vim、top、htop 都能正常用，Ctrl+C 可以终止远程进程。窗口大小变化终端也跟着变。

> **[截图位置：交互 Shell 里跑 vim 或 htop]**

交互 Shell 里输入 rz 会弹出文件选择框，选文件后通过 SFTP 上传到远程当前目录。sz 则是弹出保存对话框，下载远程文件。

```bash
[root@server ~]# rz                          # 弹窗选文件，上传
[root@server ~]# sz /etc/nginx/nginx.conf    # 弹窗选位置，下载
```

服务器端不需要装任何东西。原理是检测到 "command not found" 时自动走 SFTP 通道。

> **[截图位置：rz 弹窗 → 进度条 → 回到 Shell]**

---

AI 编程助手可以通过 win-sshpass 连接服务器。命令行加 -t 设超时、-retry 设重试次数，适合自动化场景：

```bash
win-sshpass -e -h server -t 30 -retry 3 -c 'systemctl status app'
```

> **[截图位置：AI 助手界面，用户指令 → 调用 win-sshpass → 完成操作]**

---

写个 bat 做日常巡检：

```batch
set SSHPASS=MyPass

for %%H in (web-01 web-02 db-01 cache-01) do (
    echo === %%H ===
    win-sshpass -e -h %%H -u root -c "uptime; free -h | grep Mem; df -h / | tail -1"
)
```

放 Windows 任务计划里定时跑。

> **[截图位置：批量脚本输出]**

---

参数列表：

| 参数 | 说明 |
|------|------|
| `-p` | 密码 |
| `-i` | 私钥路径 |
| `-f` | 密码文件或配置文件 |
| `-e` | 从环境变量 SSHPASS 读密码 |
| `-h` | 主机地址 |
| `-u` | 用户名，默认 root |
| `-P` | 端口，默认 22 |
| `-c` | 执行的命令 |
| `-local` | 本地路径 |
| `-remote` | 远程路径 |
| `-d` | 下载模式 |
| `-k` | 启用严格主机密钥验证 |
| `-t` | 操作超时（秒），0 不限 |
| `-ct` | 连接超时（秒），默认 10 |
| `-retry` | 连接重试次数，默认 3 |
| `-v` | 显示版本 |
| `-help` | 显示帮助 |

> **[截图位置：终端运行 win-sshpass -v]**

---

*GitHub: [chuccp/win-sshpass](https://github.com/chuccp/win-sshpass)*
