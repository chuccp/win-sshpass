package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// SSHClient creates an SSH client connection
func SSHClient(config *Config) (*ssh.Client, error) {
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

	address := net.JoinHostPort(config.Host, config.Port)
	var dialer net.Dialer
	if config.ConnectTimeout > 0 {
		dialer.Timeout = time.Duration(config.ConnectTimeout) * time.Second
	}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	// set handshake deadline using ConnectTimeout, then clear after handshake
	if config.ConnectTimeout > 0 {
		conn.SetDeadline(time.Now().Add(time.Duration(config.ConnectTimeout) * time.Second))
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

// getTerminalSize gets the current terminal size using x/term
func getTerminalSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w == 0 || h == 0 {
		return 80, 24
	}
	return w, h
}

// sendWindowChange sends a window-change request to the SSH session
func sendWindowChange(session *ssh.Session) {
	cols, rows := getTerminalSize()
	session.SendRequest("window-change", false, ssh.Marshal(struct {
		Columns uint32
		Rows    uint32
		Width   uint32
		Height  uint32
	}{uint32(cols), uint32(rows), 0, 0}))
}

// runShell starts an interactive shell with dynamic terminal resizing
func runShell(client *ssh.Client) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// set standard I/O
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// only request PTY when stdin is a terminal (interactive use).
	// When stdin is a pipe, skip PTY to avoid echo issues and allow clean piping.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		cols, rows := getTerminalSize()

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
		watchTerminalResize(session, done)
	}

	// start remote shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return session.Wait()
}

// executeCommand executes a single command
func executeCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	return session.Run(command)
}
