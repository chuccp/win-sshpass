package sshpass

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// --- String validation ---

// isAllDigits checks if a string consists only of digit characters
func isAllDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// isValidPort checks if a string is a valid network port number (1-65535)
func isValidPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 65535
}

// --- String manipulation ---

// JoinArgs joins a string slice with space
func JoinArgs(args []string) string {
	return strings.Join(args, " ")
}

// SplitPaths splits a path string by comma or space separator.
// Returns error if complex paths (containing '/' or '\') are space-separated.
// name identifies which parameter for error messages (e.g., "local" or "remote").
func SplitPaths(s, name string) ([]string, error) {
	var paths []string
	if strings.Contains(s, ",") {
		for _, p := range strings.Split(s, ",") {
			if p = strings.TrimSpace(p); p != "" {
				paths = append(paths, p)
			}
		}
	} else if strings.Contains(s, " ") {
		for _, p := range strings.Fields(s) {
			if strings.ContainsAny(p, "/\\") {
				return nil, fmt.Errorf("path %q contains a path separator. Please use commas to separate multiple %s paths (e.g., -%s \"./a/file.txt,./b/file.txt\")", p, name, name)
			}
		}
		paths = strings.Fields(s)
	} else {
		paths = []string{s}
	}
	return paths, nil
}

// --- Path helpers ---

// isWindowsLocalPath checks if a path looks like a Windows local path
// e.g. "C:/Users/..." or "C:\\Users\\..."
func isWindowsLocalPath(p string) bool {
	if len(p) < 3 {
		return false
	}
	// match C:/ or C:\  pattern
	return (p[1] == ':' && (p[2] == '/' || p[2] == '\\')) &&
		((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z'))
}

// CleanRemotePath normalizes a remote path for SFTP use.
//   - Strips the leading "/" from "//" prefix used to bypass Git Bash path conversion
//     (e.g. "//tmp/file" -> "/tmp/file")
//   - Returns an error when the path looks like a Windows local path caused by Git
//     Bash path conversion, instead of terminating the process.
func CleanRemotePath(p string) (string, error) {
	// "//" prefix: user intentionally used it to bypass Git Bash conversion
	// e.g. "//tmp/file" should become "/tmp/file"
	if strings.HasPrefix(p, "//") {
		return p[1:], nil
	}
	// detect Git Bash path conversion: /tmp/file became C:/Users/.../tmp/file
	if isWindowsLocalPath(p) {
		return "", fmt.Errorf("remote path %q looks like a Windows local path (Git Bash path conversion); use '//' prefix to avoid conversion, e.g. //tmp/file instead of /tmp/file", p)
	}
	return p, nil
}

// joinRemotePath joins remote path elements using Unix-style / separator
func joinRemotePath(elems ...string) string {
	return path.Join(elems...)
}

// joinLocalPath joins local path elements using the OS-specific separator
func joinLocalPath(elems ...string) string {
	return filepath.Join(elems...)
}

// remoteBaseName returns the last element of a remote (Unix-style) path
func remoteBaseName(p string) string {
	return path.Base(p)
}

// localBaseName returns the last element of a local path
func localBaseName(p string) string {
	return filepath.Base(p)
}

// remoteDirName returns the directory portion of a remote (Unix-style) path
func remoteDirName(p string) string {
	return path.Dir(p)
}

// localDirName returns the directory portion of a local path
func localDirName(p string) string {
	return filepath.Dir(p)
}

// toSlash converts Windows backslash paths to forward slashes
func toSlash(p string) string {
	return filepath.ToSlash(p)
}

// --- String parsing ---

// ParseUserHostPath parses user@host:path format, supporting IPv6.
// Returns user, host, path.
func ParseUserHostPath(arg string) (user, host, remotePath string) {
	atIdx := strings.Index(arg, "@")
	if atIdx <= 0 {
		return "", "", ""
	}
	user = arg[:atIdx]
	remainder := arg[atIdx+1:]

	// check if IPv6 address (starts with [)
	if strings.HasPrefix(remainder, "[") {
		// IPv6 format: [::1]:path or [2001:db8::1]:path
		closeBracket := strings.Index(remainder, "]")
		if closeBracket > 0 {
			host = remainder[:closeBracket+1] // including square brackets
			// check if there is a path after ]:
			if closeBracket+1 < len(remainder) && remainder[closeBracket+1] == ':' {
				remotePath = remainder[closeBracket+2:]
			}
		}
	} else {
		// IPv4 or hostname: host:path
		colonIdx := strings.Index(remainder, ":")
		if colonIdx > 0 {
			host = remainder[:colonIdx]
			remotePath = remainder[colonIdx+1:]
		} else {
			host = remainder
		}
	}
	return user, host, remotePath
}

// --- Timeout helpers ---

// setupOperationTimeout creates a timer that calls closeFn after the given
// timeout. out receives the "Operation timed out" notice. It returns a reset
// function to extend the deadline and a stop function to cancel the timer.
// When timeout <= 0, no timer is created and the returned stop is a no-op.
func setupOperationTimeout(out io.Writer, closeFn func(), timeout int) (reset func(), stop func()) {
	if timeout > 0 {
		dur := time.Duration(timeout) * time.Second
		timer := time.AfterFunc(dur, func() {
			fmt.Fprintln(out, "Operation timed out")
			closeFn()
		})
		reset = func() {
			timer.Reset(dur)
		}
		stop = func() { timer.Stop() }
	} else {
		stop = func() {}
	}
	return
}

// --- Error helpers ---

type exitStatusError interface {
	error
	ExitStatus() int
}

// ExitCodeFromError extracts the remote command exit code from err, if present.
func ExitCodeFromError(err error) (int, bool) {
	var exitErr exitStatusError
	if errors.As(err, &exitErr) {
		return exitErr.ExitStatus(), true
	}
	return 0, false
}

// onInterrupt sets up a handler that calls cleanup when the process receives
// an interrupt signal (Ctrl+C). The cleanup should close the underlying
// connection so the main goroutine unblocks and exits through the normal path
// (running deferred functions). Only the first signal is honored.
//
// The returned stop function unregisters the handler and releases the
// goroutine; it must be called when the owning resource is closed (e.g. in
// Client.Close) to avoid leaking the goroutine in long-running processes. The
// library does not register any signal handler unless explicitly requested via
// WithSignalHandler.
func onInterrupt(cleanup func()) (stop func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	done := make(chan struct{})
	go func() {
		select {
		case <-ch:
			cleanup()
		case <-done:
		}
	}()
	return func() {
		signal.Stop(ch)
		close(done)
	}
}
