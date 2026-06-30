package sshpass

import "testing"

func TestDetectCommandTypeOnlyUsesLeadingCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want CommandType
	}{
		{name: "empty defaults to ssh", args: nil, want: CommandSSH},
		{name: "scp subcommand", args: []string{"scp", "file.txt", "user@example.com:/tmp/"}, want: CommandSCP},
		{name: "rsync subcommand", args: []string{"rsync", "-avz", "./", "user@example.com:/backup/"}, want: CommandRsync},
		{name: "ssh command containing scp stays ssh", args: []string{"ssh", "user@example.com", "scp", "file.txt"}, want: CommandSSH},
		{name: "plain remote command containing rsync stays ssh", args: []string{"ssh", "user@example.com", "rsync", "--version"}, want: CommandSSH},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectCommandType(tt.args); got != tt.want {
				t.Fatalf("DetectCommandType(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestParseSSHArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantUser    string
		wantHost    string
		wantPort    string
		wantKeyPath string
		wantCmd     string
	}{
		{
			name:     "user@host",
			args:     []string{"ssh", "root@1.2.3.4"},
			wantUser: "root", wantHost: "1.2.3.4",
			wantPort: "", wantKeyPath: "", wantCmd: "",
		},
		{
			name:     "user@host with command",
			args:     []string{"ssh", "deploy@example.com", "ls", "-la"},
			wantUser: "deploy", wantHost: "example.com",
			wantPort: "", wantKeyPath: "", wantCmd: "ls -la",
		},
		{
			name:     "custom port",
			args:     []string{"ssh", "-p", "2222", "root@host"},
			wantUser: "root", wantHost: "host",
			wantPort: "2222", wantKeyPath: "", wantCmd: "",
		},
		{
			name:     "identity file",
			args:     []string{"ssh", "-i", "/key", "root@host"},
			wantUser: "root", wantHost: "host",
			wantPort: "", wantKeyPath: "/key", wantCmd: "",
		},
		{
			name:     "skip -o option",
			args:     []string{"ssh", "-o", "StrictHostKeyChecking=no", "root@host"},
			wantUser: "root", wantHost: "host",
			wantPort: "", wantKeyPath: "", wantCmd: "",
		},
		{
			name:     "skip -v flag",
			args:     []string{"ssh", "-v", "root@host"},
			wantUser: "root", wantHost: "host",
			wantPort: "", wantKeyPath: "", wantCmd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, cmd := ParseSSHArgs(tt.args)
			if cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if cfg.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", cfg.Host, tt.wantHost)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
			if cfg.KeyPath != tt.wantKeyPath {
				t.Errorf("KeyPath = %q, want %q", cfg.KeyPath, tt.wantKeyPath)
			}
			if cmd != tt.wantCmd {
				t.Errorf("Cmd = %q, want %q", cmd, tt.wantCmd)
			}
		})
	}
}

func TestParseSCPArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantUser string
		wantHost string
		wantPort string
		wantNArg int // number of scp args remaining
	}{
		{
			name:     "upload file",
			args:     []string{"scp", "file.txt", "root@host:/tmp/"},
			wantUser: "root", wantHost: "host", wantPort: "", wantNArg: 2,
		},
		{
			name:     "with -P port",
			args:     []string{"scp", "-P", "2222", "file.txt", "root@host:/tmp/"},
			wantUser: "root", wantHost: "host", wantPort: "2222", wantNArg: 2,
		},
		{
			name:     "with -r flag skipped",
			args:     []string{"scp", "-r", "dir/", "root@host:/tmp/"},
			wantUser: "root", wantHost: "host", wantPort: "", wantNArg: 2,
		},
		{
			name:     "with -q and -C flags skipped",
			args:     []string{"scp", "-q", "-C", "file.txt", "root@host:/tmp/"},
			wantUser: "root", wantHost: "host", wantPort: "", wantNArg: 2,
		},
		{
			name:     "with -i key",
			args:     []string{"scp", "-i", "/key", "file.txt", "root@host:/tmp/"},
			wantUser: "root", wantHost: "host", wantPort: "", wantNArg: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, scpArgs := ParseSCPArgs(tt.args)
			if cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if cfg.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", cfg.Host, tt.wantHost)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
			if len(scpArgs) != tt.wantNArg {
				t.Errorf("len(scpArgs) = %d, want %d, args=%v", len(scpArgs), tt.wantNArg, scpArgs)
			}
		})
	}
}

func TestParseRsyncArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantUser string
		wantHost string
		wantPort string
	}{
		{
			name:     "upload",
			args:     []string{"rsync", "-avz", "./", "root@host:/backup/"},
			wantUser: "root", wantHost: "host", wantPort: "",
		},
		{
			name:     "with -e ssh",
			args:     []string{"rsync", "-e", "ssh", "./", "root@host:/backup/"},
			wantUser: "root", wantHost: "host", wantPort: "",
		},
		{
			name:     "with --rsh=ssh",
			args:     []string{"rsync", "--rsh=ssh", "./", "root@host:/backup/"},
			wantUser: "root", wantHost: "host", wantPort: "",
		},
		{
			name:     "with --port=2222",
			args:     []string{"rsync", "--port=2222", "./", "root@host:/backup/"},
			wantUser: "root", wantHost: "host", wantPort: "2222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := ParseRsyncArgs(tt.args)
			if cfg.User != tt.wantUser {
				t.Errorf("User = %q, want %q", cfg.User, tt.wantUser)
			}
			if cfg.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", cfg.Host, tt.wantHost)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
		})
	}
}
