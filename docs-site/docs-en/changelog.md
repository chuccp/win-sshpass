# Changelog

## v0.3.2

- Add Linux support (amd64, arm64)
- Add proxy support: SOCKS5/SOCKS4/HTTP/HTTPS proxy tunneling
- Add breakpoint-resume for SFTP file transfers (`-resume` flag)
- Add hash and verify subcommands (md5, sha1, sha256, sha512)
- Add WinGet installation support
- Improve proxy timeout handling
- Initial documentation site

## v0.3.1

- Fix timeout reset during SFTP transfers
- Improve rz/sz file transfer stability

## v0.3.0

- Extract SDK package and CLI entry point
- Add `WithProgress`, `WithFileSelector`, `WithSignalHandler` options
- SDK no longer contains UI code; CLI-side adapters are separate

## v0.2.0

- Add SCP-style transfer support
- Add Rsync-style transfer support
- Add configuration file support
- Add connection retry with exponential backoff

## v0.1.0

- Initial release
- SSH password/key authentication
- Interactive shell (raw terminal mode)
- SFTP file transfer
- Dynamic terminal resizing
- rz/sz shell transfer
