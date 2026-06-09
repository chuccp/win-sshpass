package main

import (
	"fmt"
	"strings"
)

// CommandType represents the command type
type CommandType int

const (
	CommandSSH CommandType = iota
	CommandSCP
	CommandRsync
)

// parseSSHArgs parses ssh-style arguments (user@host or -p port user@host)
func parseSSHArgs(args []string) (*Config, string) {
	config := newDefaultConfig()
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
			command = joinArgs(args[i:])
			break
		}
		i++
	}

	return config, command
}

// parseSCPArgs parses scp command arguments
func parseSCPArgs(args []string) (*Config, []string) {
	config := newDefaultConfig()
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
			config.setUserHostFromArg(arg)
		}
		scpArgs = append(scpArgs, arg)
		i++
	}

	return config, scpArgs
}

// parseRsyncArgs parses rsync command arguments
func parseRsyncArgs(args []string) (*Config, []string) {
	config := newDefaultConfig()
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
			config.setUserHostFromArg(arg)
		}
		rsyncArgs = append(rsyncArgs, arg)
		i++
	}

	return config, rsyncArgs
}

// detectCommandType detects the command type
func detectCommandType(args []string) CommandType {
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

// printUsage prints the usage
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] ssh [user@host] [command]")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] scp [-r] [options] [user@host:]path")
	fmt.Println("  sshpass [-p <password> | -f <passfile>] rsync [options] [user@host:]path")
	fmt.Println("  sshpass -i <keypath> ssh [user@host] [command]")
	fmt.Println("  sshpass -f <configfile> [-c <command>]")
	fmt.Println("  sshpass -h <host> -p <password> [-u <user>] [-P <port>]")
	fmt.Println("  sshpass -h <host> -p <password> -local <file> -remote <path>  (upload)")
	fmt.Println("  sshpass -h <host> -p <password> -local <path> -remote <file> -d (download)")
	fmt.Println("\nOptions:")
	fmt.Println("  -p <password>      specify password directly")
	fmt.Println("  -f <file>          read password from file (single line) or config file")
	fmt.Println("  -i <key>           use private key authentication")
	fmt.Println("  -e                 read password from environment variable SSHPASS")
	fmt.Println("  -k                 enable strict host key verification (use known_hosts file)")
	fmt.Println("  -t <seconds>       total operation timeout in seconds (0 = no limit, default: 0)")
	fmt.Println("  -ct <seconds>      TCP connection timeout in seconds (default: 10)")
	fmt.Println("  -retry <n>         total connection attempts (default: 3, 0 = no retry)")
	fmt.Println("  -local <path>      local file path(s), comma-separated for multiple files")
	fmt.Println("  -remote <path>     remote file path (for upload/download)")
	fmt.Println("  -d                 download mode (remote to local)")
	fmt.Println("  -v                 show version")
	fmt.Println("  -help              show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  sshpass -p 'pass' ssh user@example.com 'whoami'")
	fmt.Println("  sshpass -f pass.txt ssh user@example.com")
	fmt.Println("  SSHPASS='pass' sshpass -e ssh user@example.com")
	fmt.Println("  sshpass -i ~/.ssh/id_ed25519 ssh user@example.com")
	fmt.Println("  sshpass -p 'pass' scp -r file.txt user@example.com:/tmp/")
	fmt.Println("  sshpass -p 'pass' rsync -avz ./ user@example.com:/backup/")
	fmt.Println("  sshpass -p 'pass' -h example.com -local file1.txt,file2.txt -remote /tmp/")
	fmt.Println("  sshpass -p 'pass' -h example.com -local ./backup -remote /data/file.tar.gz -d")
}

// printVersion prints version info
func printVersion() {
	fmt.Printf("sshpass version %s (Windows)\n", version)
}
