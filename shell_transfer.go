package sshpass

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/term"
)

// rzszMonitor monitors SSH session OUTPUT for rz/sz commands that the remote
// shell can't find. When "command not found" is detected and the command was
// rz or sz, the handler runs locally via SFTP.
//
// This approach requires NO input interception at all — input goes directly
// to the remote shell. This means full echo, tab completion, command history,
// copy-paste, and all other terminal features work normally. rz/sz is detected
// purely from the remote's output, so it works regardless of how the command
// was entered.
type rzszMonitor struct {
	sftpClient   *sftp.Client
	stdin        *os.File
	oldState     *term.State
	selector     FileSelector
	stdout       io.Writer
	stderr       io.Writer
	logger       *slog.Logger
	resetTimeout func()
	progress     ProgressFunc

	mu      sync.Mutex
	recent  []byte // rolling buffer of recent output
	Handled bool   // set when a command is being handled (prevents re-trigger)
	// test hooks
	onRZ func(localPath string)
	onSZ func(remotePath, localPath string)
}

func newRzszMonitor(stdin *os.File, sftpClient *sftp.Client, oldState *term.State, selector FileSelector, stdout, stderr io.Writer, logger *slog.Logger, resetTimeout func(), progress ProgressFunc) *rzszMonitor {
	return &rzszMonitor{
		stdin:        stdin,
		sftpClient:   sftpClient,
		oldState:     oldState,
		selector:     selector,
		stdout:       stdout,
		stderr:       stderr,
		logger:       logger,
		resetTimeout: resetTimeout,
		progress:     progress,
	}
}

// outputWriter wraps the session stdout writer, passing all output through
// while scanning for rz/sz "command not found" patterns.
type outputWriter struct {
	monitor *rzszMonitor
	out     io.Writer
}

func (w *outputWriter) Write(p []byte) (int, error) {
	// Always pass output through immediately
	n, err := w.out.Write(p)

	w.monitor.mu.Lock()

	// Skip if already handling a command
	if w.monitor.Handled {
		w.monitor.mu.Unlock()
		return n, err
	}

	// Update rolling buffer
	w.monitor.recent = append(w.monitor.recent, p...)
	if len(w.monitor.recent) > 4096 {
		w.monitor.recent = w.monitor.recent[len(w.monitor.recent)-4096:]
	}

	// Check if the new output contains "not found"
	if containsNotFound(p) {
		// Look for rz/sz command in the rolling buffer, near the "not found" line
		cmd, args := extractCommandFromNotFound(w.monitor.recent)
		if cmd != "" {
			w.monitor.Handled = true
			w.monitor.recent = w.monitor.recent[:0]
			w.monitor.mu.Unlock()

			// Run handler (blocks this goroutine — fine, remote is idle)
			if cmd == "rz" {
				path := ""
				if len(args) > 0 {
					path = args[0]
				}
				if w.monitor.onRZ != nil {
					w.monitor.onRZ(path)
				} else {
					w.monitor.handleRZ(path)
				}
			} else if cmd == "sz" {
				remotePath := ""
				localPath := ""
				if len(args) > 0 {
					remotePath = args[0]
				}
				if len(args) > 1 {
					localPath = args[1]
				}
				if w.monitor.onSZ != nil {
					w.monitor.onSZ(remotePath, localPath)
				} else {
					w.monitor.handleSZ(remotePath, localPath)
				}
			}

			w.monitor.mu.Lock()
			w.monitor.Handled = false
			w.monitor.mu.Unlock()
			return n, err
		}
	}

	w.monitor.mu.Unlock()
	return n, err
}

// containsNotFound checks if the output contains a "command not found" pattern.
func containsNotFound(p []byte) bool {
	return bytes.Contains(p, []byte("not found")) ||
		bytes.Contains(p, []byte("未找到命令")) ||
		bytes.Contains(p, []byte("No such file or directory"))
}

