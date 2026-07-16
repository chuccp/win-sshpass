package sshpass

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestProgressWriterCountsAndCallbacks(t *testing.T) {
	var calls []struct {
		desc  string
		sent  int64
		total int64
	}
	fn := func(desc string, sent, total int64) {
		calls = append(calls, struct {
			desc  string
			sent  int64
			total int64
		}{desc, sent, total})
	}

	var underlying bytes.Buffer
	pw := &progressWriter{w: &underlying, desc: "uploading", total: 100, fn: fn}

	n, err := pw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 5 {
		t.Errorf("wrote %d bytes, want 5", n)
	}
	if underlying.String() != "hello" {
		t.Errorf("underlying = %q, want hello", underlying.String())
	}
	if len(calls) != 1 {
		t.Fatalf("expected 1 callback, got %d", len(calls))
	}
	if calls[0].sent != 5 || calls[0].total != 100 || calls[0].desc != "uploading" {
		t.Errorf("callback = %+v, want sent=5 total=100 desc=uploading", calls[0])
	}

	// Second write accumulates.
	pw.Write([]byte(" world"))
	if len(calls) != 2 {
		t.Fatalf("expected 2 callbacks, got %d", len(calls))
	}
	if calls[1].sent != 11 {
		t.Errorf("second callback sent = %d, want 11", calls[1].sent)
	}
}

func TestProgressWriterNilCallbackNoPanic(t *testing.T) {
	var underlying bytes.Buffer
	pw := &progressWriter{w: &underlying, total: 100, fn: nil}
	_, err := pw.Write([]byte("data"))
	if err != nil {
		t.Fatalf("write with nil fn should not error: %v", err)
	}
}

func TestProgressReaderCountsAndCallbacks(t *testing.T) {
	var lastSent int64
	fn := func(desc string, sent, total int64) {
		lastSent = sent
	}

	source := bytes.NewReader([]byte("abcdefghij")) // 10 bytes
	pr := &progressReader{r: source, desc: "downloading", total: 10, fn: fn}

	buf := make([]byte, 4)
	n, _ := pr.Read(buf)
	if n != 4 {
		t.Errorf("read %d bytes, want 4", n)
	}
	if lastSent != 4 {
		t.Errorf("lastSent = %d, want 4", lastSent)
	}

	n, _ = pr.Read(buf)
	if lastSent != 8 {
		t.Errorf("lastSent = %d, want 8 (accumulated)", lastSent)
	}
}

func TestTimeoutWriterCallsReset(t *testing.T) {
	var resetCount int
	reset := func() { resetCount++ }

	var underlying bytes.Buffer
	tw := &timeoutWriter{w: &underlying, reset: reset}

	tw.Write([]byte("data"))
	if resetCount != 1 {
		t.Errorf("reset called %d times, want 1", resetCount)
	}
}

func TestTimeoutWriterNilResetNoPanic(t *testing.T) {
	var underlying bytes.Buffer
	tw := &timeoutWriter{w: &underlying, reset: nil}
	_, err := tw.Write([]byte("data"))
	if err != nil {
		t.Fatalf("write with nil reset should not error: %v", err)
	}
}

func TestTimeoutReaderCallsReset(t *testing.T) {
	var resetCount int
	reset := func() { resetCount++ }

	source := bytes.NewReader([]byte("data"))
	tr := &timeoutReader{r: source, reset: reset}

	buf := make([]byte, 4)
	tr.Read(buf)
	if resetCount != 1 {
		t.Errorf("reset called %d times, want 1", resetCount)
	}
}

func TestProgressWriterWriteError(t *testing.T) {
	// A writer that always errors.
	errWriter := &errWriter{err: errors.New("disk full")}
	pw := &progressWriter{w: errWriter, fn: func(string, int64, int64) {}}
	_, err := pw.Write([]byte("data"))
	if err == nil {
		t.Fatal("expected error from underlying writer")
	}
	if err.Error() != "disk full" {
		t.Errorf("err = %v, want disk full", err)
	}
}

// errWriter is a test helper that always returns an error on Write.
type errWriter struct{ err error }

func (w *errWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestProgressReaderReadError(t *testing.T) {
	source := &errReader{err: errors.New("read failed")}
	pr := &progressReader{r: source, fn: func(string, int64, int64) {}}
	_, err := pr.Read(make([]byte, 4))
	if err == nil {
		t.Fatal("expected error from underlying reader")
	}
}

// errReader is a test helper that always returns an error on Read.
type errReader struct{ err error }

func (r *errReader) Read(p []byte) (int, error) {
	return 0, r.err
}

// Ensure progressWriter/progressReader implement io.Writer/io.Reader.
func TestProgressTypesImplementIO(t *testing.T) {
	var _ io.Writer = (*progressWriter)(nil)
	var _ io.Reader = (*progressReader)(nil)
	var _ io.Writer = (*timeoutWriter)(nil)
	var _ io.Reader = (*timeoutReader)(nil)
}
