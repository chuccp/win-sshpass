# File Transfer

win-sshpass provides multiple file transfer methods: direct SFTP, SCP style, Rsync style, and rz/sz in interactive shell.

## Direct SFTP Transfer

### Upload Files

```bash
# Upload single file
win-sshpass -h host -p 'pass' -local ./file.txt -remote /tmp/file.txt

# Upload multiple files (comma-separated)
win-sshpass -h host -p 'pass' -local "a.txt,b.txt,c.txt" -remote /tmp/

# Upload multiple files (space-separated, for simple paths only)
win-sshpass -h host -p 'pass' -local "a.txt b.txt c.txt" -remote /tmp/

# Upload directory (auto-recursive)
win-sshpass -h host -p 'pass' -local ./dist -remote /var/www/html
```

### Download Files

```bash
# Download file
win-sshpass -h host -p 'pass' -d -remote /tmp/file.txt -local ./file.txt

# Download directory
win-sshpass -h host -p 'pass' -d -remote /var/log/nginx -local ./logs
```

### Progress Bar

SFTP transfers automatically display a progress bar:

```
Uploading app.jar  45% |████████████         |  45MB/100MB  10MB/s
```

## SCP Style Transfer

Compatible with standard scp command syntax:

```bash
# Upload file
win-sshpass -p 'pass' scp ./file.txt user@host:/tmp/

# Upload directory (-r for recursive)
win-sshpass -p 'pass' scp -r ./dist user@host:/var/www/html

# Specify port (-P, note: uppercase)
win-sshpass -p 'pass' scp -P 2222 ./file.txt user@host:/tmp/

# Download file
win-sshpass -p 'pass' scp user@host:/tmp/file.txt ./

# Download directory
win-sshpass -p 'pass' scp -r user@host:/var/log/nginx ./logs
```

## Rsync Style Transfer

Compatible with rsync command syntax:

```bash
# Upload
win-sshpass -p 'pass' rsync -avz ./ user@host:/backup/

# Download
win-sshpass -p 'pass' rsync -avz user@host:/data/ ./local-data/

# Specify port
win-sshpass -p 'pass' rsync --port=2222 -avz ./ user@host:/backup/
```

## Shell rz/sz Transfer

In interactive shell, you can use `rz` and `sz` commands for file transfer — **no software needs to be installed on the remote server**.

```bash
# First open interactive shell
win-sshpass -p 'pass' ssh user@host

# In remote shell:
rz                              # Upload file (opens file picker)
rz /local/path/to/file          # Upload specific local file
sz /remote/path/to/file         # Download file (opens save dialog)
sz /remote/path/to/file /local  # Download to specific local path
```

### How It Works

When the remote shell reports `rz`/`sz: command not found`, win-sshpass intercepts the error and performs the transfer over SFTP instead. Both files and directories are supported, with progress bars.

!!! info "No Remote Installation Needed"
    rz/sz transfer is implemented via SFTP — no need to install the lrzsz package on the remote server.

## Git Bash Path Note

When using Git Bash, remote paths starting with `/` are automatically converted to Windows paths. Use the `//` prefix to avoid this:

```bash
# Wrong: /tmp will be converted to Windows path
win-sshpass ... -remote /tmp/file.txt

# Correct: use double slashes
win-sshpass ... -remote //tmp/file.txt
```

## Next Steps

- [SCP & Rsync](scp-rsync.md) - More SCP/Rsync usage
- [Interactive Shell](shell.md) - Full shell mode features