// extractCommandFromNotFound finds the "not found" error line in the buffer,
// extracts the command name from it, then looks at the preceding line (the
// echoed command) for arguments. This avoids false positives from old buffer
// content.
func extractCommandFromNotFound(buf []byte) (string, []string) {
	s := string(buf)

	// Find "not found" (case-insensitive)
	lowerS := strings.ToLower(s)
	nfIdx := strings.Index(lowerS, "not found")
	if nfIdx < 0 {
		return "", nil
	}

	// Find the start of the line containing "not found"
	lineStart := strings.LastIndex(s[:nfIdx], "\n")
	if lineStart < 0 {
		lineStart = 0
	} else {
		lineStart++
	}

	// Extract the "not found" line
	lineEnd := strings.Index(s[nfIdx:], "\n")
	if lineEnd < 0 {
		lineEnd = len(s) - nfIdx
	}
	nfLine := s[lineStart : nfIdx+lineEnd]

	// Extract command name from the error line
	// bash: "-bash: rz: command not found"
	// zsh:  "zsh: command not found: rz"
	// sh:   "sh: rz: not found"
	cmd := ""
	for _, candidate := range []string{"rz", "sz"} {
		for _, f := range strings.Fields(nfLine) {
			f = strings.TrimSuffix(f, ":")
			if f == candidate {
				cmd = candidate
				break
			}
		}
		if cmd != "" {
			break
		}
	}
	if cmd == "" {
		return "", nil
	}

	// Find the echoed command line (the line before "not found")
	prevLineEnd := lineStart - 1
	if prevLineEnd <= 0 {
		return cmd, nil
	}
	prevLineStart := strings.LastIndex(s[:prevLineEnd], "\n")
	if prevLineStart < 0 {
		prevLineStart = 0
	} else {
		prevLineStart++
	}
	echoedLine := strings.TrimRight(s[prevLineStart:prevLineEnd], "\r")

	// Parse args from the echoed line
	fields := strings.Fields(echoedLine)
	for i, f := range fields {
		if f == cmd {
			return cmd, fields[i+1:]
		}
	}

	return cmd, nil
}

// --- Handlers ---

func (m *rzszMonitor) handleRZ(localPath string) {
	m.restoreTerminal()
	defer m.enterRawMode()

	if localPath == "" && m.selector != nil {
		path, err := m.selector.OpenFile()
		if err != nil {
			m.logger.Error("file dialog error", "op", "open", "err", err)
		} else {
			localPath = path
		}
	}
	if localPath == "" {
		fmt.Fprint(m.stdout, "Local file path to upload: ")
		localPath = readLineFromStdin(m.stdin)
	}

	if localPath == "" {
		fmt.Fprintln(m.stdout, "Upload cancelled")
		return
	}

	remoteCwd, err := m.sftpClient.Getwd()
	if err != nil {
		remoteCwd = "."
	}

	fmt.Fprintf(m.stdout, "Uploading %s -> %s...\n", localPath, remoteCwd)
	if err := uploadFile(m.sftpClient, localPath, remoteCwd, m.resetTimeout, m.progress, false); err != nil {
		fmt.Fprintf(m.stderr, "Upload failed: %v\n", err)
	} else {
		fmt.Fprintln(m.stdout, "Upload complete")
	}
}

func (m *rzszMonitor) handleSZ(remotePath, localPath string) {
	if strings.HasPrefix(remotePath, "//") {
		remotePath = remotePath[1:]
	}

	m.restoreTerminal()
	defer m.enterRawMode()

	if localPath == "" && m.selector != nil {
		defaultName := remoteBaseName(remotePath)
		path, err := m.selector.SaveFile(defaultName)
		if err != nil {
			m.logger.Error("file dialog error", "op", "save", "err", err)
			localPath = defaultName
		} else if path != "" {
			localPath = path
		} else {
			fmt.Fprintln(m.stdout, "Download cancelled")
			return
		}
	}
	if localPath == "" {
		localPath = remoteBaseName(remotePath)
	}

	fmt.Fprintf(m.stdout, "Downloading %s -> %s...\n", remotePath, localPath)
	if err := downloadFile(m.sftpClient, remotePath, localPath, m.resetTimeout, m.progress, false); err != nil {
		fmt.Fprintf(m.stderr, "Download failed: %v\n", err)
	} else {
		fmt.Fprintln(m.stdout, "Download complete")
	}
}

func (m *rzszMonitor) restoreTerminal() {
	if m.oldState != nil {
		term.Restore(int(m.stdin.Fd()), m.oldState)
	}
}

func (m *rzszMonitor) enterRawMode() {
	state, err := term.MakeRaw(int(m.stdin.Fd()))
	if err == nil {
		m.oldState = state
	}
}

func readLineFromStdin(r io.Reader) string {
	var buf []byte
	for {
		var b [1]byte
		n, err := r.Read(b[:])
		if err != nil || n == 0 {
			break
		}
		if b[0] == '\r' || b[0] == '\n' {
			break
		}
		buf = append(buf, b[0])
	}
	return strings.TrimSpace(string(buf))
}
