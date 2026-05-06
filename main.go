package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// splitPaths splits a path string by comma or space separator.
// Returns error if complex paths (containing '/' or '\') are space-separated.
// name identifies which parameter for error messages (e.g., "local" or "remote").
func splitPaths(s, name string) ([]string, error) {
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

func main() {
	// command line arguments
	configFile := flag.String("f", "", "password file or config file path")
	host := flag.String("h", "", "host address")
	user := flag.String("u", "", "username (default: root)")
	password := flag.String("p", "", "password")
	port := flag.String("P", "22", "port")
	keyPath := flag.String("i", "", "private key file path")
	command := flag.String("c", "", "command to execute")
	localPath := flag.String("local", "", "local file path(s), comma or space separated for multiple files (for upload/download)")
	remotePath := flag.String("remote", "", "remote file path (for upload/download)")
	download := flag.Bool("d", false, "download mode (remote to local)")
	useEnv := flag.Bool("e", false, "read password from environment variable SSHPASS")
	strictHostKey := flag.Bool("k", false, "enable strict host key verification")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	// display version
	if *showVersion {
		printVersion()
		return
	}

	var config *Config
	var err error
	var cmdToRun string

	// get remaining arguments (for sshpass-style commands)
	remainingArgs := flag.Args()

	// get password: priority -p > config file > password file > -e > SSHPASS
	pass := *password
	if *configFile != "" {
		config, pass, err = loadConfigOrPasswordFile(*configFile, pass, *strictHostKey)
		if err != nil {
			fatalError("Error: %v", err)
		}
	}
	if pass == "" && *useEnv {
		pass = getEnvPassword()
	}
	if config != nil {
		mergeConfig(config, nil, pass, *keyPath, *host, *user, *port)
		config.StrictHostKey = *strictHostKey
	}

	// detect command type
	cmdType := detectCommandType(remainingArgs)

	// handle based on command type
	switch cmdType {
	case CommandSCP:
		cfgConfig, scpArgs := parseSCPArgs(remainingArgs)
		mergeConfig(cfgConfig, config, pass, *keyPath, *host, *user, *port)
		cfgConfig.StrictHostKey = *strictHostKey
		config = cfgConfig
		if err := runSCP(config, scpArgs); err != nil {
			fatalError("SCP failed: %v", err)
		}
		return

	case CommandRsync:
		cfgConfig, rsyncArgs := parseRsyncArgs(remainingArgs)
		mergeConfig(cfgConfig, config, pass, *keyPath, *host, *user, *port)
		cfgConfig.StrictHostKey = *strictHostKey
		config = cfgConfig
		if err := runRsync(config, rsyncArgs); err != nil {
			fatalError("Rsync failed: %v", err)
		}
		return
	}

	// SSH command handling
	if config == nil {
		if len(remainingArgs) > 0 && (pass != "" || *keyPath != "") {
			// sshpass style: -p password or -i keyfile ssh user@host [command]
			config, cmdToRun = parseSSHArgs(remainingArgs)
			// if -h flag was used and no user@host found in args, use remaining args as command
			if config.Host == "" && *host != "" {
				config.Host = *host
				cmdToRun = joinArgs(remainingArgs)
			}
			if pass != "" {
				config.Password = pass
			}
			if *keyPath != "" {
				config.KeyPath = *keyPath
			}
			if *user != "" {
				config.User = *user
			}
			if *port != "" && *port != "22" {
				config.Port = *port
			}
			config.StrictHostKey = *strictHostKey
		} else if *host != "" && (pass != "" || *keyPath != "") {
			// read from command line arguments (including file transfer mode)
			config = newDefaultConfig()
			config.Host = *host
			config.Password = pass
			config.Port = *port
			config.KeyPath = *keyPath
			config.StrictHostKey = *strictHostKey
			if *user != "" {
				config.User = *user
			}
		} else {
			printUsage()
			os.Exit(1)
		}
	}

	// apply default user if still empty
	applyUserDefault(config)

	// validate config
	if err := config.validate(); err != nil {
		fatalError("Error: %v", err)
	}

	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		fatalError("SSH connection failed: %v", err)
	}
	defer client.Close()

	// file transfer
	if *localPath != "" && *remotePath != "" {
		localPaths, err := splitPaths(*localPath, "local")
		if err != nil {
			fatalError("Error: %v", err)
		}
		remotePaths, err := splitPaths(*remotePath, "remote")
		if err != nil {
			fatalError("Error: %v", err)
		}
		for i := range remotePaths {
			remotePaths[i] = cleanRemotePath(remotePaths[i])
		}

		if *download {
			// ensure local target directories exist
			for _, lp := range localPaths {
				if err := os.MkdirAll(lp, 0755); err != nil {
					fatalError("Error: failed to create local directory %q: %v", lp, err)
				}
			}
			for _, rPath := range remotePaths {
				for _, lp := range localPaths {
					fmt.Printf("Downloading %s -> %s...\n", rPath, lp)
					if err := downloadFile(client, rPath, lp); err != nil {
						fatalError("Download failed: %v", err)
					}
				}
			}
			fmt.Println("Download successful!")
		} else {
			// ensure remote target directories exist
			for _, rPath := range remotePaths {
				if err := ensureRemoteDir(client, rPath); err != nil {
					fatalError("Error: failed to create remote directory %q: %v", rPath, err)
				}
			}
			for _, lp := range localPaths {
				for _, rPath := range remotePaths {
					fmt.Printf("Uploading %s -> %s...\n", lp, rPath)
					if err := uploadFile(client, lp, rPath); err != nil {
						fatalError("Upload failed: %v", err)
					}
				}
			}
			fmt.Println("Upload successful!")
		}
		return
	} else if *localPath != "" || *remotePath != "" {
		fatalError("Error: file transfer requires both -local and -remote arguments")
	}

	// execute command
	cmd := *command
	if cmd == "" {
		cmd = cmdToRun
	}

	if cmd != "" {
		err = executeCommand(client, cmd)
	} else {
		err = runShell(client)
	}

	if err != nil && !isClosedConnError(err) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if code, ok := exitCodeFromError(err); ok {
			os.Exit(code)
		}
		os.Exit(1)
	}
}
