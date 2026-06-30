package sshpass

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client is a connected SSH client. It owns a single underlying *ssh.Client
// connection and exposes high-level operations (Exec, Shell, SFTP) configured
// through Options. A Client must be closed with Close when no longer needed.
type Client struct {
	config    *Config
	sshClient *ssh.Client

	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	progress ProgressFunc
	selector FileSelector
	signal   bool

	// operation-timeout machinery
	resetTimeout func()
	stopTimer    func()
	timedOut     atomic.Bool

	// interrupt-handler cleanup (nil unless WithSignalHandler is used)
	stopSignal func()

	closeOnce sync.Once
	closeErr  error
}

// NewClient establishes an SSH connection using config and returns a Client
// ready to execute commands, start a shell, or transfer files. Optional
// configuration is applied through opts (see WithStdin, WithStdout,
// WithProgress, WithSignalHandler, etc.).
//
// If config.Timeout > 0, an operation timer is armed that closes the
// underlying connection when the deadline elapses; subsequent Exec/Shell/SFTP
// calls will return an error and TimedOut will report true.
func NewClient(config *Config, opts ...Option) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config must not be nil")
	}

	c := &Client{
		config: config,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		// progress and selector default to nil: no progress reporting, and
		// rz/sz falls back to a stdin prompt. CLI/embedders inject UI via opts.
	}
	for _, opt := range opts {
		opt(c)
	}

	sshClient, err := dial(config, c.stderr)
	if err != nil {
		return nil, err
	}
	c.sshClient = sshClient

	// set up operation timeout (timer resets on each data transfer; closes the
	// connection when it fires).
	c.resetTimeout, c.stopTimer = setupOperationTimeout(c.stderr, func() {
		c.timedOut.Store(true)
		c.sshClient.Close()
	}, config.Timeout)

	if c.signal {
		// Register an interrupt handler that closes the connection so the main
		// goroutine unblocks. The returned stop function is invoked by Close to
		// unregister the handler and release the goroutine, preventing leaks in
		// long-running processes that create many clients.
		c.stopSignal = onInterrupt(func() { c.sshClient.Close() })
	}

	return c, nil
}

// Config returns the Config the client was created with.
func (c *Client) Config() *Config { return c.config }

// SSHClient returns the underlying *ssh.Client for advanced use. Callers must
// not close it; use Client.Close instead.
func (c *Client) SSHClient() *ssh.Client { return c.sshClient }

// Exec runs a single command on the remote host, streaming I/O through the
// client's configured stdin/stdout/stderr. It returns the command's error, if
// any.
func (c *Client) Exec(cmd string) error {
	return executeCommand(c, cmd)
}

// Shell starts an interactive remote shell with PTY and terminal-resize
// support. When stdin is a terminal, rz/sz commands not installed on the
// server fall back to SFTP-based transfer via the configured FileSelector.
func (c *Client) Shell() error {
	return runShell(c)
}

// SFTP opens an SFTP sub-channel over the client's SSH connection and returns
// an *SFTPClient for file uploads/downloads. The returned SFTPClient must be
// closed when done; closing it does not close the underlying SSH connection.
func (c *Client) SFTP() (*SFTPClient, error) {
	sftpClient, err := sftp.NewClient(c.sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	return &SFTPClient{
		sftpClient:   sftpClient,
		resetTimeout: c.resetTimeout,
		progress:     c.progress,
	}, nil
}

// Close stops the operation timer and interrupt handler (if any) and closes
// the underlying SSH connection. It is idempotent and safe to call multiple
// times; subsequent calls return the same error as the first.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		if c.stopSignal != nil {
			c.stopSignal()
		}
		if c.stopTimer != nil {
			c.stopTimer()
		}
		if c.sshClient != nil {
			c.closeErr = c.sshClient.Close()
		}
	})
	return c.closeErr
}

// TimedOut reports whether the operation timeout has fired. When true, the
// most recent Exec/Shell/SFTP error is due to the deadline elapsing.
func (c *Client) TimedOut() bool {
	return c.timedOut.Load()
}
