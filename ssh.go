package sshpass

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// Dial creates an SSH client connection with automatic retry on transient
// failures. It honors the Config's authentication methods (private key and/or
// password), host key verification, connect timeout, and retry settings.
//
// Retry/backoff status messages are written to os.Stderr. NewClient routes
// these messages to the configured stderr stream (see WithStderr) instead.
func Dial(config *Config) (*ssh.Client, error) {
	return dial(config, os.Stderr)
}

// dial is the implementation of Dial that writes retry/backoff messages to the
// given writer, so NewClient can honor WithStderr for retry output.
func dial(config *Config, stderr io.Writer) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	// use private key authentication first
	if config.KeyPath != "" {
		key, err := os.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			hint := ""
			if strings.Contains(err.Error(), "passphrase") || strings.Contains(err.Error(), "encrypted") || strings.Contains(err.Error(), "password protected") {
				hint = " (key is passphrase-protected; this tool does not support encrypted private keys)"
			}
			return nil, fmt.Errorf("failed to parse private key: %w%s", err, hint)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// add password authentication if available (as fallback or primary)
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided (password or key required)")
	}

	// set host key verification callback
	var hostKeyCallback ssh.HostKeyCallback
	if config.StrictHostKey {
		// use known_hosts file for verification
		knownHostsPath := getKnownHostsPath()
		callback, err := knownhosts.New(knownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read known_hosts file (%s): %w\nhint: connect to the server manually first to add the host key", knownHostsPath, err)
		}
		hostKeyCallback = callback
	} else {
		// ignore host key verification (default)
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	attempts := max(1, config.Retries)

	// Strip IPv6 brackets if present — ParseUserHostPath and ParseSSHArgs store
	// IPv6 hosts as "[::1]" (with brackets), but net.JoinHostPort adds its own
	// brackets, so without stripping we'd get "[[::1]]:22" which is invalid.
	host := strings.TrimSuffix(strings.TrimPrefix(config.Host, "["), "]")
	address := net.JoinHostPort(host, config.Port)
	var lastErr error

	for i := 0; i < attempts; i++ {
		if i > 0 {
			// exponential backoff: 2s, 4s, 8s, 16s, then capped at 30s
			shift := i - 1
			if shift > 4 {
				shift = 4
			}
			delay := time.Duration(1<<shift) * 2 * time.Second
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			fmt.Fprintf(stderr, "Retrying connection (attempt %d/%d) in %v...\n", i+1, attempts, delay)
			time.Sleep(delay)
		}

		client, err := dialAndHandshake(address, sshConfig, config.ConnectTimeout, config.ProxyURL)
		if err == nil {
			return client, nil
		}

		lastErr = err

		// do not retry authentication failures
		if isAuthError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("connection failed after %d attempts: %w", attempts, lastErr)
}

// SSHClient is an alias for Dial, retained for compatibility with code that
// referenced the original function name.
func SSHClient(config *Config) (*ssh.Client, error) {
	return Dial(config)
}

// dialAndHandshake performs a single TCP dial + SSH handshake. When proxyURL
// is non-empty, the TCP connection is tunneled through the specified proxy
// (SOCKS4/5 or HTTP CONNECT) before the SSH handshake begins.
func dialAndHandshake(address string, sshConfig *ssh.ClientConfig, connectTimeout int, proxyURL string) (*ssh.Client, error) {
	var conn net.Conn
	var err error
	if proxyURL != "" {
		conn, err = proxyDial(proxyURL, address, connectTimeout)
	} else {
		var dialer net.Dialer
		if connectTimeout > 0 {
			dialer.Timeout = time.Duration(connectTimeout) * time.Second
		}
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// set handshake deadline using ConnectTimeout, then clear after handshake
	if connectTimeout > 0 {
		conn.SetDeadline(time.Now().Add(time.Duration(connectTimeout) * time.Second))
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ssh handshake failed: %w", err)
	}
	// clear deadline so the connection can live for the session duration
	conn.SetDeadline(time.Time{})
	return ssh.NewClient(sshConn, chans, reqs), nil
}

// isAuthError checks if an error is an unrecoverable authentication failure.
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unable to authenticate") ||
		strings.Contains(s, "no supported methods remain")
}

// getKnownHostsPath returns the known_hosts file path
func getKnownHostsPath() string {
	// use environment variable if set
	if path := os.Getenv("SSH_KNOWN_HOSTS"); path != "" {
		return path
	}

	// Windows default path: %USERPROFILE%\.ssh\known_hosts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}
	return joinLocalPath(homeDir, ".ssh", "known_hosts")
}

// getTerminalSize gets the current terminal size using x/term. fd is the file
// descriptor of the terminal to query (typically the stdin fd of an interactive
// shell session).
func getTerminalSize(fd int) (int, int) {
	w, h, err := term.GetSize(fd)
	if err != nil || w == 0 || h == 0 {
		return 80, 24
	}
	return w, h
}

// sendWindowChange sends a window-change request to the SSH session. fd is the
// terminal file descriptor used to read the current size.
func sendWindowChange(session *ssh.Session, fd int) {
	cols, rows := getTerminalSize(fd)
	session.SendRequest("window-change", false, ssh.Marshal(struct {
		Columns uint32
		Rows    uint32
		Width   uint32
		Height  uint32
	}{uint32(cols), uint32(rows), 0, 0}))
}

// runShell starts an interactive shell with dynamic terminal resizing using the
// Client's I/O streams. When stdin is a terminal, a PTY is requested and the
// local terminal is put into raw mode; an rz/sz monitor is attached so that
// remote rz/sz commands that are not installed on the server fall back to
// SFTP-based transfer.
func runShell(c *Client) error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// set standard I/O (stdin is set per-mode below)
	session.Stdout = c.stdout
	session.Stderr = c.stderr

	// only request PTY when stdin is a terminal (interactive use).
	// When stdin is a pipe, skip PTY to avoid echo issues and allow clean piping.
	var stdinFile *os.File
	if f, ok := c.stdin.(*os.File); ok {
		stdinFile = f
	}

	if stdinFile != nil && term.IsTerminal(int(stdinFile.Fd())) {
		stdinFd := int(stdinFile.Fd())
		// put local terminal into raw mode so keystrokes are sent directly
		// to the remote shell without local echo or line buffering
		oldState, err := term.MakeRaw(stdinFd)
		if err != nil {
			return fmt.Errorf("failed to set raw terminal: %w", err)
		}
		defer func() {
			term.Restore(stdinFd, oldState)
		}()

		cols, rows := getTerminalSize(stdinFd)

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}

		if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
			return fmt.Errorf("failed to request terminal: %w", err)
		}

		// watch for terminal resize events (platform-specific)
		done := make(chan struct{})
		defer close(done)
		watchTerminalResize(session, done, stdinFd)

		// create SFTP sub-channel for rz/sz file transfer support
		sftpClient, sftpErr := sftp.NewClient(c.sshClient)
		if sftpErr != nil {
			session.Stdin = stdinFile
			session.Stdout = c.stdout
		} else {
			defer sftpClient.Close()
			monitor := newRzszMonitor(stdinFile, sftpClient, oldState, c.selector, c.stdout, c.stderr, c.resetTimeout, c.progress)
			session.Stdin = stdinFile
			session.Stdout = &outputWriter{monitor: monitor, out: c.stdout}
		}
	} else {
		session.Stdin = c.stdin
	}

	// start remote shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return session.Wait()
}

// executeCommand executes a single command using the Client's I/O streams.
func executeCommand(c *Client, command string) error {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = c.stdin
	session.Stdout = c.stdout
	session.Stderr = c.stderr

	return session.Run(command)
}
