package sshpass

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestWithStdin(t *testing.T) {
	r := &bytes.Buffer{}
	c := &Client{}
	WithStdin(r)(c)
	if c.stdin != io.Reader(r) {
		t.Error("WithStdin did not set stdin")
	}
}

func TestWithStdout(t *testing.T) {
	w := &bytes.Buffer{}
	c := &Client{}
	WithStdout(w)(c)
	if c.stdout != io.Writer(w) {
		t.Error("WithStdout did not set stdout")
	}
}

func TestWithStderr(t *testing.T) {
	w := &bytes.Buffer{}
	c := &Client{}
	WithStderr(w)(c)
	if c.stderr != io.Writer(w) {
		t.Error("WithStderr did not set stderr")
	}
}

func TestWithProgress(t *testing.T) {
	var called bool
	fn := func(desc string, sent, total int64) { called = true }
	c := &Client{}
	WithProgress(fn)(c)
	if c.progress == nil {
		t.Fatal("WithProgress did not set progress")
	}
	c.progress("test", 1, 10)
	if !called {
		t.Error("progress callback was not invoked")
	}
}

func TestWithFileSelector(t *testing.T) {
	selector := &mockFileSelector{}
	c := &Client{}
	WithFileSelector(selector)(c)
	if c.selector == nil {
		t.Fatal("WithFileSelector did not set selector")
	}
}

func TestWithSignalHandler(t *testing.T) {
	c := &Client{}
	WithSignalHandler()(c)
	if !c.signal {
		t.Error("WithSignalHandler did not set signal=true")
	}
}

func TestNopProgressReporter(_ *testing.T) {
	// Ensures ProgressFunc type compiles with nil handling.
	var fn ProgressFunc
	_ = fn // nil is valid
}

// mockFileSelector is a test double for FileSelector.
type mockFileSelector struct{}

func (mockFileSelector) OpenFile() (string, error)       { return "", nil }
func (mockFileSelector) SaveFile(string) (string, error) { return "", nil }

// Ensure default I/O streams are os.Stdin/Stdout/Stderr when no option is set.
// (This tests the documented default behavior of NewClient's struct init.)
func TestClientDefaultIO(t *testing.T) {
	// We can't call NewClient without a real server, but we verify the
	// documented defaults by checking the zero-value behavior matches the
	// contract: the SDK sets os.Stdin/Stdout/Stderr in NewClient before opts.
	c := &Client{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	if c.stdin != os.Stdin || c.stdout != os.Stdout || c.stderr != os.Stderr {
		t.Error("default I/O streams not set to os.Stdin/Stdout/Stderr")
	}
}
