package sshpass

import (
	"bytes"
	"testing"
)

func TestExtractCommandFromNotFound(t *testing.T) {
	tests := []struct {
		name     string
		buf      string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "sz command bash",
			buf:      "[root@host ~]# sz CLAUDE.md\r\n-bash: sz: command not found\r\n",
			wantCmd:  "sz",
			wantArgs: []string{"CLAUDE.md"},
		},
		{
			name:     "rz command bash",
			buf:      "[root@host ~]# rz\r\n-bash: rz: command not found\r\n",
			wantCmd:  "rz",
			wantArgs: []string{},
		},
		{
			name:     "rz no args",
			buf:      "rz\r\n-bash: rz: command not found\r\n",
			wantCmd:  "rz",
			wantArgs: []string{},
		},
		{
			name:     "sz with tab completion",
			buf:      "[root@host ~]# sz CLAUDE.md\r\n-bash: sz: command not found\r\n",
			wantCmd:  "sz",
			wantArgs: []string{"CLAUDE.md"},
		},
		{
			name:     "ls not found - should not trigger",
			buf:      "[root@host ~]# nonexistent_cmd\r\n-bash: nonexistent_cmd: command not found\r\n",
			wantCmd:  "",
			wantArgs: nil,
		},
		{
			name:     "zsh format",
			buf:      "[host]# sz /path/file\r\nzsh: command not found: sz\r\n",
			wantCmd:  "sz",
			wantArgs: []string{"/path/file"},
		},
		{
			name:     "old sz in buffer should not false-positive",
			buf:      "[root@host ~]# ls\r\nfile_sz.txt\r\n[root@host ~]# rz\r\n-bash: rz: command not found\r\n",
			wantCmd:  "rz",
			wantArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := extractCommandFromNotFound([]byte(tt.buf))
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
				return
			}
			for i, a := range args {
				if a != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, a, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestContainsNotFound(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"-bash: rz: command not found\r\n", true},
		{"zsh: command not found: rz\r\n", true},
		{"sh: rz: not found\r\n", true},
		{"total 0\r\n", false},
		{"drwxr-xr-x 2 root root 4096 Jun 23 10:00 .\r\n", false},
	}

	for _, tt := range tests {
		if got := containsNotFound([]byte(tt.input)); got != tt.want {
			t.Errorf("containsNotFound(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestOutputWriterTriggersSZ(t *testing.T) {
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	var szRemote string
	var szCalled bool
	monitor.onSZ = func(remote, local string) {
		szCalled = true
		szRemote = remote
	}

	// Simulate remote echoing the command
	writer.Write([]byte("sz CLAUDE.md\r\n"))
	// Simulate "command not found" error
	writer.Write([]byte("-bash: sz: command not found\r\n"))

	if !szCalled {
		t.Fatal("onSZ not called")
	}
	if szRemote != "CLAUDE.md" {
		t.Errorf("onSZ remote = %q, want %q", szRemote, "CLAUDE.md")
	}
	// Verify output was passed through
	if !bytes.Contains(outBuf.Bytes(), []byte("sz CLAUDE.md")) {
		t.Errorf("output missing echoed command")
	}
	if !bytes.Contains(outBuf.Bytes(), []byte("command not found")) {
		t.Errorf("output missing error message")
	}
}

func TestOutputWriterTriggersRZ(t *testing.T) {
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	var rzCalled bool
	monitor.onRZ = func(path string) {
		rzCalled = true
	}

	writer.Write([]byte("rz\r\n"))
	writer.Write([]byte("-bash: rz: command not found\r\n"))

	if !rzCalled {
		t.Fatal("onRZ not called")
	}
}

func TestOutputWriterNoTriggerForOtherCommands(t *testing.T) {
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	var called bool
	monitor.onRZ = func(path string) { called = true }
	monitor.onSZ = func(remote, local string) { called = true }

	writer.Write([]byte("ls\r\n"))
	writer.Write([]byte("-bash: ls: command not found\r\n"))

	if called {
		t.Error("handler should not be called for ls")
	}
}

func TestOutputWriterPassthrough(t *testing.T) {
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	writer.Write([]byte("ls -la\r\ntotal 0\r\n"))

	if outBuf.String() != "ls -la\r\ntotal 0\r\n" {
		t.Errorf("Output = %q, want %q", outBuf.String(), "ls -la\r\ntotal 0\r\n")
	}
}

func TestContainsNotFoundChinese(t *testing.T) {
	// Chinese locale "未找到命令" should be detected.
	if !containsNotFound([]byte("-bash: rz: 未找到命令\r\n")) {
		t.Error("expected '未找到命令' to be detected as not-found")
	}
}

func TestContainsNotFoundNoSuchFile(t *testing.T) {
	// "No such file or directory" should be detected (covers sh fallback).
	if !containsNotFound([]byte("sh: rz: No such file or directory\r\n")) {
		t.Error("expected 'No such file or directory' to be detected")
	}
}

func TestExtractCommandFromNotFoundShFormat(t *testing.T) {
	// sh format: "sh: rz: not found"
	cmd, args := extractCommandFromNotFound([]byte("sh: rz: not found\r\n"))
	if cmd != "rz" {
		t.Errorf("cmd = %q, want rz", cmd)
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty", args)
	}
}

func TestExtractCommandFromNotFoundNotInBuffer(t *testing.T) {
	// "not found" present but no rz/sz command.
	cmd, _ := extractCommandFromNotFound([]byte("-bash: ls: command not found\r\n"))
	if cmd != "" {
		t.Errorf("cmd = %q, want empty (ls is not rz/sz)", cmd)
	}
}

func TestReadLineFromStdin(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello\n", "hello"},
		{"carriage return", "world\r\n", "world"},
		{"empty", "\n", ""},
		{"no newline", "noline", "noline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewBufferString(tt.input)
			got := readLineFromStdin(r)
			if got != tt.want {
				t.Errorf("readLineFromStdin(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestOutputWriterHandledFlagResets(t *testing.T) {
	// After handling rz, the Handled flag should reset to false, allowing a
	// subsequent rz/sz to be handled again.
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	callCount := 0
	monitor.onRZ = func(path string) { callCount++ }

	// First rz.
	writer.Write([]byte("rz\r\n"))
	writer.Write([]byte("-bash: rz: command not found\r\n"))
	if callCount != 1 {
		t.Fatalf("expected 1 call after first rz, got %d", callCount)
	}

	// Second rz should also trigger (Handled was reset).
	writer.Write([]byte("rz\r\n"))
	writer.Write([]byte("-bash: rz: command not found\r\n"))
	if callCount != 2 {
		t.Fatalf("expected 2 calls after second rz, got %d", callCount)
	}
}

func TestOutputWriterLargeOutputDoesNotPanic(t *testing.T) {
	// Writing a large chunk (> 4096 rolling buffer) should not panic.
	monitor := &rzszMonitor{}
	var outBuf bytes.Buffer
	writer := &outputWriter{monitor: monitor, out: &outBuf}

	large := make([]byte, 8192)
	for i := range large {
		large[i] = 'x'
	}
	writer.Write(large)
	if outBuf.Len() != 8192 {
		t.Errorf("output len = %d, want 8192", outBuf.Len())
	}
}
