package main

import (
	"flag"
	"fmt"
	"os"
	"sync/atomic"
)

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
	timeout := flag.Int("t", 0, "total operation timeout in seconds (0 = no limit)")
	connectTimeout := flag.Int("ct", 10, "TCP connection timeout in seconds")
	retries := flag.Int("retry", 3, "total connection attempts (default 3)")
	showVersion := flag.Bool("v", false, "show version")
	showHelp := flag.Bool("help", false, "show help")
	flag.Parse()

	// display help
	if *showHelp {
		printUsage()
		return
	}

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
	cliOverride := &Config{
		Password:       pass,
		KeyPath:        *keyPath,
		Host:           *host,
		User:           *user,
		Timeout:        -1, // -1 = not set; allows explicit 0 to override config file
		ConnectTimeout: -1, // -1 = not set; allows explicit 0 to override config file
		Retries:        -1, // -1 = not set; allows explicit 0 to override config file
	}
	// Only include Port, Timeout, ConnectTimeout, Retries, and StrictHostKey
	// if explicitly set by the user. Without flag.Visit, their default values
	// would always override config file values since we can't distinguish
	// "not set" from "set to default".
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "P":
			cliOverride.Port = *port
		case "t":
			cliOverride.Timeout = *timeout
		case "ct":
			cliOverride.ConnectTimeout = *connectTimeout
		case "retry":
			cliOverride.Retries = *retries
		case "k":
			cliOverride.StrictHostKey = *strictHostKey
		}
	})
	if config != nil {
		mergeConfig(config, nil, cliOverride)
	}

	// detect command type
	cmdType := detectCommandType(remainingArgs)

	// handle based on command type
	switch cmdType {
	case CommandSCP:
		scpParsed, scpArgs := parseSCPArgs(remainingArgs)
		cfgConfig := newDefaultConfig()
		mergeConfig(cfgConfig, config, scpParsed)  // config file as src, scp-parsed as override
		mergeConfig(cfgConfig, nil, cliOverride)    // CLI as final override
		applyUserDefault(cfgConfig)
		cfgConfig.normalize()
		config = cfgConfig
		if err := runSCP(config, scpArgs); err != nil {
			fatalError("SCP failed: %v", err)
		}
		return

	case CommandRsync:
		rsyncParsed, rsyncArgs := parseRsyncArgs(remainingArgs)
		cfgConfig := newDefaultConfig()
		mergeConfig(cfgConfig, config, rsyncParsed) // config file as src, rsync-parsed as override
		mergeConfig(cfgConfig, nil, cliOverride)     // CLI as final override
		applyUserDefault(cfgConfig)
		cfgConfig.normalize()
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
			mergeConfig(config, nil, cliOverride)
		} else if *host != "" && (pass != "" || *keyPath != "") {
			// read from command line arguments (including file transfer mode)
			config = newDefaultConfig()
			mergeConfig(config, nil, cliOverride)
		} else {
			printUsage()
			os.Exit(1)
		}
	} else if len(remainingArgs) > 0 {
		// config from file, but remaining args may override host/user or provide command
		sshArgs, cmd := parseSSHArgs(remainingArgs)
		if sshArgs.Host != "" {
			config.Host = sshArgs.Host
			if sshArgs.User != "" {
				config.User = sshArgs.User
			}
		}
		if sshArgs.Port != "" {
			config.Port = sshArgs.Port
		}
		if sshArgs.KeyPath != "" {
			config.KeyPath = sshArgs.KeyPath
		}
		if cmd != "" {
			cmdToRun = cmd
		}
	}

	// apply defaults and normalize
	applyUserDefault(config)
	config.normalize()
	if config.Port == "" {
		config.Port = "22"
	}

	// validate config
	if err := config.validate(); err != nil {
		fatalError("Error: %v", err)
	}

	// file transfer path — uses connectSFTP which manages its own SSH connection,
	// timeout, and interrupt handling.
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

		conn, err := connectSFTP(config)
		if err != nil {
			fatalError("SFTP connection failed: %v", err)
		}
		defer conn.Close()

		if *download {
			for _, rPath := range remotePaths {
				for _, lp := range localPaths {
					fmt.Printf("Downloading %s -> %s...\n", rPath, lp)
					if err := conn.Download(rPath, lp); err != nil {
						fatalError("Download failed: %v", err)
					}
				}
			}
			fmt.Println("Download successful!")
		} else {
			for _, lp := range localPaths {
				for _, rPath := range remotePaths {
					fmt.Printf("Uploading %s -> %s...\n", lp, rPath)
					if err := conn.Upload(lp, rPath); err != nil {
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

	// command/shell path
	client, err := SSHClient(config)
	if err != nil {
		fatalError("SSH connection failed: %v", err)
	}
	onInterrupt(func() { client.Close() })
	defer client.Close()

	// set up operation timeout (timer resets on each data transfer)
	var timedOut atomic.Bool
	_, stopTimer := setupOperationTimeout(func() {
		timedOut.Store(true)
		client.Close()
	}, config.Timeout)
	defer stopTimer()

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

	if err != nil {
		if timedOut.Load() {
			os.Exit(1)
		}
		if !isClosedConnError(err) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if code, ok := exitCodeFromError(err); ok {
				os.Exit(code)
			}
			os.Exit(1)
		}
	}
}
