# SCP & Rsync

win-sshpass is compatible with standard scp and rsync command syntax, with file transfer implemented via SFTP under the hood.

## SCP Style Transfer

### Basic Syntax

```bash
win-sshpass -p <password> scp [options] <source> <target>
```

### Upload Files

```bash
# Upload single file
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# Upload with specific filename
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/newname.txt

# Upload directory (-r for recursive)
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html
```

### Download Files

```bash
# Download file
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# Download directory
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

### Specify Port

scp uses uppercase `-P` for port (unlike ssh's lowercase `-p`):

```bash
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/
```

### Supported Options

| Option | Description |
|--------|-------------|
| `-r` | Recursive directory copy |
| `-P <port>` | Specify port |
| `-i <key>` | Specify private key |
| `-q` | Quiet mode |
| `-C` | Compression (handled by SFTP) |
| `-v` | Verbose output |

## Rsync Style Transfer

### Basic Syntax

```bash
win-sshpass -p <password> rsync [options] <source> <target>
```

### Upload

```bash
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/
```

### Download

```bash
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/
```

### Specify Port

rsync uses `--port=` for port:

```bash
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

### Supported Options

| Option | Description |
|--------|-------------|
| `-a` | Archive mode |
| `-v` | Verbose output |
| `-z` | Compress during transfer |
| `--port=N` | Specify port |
| `-e ssh` | Specify remote shell (ignored) |

## SCP vs Rsync vs SFTP

| Method | Syntax | Best For |
|--------|--------|----------|
| SCP | Standard scp syntax | Simple file copy |
| Rsync | Standard rsync syntax | Incremental sync (note: current impl is full transfer) |
| SFTP | `-local` / `-remote` flags | Flexible file transfer, multi-file support |

!!! note "Note"
    win-sshpass's rsync implementation uses SFTP under the hood and does not support rsync's incremental sync algorithm. For true incremental sync, install rsync on the remote server and use it directly over SSH.

## Next Steps

- [File Transfer](file-transfer.md) - More SFTP direct transfer usage
- [Configuration File](config-file.md) - Simplify commands with config files
