# Changelog

## v0.8.1

- Add SSH key pair generation subcommand (`keygen`) — Ed25519 and RSA, deploy public keys to enable password-less login
- Add Docker-based integration test suite (71 tests covering all features)
- Update documentation site to go-web-frame style with Material icons, styled buttons, and richer theme
- Add key generation SDK functions: `GenerateKeyPair`, `GenerateRSAKeyPair`, `SaveKeyPair`, `DeployPublicKey`, `DefaultKeyPath`
- Update all 4-language READMEs and docs-site with keygen, Docker testing, proxy, and hash/verify sections
- Various documentation fixes and improvements

## v0.7.2

- Update all READMEs and docs-site with recent features
- Enable WinGet auto-submit in release workflow
- Add Linux cross-compilation support
- Add proxy support: SOCKS5/SOCKS4/HTTP/HTTPS proxy tunneling (`-proxy` flag)
- Add breakpoint-resume for SFTP file transfers (`-resume` flag)
- Add hash and verify subcommands (MD5, SHA-1, SHA-256, SHA-512)
- Fix proxy timeout handling with comprehensive test coverage

## v0.7.1

- Add MkDocs documentation site with English, 简体中文, 繁體中文, and 日本語 translations
- Extract SDK package (`package sshpass`) and CLI entry point (`cmd/sshpass`)
- Add Scoop installation instructions
- Add Star reminder to all READMEs

## v0.6.4

- Update release workflow

## v0.6.3

- Update release workflow and pipeline
- Add rz/sz file transfer support in interactive shell (with SFTP fallback)
- Sync all READMEs with latest features

## v0.6.2

- Add interactive shell and config file positional command examples to all READMEs
- Fix raw terminal mode (proper echo, signal forwarding, full-screen app support)
- Fix error exit codes and config file command handling
- Fix exponential backoff overflow
- Refactor to eliminate code duplication and improve robustness

## v0.6.1

- Add ARM64 architecture build support
- Add ARM64 download tables to documentation
- Add port number validation in config

## v0.5.1

- Add configurable connection timeout (`-ct`) and operation timeout (`-t`)
- Multiple bug fixes and improvements
- Fix Git Bash path conversion in examples (`//` prefix)

## v0.4.x

- WiX v7 MSI installer support
- macOS DMG/pkg packaging
- Connection retry with exponential backoff
- SFTP upload/download with progress bars
- SCP-style and Rsync-style command syntax
- Configuration file support (`-f`)
- Interactive shell with raw terminal mode
- Dynamic terminal resizing (SIGWINCH / polling)

## v0.3.x

- macOS .pkg installer and .tar.gz packaging
- Pipeline improvements

## v0.2.x

- SCP and Rsync transfer support
- Configuration file support
- Connection retry with exponential backoff
- Build pipeline and release automation

## v0.1.0

- Initial release
- SSH password and private key authentication
- Remote command execution
- Interactive shell with raw terminal mode
- SFTP file transfer
- Dynamic terminal resizing
