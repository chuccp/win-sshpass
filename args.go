package sshpass

import "strings"

// CommandType represents the command type.
type CommandType int

const (
	CommandSSH   CommandType = iota // ssh session
	CommandSCP                      // scp file transfer
	CommandRsync                    // rsync file transfer
)

// ParseSSHArgs parses ssh-style arguments (user@host or -p port user@host).
// It returns a Config populated from the arguments and any trailing command
// string.
func ParseSSHArgs(args []string) (*Config, string) {
	config := NewConfig()
	config.Port = "" // clear default; only set if -p is explicitly used
	var command string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "ssh" {
			// skip the ssh command itself
			i++
			continue
		}
		if arg == "-p" && i+1 < len(args) {
			config.Port = args[i+1]
			i += 2
			continue
		}
		if arg == "-i" && i+1 < len(args) {
			config.KeyPath = args[i+1]
			i += 2
			continue
		}
		if arg == "-o" && i+1 < len(args) {
			// skip ssh options like StrictHostKeyChecking=no
			i += 2
			continue
		}
		if arg == "-v" {
			// skip verbose flag (not supported natively)
			i++
			continue
		}
		if strings.Contains(arg, "@") {
			// user@host format (supports IPv6)
			parts := strings.SplitN(arg, "@", 2)
			if len(parts) == 2 {
				config.User = parts[0]
				config.Host = parts[1]
			}
			i++
			continue
		}
		// remaining args as command
		if config.Host != "" {
			command = JoinArgs(args[i:])
			break
		}
		i++
	}

	return config, command
}

// ParseSCPArgs parses scp command arguments. It returns a Config populated from
// scp-specific flags and the remaining scp arguments (including any
// user@host:path token).
func ParseSCPArgs(args []string) (*Config, []string) {
	config := NewConfig()
	config.User = "" // clear default; only set if user@host:path is found
	config.Port = "" // clear default; only set if -P is explicitly used
	var scpArgs []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "scp" {
			i++
			continue
		}
		if arg == "-P" && i+1 < len(args) {
			// scp uses uppercase -P for port
			config.Port = args[i+1]
			i += 2
			continue
		}
		if arg == "-i" && i+1 < len(args) {
			config.KeyPath = args[i+1]
			i += 2
			continue
		}
		if arg == "-o" && i+1 < len(args) {
			i += 2
			continue
		}
		if arg == "-r" || arg == "-q" || arg == "-C" || arg == "-v" {
			// flags handled natively by SFTP implementation: recursive, quiet, compression, verbose
			i++
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			config.SetUserHostFromArg(arg)
		}
		scpArgs = append(scpArgs, arg)
		i++
	}

	return config, scpArgs
}

// ParseRsyncArgs parses rsync command arguments. It returns a Config populated
// from rsync-specific flags and the remaining rsync arguments (including any
// user@host:path token).
func ParseRsyncArgs(args []string) (*Config, []string) {
	config := NewConfig()
	config.User = "" // clear default; only set if user@host:path is found
	config.Port = "" // clear default; only set if --port= is explicitly used
	var rsyncArgs []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "rsync" {
			i++
			continue
		}
		if arg == "-e" && i+1 < len(args) {
			// skip -e ssh option
			i += 2
			continue
		}
		if strings.HasPrefix(arg, "--rsh=") {
			// skip --rsh=ssh option
			i++
			continue
		}
		if strings.HasPrefix(arg, "--port=") {
			// --port=N format (rsync uses -p for --perms, not port)
			portPart := arg[len("--port="):]
			if isAllDigits(portPart) {
				config.Port = portPart
			}
			i++
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			config.SetUserHostFromArg(arg)
		}
		rsyncArgs = append(rsyncArgs, arg)
		i++
	}

	return config, rsyncArgs
}

// DetectCommandType detects the command type from the leading argument.
func DetectCommandType(args []string) CommandType {
	if len(args) == 0 {
		return CommandSSH
	}
	switch args[0] {
	case "scp":
		return CommandSCP
	case "rsync":
		return CommandRsync
	default:
		return CommandSSH
	}
}
