package main

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
			if got := detectCommandType(tt.args); got != tt.want {
				t.Fatalf("detectCommandType(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
