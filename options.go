package sshpass

import "io"

// ProgressFunc is a callback invoked during file transfers to report progress
// as plain numbers. description identifies the transfer (e.g. "Uploading
// file.txt"). sent is the number of bytes transferred so far; total is the
// file size (always > 0 in current implementations). The callback is invoked
// on every data chunk and once with sent == 0 at the start of a transfer.
//
// The SDK performs no rendering itself; callers decide how to display progress
// (progress bar, log line, metric, etc.) based on these numbers. A nil
// ProgressFunc means no progress reporting.
type ProgressFunc func(description string, sent, total int64)

// FileSelector abstracts interactive file open/save dialogs used by the rz/sz
// shell-transfer fallback. The SDK does not provide a default implementation;
// callers that need rz/sz support must inject one (e.g. a GUI dialog or a
// stdin-based prompter). When no FileSelector is configured, rz/sz falls back
// to reading a path from stdin.
type FileSelector interface {
	// OpenFile prompts the user to select an existing file. It returns "" with
	// a nil error when the user cancels.
	OpenFile() (string, error)
	// SaveFile prompts the user to choose a destination path for saving a file.
	// defaultName is suggested as the initial filename. It returns "" with a
	// nil error when the user cancels.
	SaveFile(defaultName string) (string, error)
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithStdin sets the input stream used by Exec and Shell. Defaults to os.Stdin.
func WithStdin(r io.Reader) Option {
	return func(c *Client) { c.stdin = r }
}

// WithStdout sets the output stream used by Exec and Shell. Defaults to
// os.Stdout.
func WithStdout(w io.Writer) Option {
	return func(c *Client) { c.stdout = w }
}

// WithStderr sets the error/diagnostic stream used for retry messages and the
// operation-timeout notice. Defaults to os.Stderr.
func WithStderr(w io.Writer) Option {
	return func(c *Client) { c.stderr = w }
}

// WithProgress sets the callback used to report SFTP transfer progress as
// byte counts. By default no progress callback is set (headless-friendly).
func WithProgress(fn ProgressFunc) Option {
	return func(c *Client) { c.progress = fn }
}

// WithFileSelector sets the FileSelector used by the rz/sz shell-transfer
// fallback. When not set, rz/sz prompts for a path on stdin.
func WithFileSelector(s FileSelector) Option {
	return func(c *Client) { c.selector = s }
}

// WithSignalHandler enables registration of an os.Interrupt handler that
// closes the Client's underlying connection on Ctrl+C, allowing the main
// goroutine to unblock and run deferred cleanup. By default no signal handler
// is registered so the library does not interfere with the host process's
// signal handling.
func WithSignalHandler() Option {
	return func(c *Client) { c.signal = true }
}
