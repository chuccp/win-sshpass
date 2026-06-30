package sshpass

import (
	"bytes"
	"testing"
)

func TestParseEchoedCommand(t *testing.T) {
	tests := []struct {
		name     string
		buf      string
		pending  string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "sz with prompt",
			buf:      "\r\n[root@host ~]# sz CLAUDE.md\r\n",
			pending:  "sz",
			wantCmd:  "sz",
			wantArgs: []string{"CLAUDE.md"},
		},
		{
			name:     "rz with prompt",
			buf:      "\r\n[root@host ~]# rz\r\n",
			pending:  "rz",
			wantCmd:  "rz",
			wantArgs: []string{},
		},
		{
			name:     "sz with local path",
			buf:      "[root@host ~]# sz /remote/file /local/file\r\n",
			pending:  "sz",
			wantCmd:  "sz",
			wantArgs: []string{"/remote/file", "/local/file"},
		},
		{
			name:     "rz with arg",
			buf:      "[root@host ~]# rz /path/to/file\r\n",
			pending:  "rz",
			wantCmd:  "rz",
			wantArgs: []string{"/path/to/file"},
		},
		{
			name:     "no command found",
			buf:      "some random output\r\n",
			pending:  "sz",
			wantCmd:  "",
			wantArgs: nil,
		},
		{
			name:     "tab completed sz",
			buf:      "sz CLAUDE.md\r\n",
			pending:  "sz",
			wantCmd:  "sz",
			wantArgs: []string{"CLAUDE.md"},
		},
		{
			name:     "rzbackup not matched",
			buf:      "[root@host ~]# rzbackup\r\n",
			pending:  "rz",
			wantCmd:  "",
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := parseEchoedCommand([]byte(tt.buf), tt.pending)
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
