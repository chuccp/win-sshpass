# 快速开始

本指南将帮助你在几分钟内上手 win-sshpass。

## 5 秒：执行远程命令

```bash
# 密码登录并执行命令
win-sshpass -p 'mypassword' ssh root@192.168.1.100 'whoami'
# → root
```

## 30 秒：文件传输

```bash
# 上传文件
win-sshpass -h 192.168.1.100 -p 'mypassword' -local ./app.jar -remote /opt/app/

# 下载文件
win-sshpass -h 192.168.1.100 -p 'mypassword' -d -remote /var/log/app.log -local ./logs/
```

## 1 分钟：交互式 Shell

```bash
# 打开交互式 Shell（不指定命令即可）
win-sshpass -p 'mypassword' ssh root@192.168.1.100
```

进入后你可以：

- 正常输入命令，回显正确
- 使用 `vim`、`top`、`htop` 等全屏应用
- 按 `Ctrl+C` 中断当前命令
- 终端窗口大小变化时，远程终端自动调整

## 3 分钟：使用配置文件

创建 `server.config`：

```yaml
host: 192.168.1.100
username: root
password: mypassword
port: 22
```

使用配置文件：

```bash
# 执行命令
win-sshpass -f server.config -c 'docker ps'

# 也可以省略 -c，直接将命令作为参数
win-sshpass -f server.config 'docker ps'

# 打开交互式 Shell
win-sshpass -f server.config
```

## 5 分钟：SCP/Rsync 风格传输

```bash
# SCP 上传文件
win-sshpass -p 'mypassword' scp ./app.jar user@server:/opt/app/

# SCP 上传目录
win-sshpass -p 'mypassword' scp -r ./dist user@server:/var/www/html

# Rsync 上传
win-sshpass -p 'mypassword' rsync -avz ./ user@server:/backup/
```

## 常见用法速查

```bash
# 私钥登录
win-sshpass -i ~/.ssh/id_ed25519 ssh user@server

# 环境变量传密码（更安全）
export SSHPASS='mypassword'
win-sshpass -e ssh user@server

# 指定端口
win-sshpass -p 'pass' ssh -p 2222 user@server

# 操作超时（30秒）
win-sshpass -p 'pass' -t 30 ssh user@server 'long-running-command'

# 上传多个文件
win-sshpass -h server -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/
```

## 下一步

- [SSH 连接](../guide/ssh.md) - 深入了解认证方式
- [文件传输](../guide/file-transfer.md) - SFTP 传输完整用法
- [配置文件](../guide/config-file.md) - 管理多台服务器
