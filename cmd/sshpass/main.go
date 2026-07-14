package main

import (
	"flag"
	"fmt"
	"os"

	sshpass "github.com/chuccp/win-sshpass"
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

	// CLI-side UI adapters: progress bar (stderr) and zenity file dialogs.
	// The SDK itself ships no UI; these are injected through options.
	cliOpts := []sshpass.Option{
		sshpass.WithProgress(newCLIProgress(os.Stderr).progress),
		sshpass.WithFileSelector(cliFileSelector{}),
		sshpass.WithSignalHandler(),
	}

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

	var config *sshpass.Config
	var err error
	var cmdToRun string

	// get remaining arguments (for sshpass-style commands)
	remainingArgs := flag.Args()

	// get password: priority -p > config file > password file > -e > SSHPASS
	pass := *password
	if *configFile != "" {
		config, pass, err = sshpass.LoadConfigOrPasswordFile(*configFile, pass, *strictHostKey)
		if err != nil {
			fatalError("Error: %v", err)
		}
	}
	if pass == "" && *useEnv {
		pass = sshpass.GetEnvPassword()
	}
	cliOverride := &sshpass.Config{
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
		config.MergeConfig(nil, cliOverride)
	}

	// --- hash/verify subcommands (local file operations, no SSH connection) ---
	if len(remainingArgs) > 0 {
		switch remainingArgs[0] {
		case "hash":
			if len(remainingArgs) != 3 {
				fatalError("Usage: sshpass hash <algorithm> <file>\nAlgorithms: md5, sha1, sha256, sha512")
			}
			result, err := sshpass.HashFile(remainingArgs[2], remainingArgs[1])
			if err != nil {
				fatalError("Error: %v", err)
			}
			fmt.Println(result)
			return
		case "verify":
			if len(remainingArgs) != 4 {
				fatalError("Usage: sshpass verify <algorithm> <hash> <file>\nAlgorithms: md5, sha1, sha256, sha512")
			}
			ok, err := sshpass.VerifyFile(remainingArgs[3], remainingArgs[1], remainingArgs[2])
			if err != nil {
				fatalError("Error: %v", err)
			}
			if ok {
				fmt.Println("OK")
			} else {
				fmt.Println("FAILED")
			}
			return
		}
	}

	// detect command type
	cmdType := sshpass.DetectCommandType(remainingArgs)

	// handle based on command type
	switch cmdType {
	case sshpass.CommandSCP:
		scpParsed, scpArgs := sshpass.ParseSCPArgs(remainingArgs)
		cfgConfig := sshpass.NewConfig()
		cfgConfig.MergeConfig(config, scpParsed) // config file as src, scp-parsed as override
		cfgConfig.MergeConfig(nil, cliOverride)  // CLI as final override
		cfgConfig.ApplyUserDefault()
		cfgConfig.Normalize()
		client, err := sshpass.NewClient(cfgConfig, cliOpts...)
		if err != nil {
			fatalError("SCP connection failed: %v", err)
		}
		defer client.Close()
		if err := sshpass.RunSCP(client, scpArgs); err != nil {
			fatalError("SCP failed: %v", err)
		}
		return

	case sshpass.CommandRsync:
		rsyncParsed, rsyncArgs := sshpass.ParseRsyncArgs(remainingArgs)
		cfgConfig := sshpass.NewConfig()
		cfgConfig.MergeConfig(config, rsyncParsed) // config file as src, rsync-parsed as override
		cfgConfig.MergeConfig(nil, cliOverride)    // CLI as final override
		cfgConfig.ApplyUserDefault()
		cfgConfig.Normalize()
		client, err := sshpass.NewClient(cfgConfig, cliOpts...)
		if err != nil {
			fatalError("Rsync connection failed: %v", err)
		}
		defer client.Close()
		if err := sshpass.RunRsync(client, rsyncArgs); err != nil {
			fatalError("Rsync failed: %v", err)
		}
		return
	}

	// SSH command handling
	if config == nil {
		if len(remainingArgs) > 0 && (pass != "" || *keyPath != "") {
			// sshpass style: -p password or -i keyfile ssh user@host [command]
			config, cmdToRun = sshpass.ParseSSHArgs(remainingArgs)
			// if -h flag was used and no user@host found in args, use remaining args as command
			if config.Host == "" && *host != "" {
				config.Host = *host
				cmdToRun = sshpass.JoinArgs(remainingArgs)
			}
			config.MergeConfig(nil, cliOverride)
		} else if *host != "" && (pass != "" || *keyPath != "") {
			// read from command line arguments (including file transfer mode)
			config = sshpass.NewConfig()
			config.MergeConfig(nil, cliOverride)
		} else {
			printUsage()
			os.Exit(1)
		}
	} else if len(remainingArgs) > 0 {
		// config from file, but remaining args may override host/user or provide command
		sshArgs, cmd := sshpass.ParseSSHArgs(remainingArgs)
		if sshArgs.Host != "" {
			config.Host = sshArgs.Host
			if sshArgs.User != "" {
				config.User = sshArgs.User
			}
		} else if cmd == "" {
			// no user@host in args; treat remaining args as a command
			// (skip leading "ssh" if present)
			cmdArgs := remainingArgs
			if len(cmdArgs) > 0 && cmdArgs[0] == "ssh" {
				cmdArgs = cmdArgs[1:]
			}
			cmd = sshpass.JoinArgs(cmdArgs)
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
	config.ApplyUserDefault()
	config.Normalize()
	if config.Port == "" {
		config.Port = "22"
	}

	// validate config
	if err := config.Validate(); err != nil {
		fatalError("Error: %v", err)
	}

	// file transfer path — uses client.SFTP which shares the Client's SSH
	// connection, timeout, and interrupt handling.
	if *localPath != "" && *remotePath != "" {
		localPaths, err := sshpass.SplitPaths(*localPath, "local")
		if err != nil {
			fatalError("Error: %v", err)
		}
		remotePaths, err := sshpass.SplitPaths(*remotePath, "remote")
		if err != nil {
			fatalError("Error: %v", err)
		}
		for i := range remotePaths {
			remotePaths[i], err = sshpass.CleanRemotePath(remotePaths[i])
			if err != nil {
				fatalError("Error: %v", err)
			}
		}

		client, err := sshpass.NewClient(config, cliOpts...)
		if err != nil {
			fatalError("SFTP connection failed: %v", err)
		}
		defer client.Close()

		conn, err := client.SFTP()
		if err != nil {
			fatalError("SFTP failed: %v", err)
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
	client, err := sshpass.NewClient(config, cliOpts...)
	if err != nil {
		fatalError("SSH connection failed: %v", err)
	}
	defer client.Close()

	// execute command
	cmd := *command
	if cmd == "" {
		cmd = cmdToRun
	}

	var runErr error
	if cmd != "" {
		runErr = client.Exec(cmd)
	} else {
		runErr = client.Shell()
	}

	if runErr != nil {
		if client.TimedOut() {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", runErr)
		if code, ok := sshpass.ExitCodeFromError(runErr); ok {
			os.Exit(code)
		}
		os.Exit(1)
	}
}

// fatalError prints an error message to stderr and exits with code 1.
func fatalError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// printUsage prints the usage.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  win-sshpass [-p <password> | -f <passfile>] ssh [user@host] [command]")
	fmt.Println("  win-sshpass [-p <password> | -f <passfile>] scp [-r] [options] [user@host:]path")
	fmt.Println("  win-sshpass [-p <password> | -f <passfile>] rsync [options] [user@host:]path")
	fmt.Println("  win-sshpass -i <keypath> ssh [user@host] [command]")
	fmt.Println("  win-sshpass -f <configfile> [-c <command>]")
	fmt.Println("  win-sshpass -h <host> -p <password> [-u <user>] [-P <port>]")
	fmt.Println("  win-sshpass -h <host> -p <password> -local <file> -remote <path>  (upload)")
	fmt.Println("  win-sshpass -h <host> -p <password> -local <path> -remote <file> -d (download)")
	fmt.Println("\nOptions:")
	fmt.Println("  -p <password>      specify password directly")
	fmt.Println("  -f <file>          read password from file (single line) or config file")
	fmt.Println("  -c <command>       command to execute on the remote host")
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
	fmt.Println("  win-sshpass -p 'pass' ssh user@example.com 'whoami'")
	fmt.Println("  win-sshpass -f pass.txt ssh user@example.com")
	fmt.Println("  SSHPASS='pass' win-sshpass -e ssh user@example.com")
	fmt.Println("  win-sshpass -i ~/.ssh/id_ed25519 ssh user@example.com")
	fmt.Println("  win-sshpass -p 'pass' scp -r file.txt user@example.com:/tmp/")
	fmt.Println("  win-sshpass -p 'pass' rsync -avz ./ user@example.com:/backup/")
	fmt.Println("  win-sshpass -p 'pass' -h example.com -local file1.txt,file2.txt -remote /tmp/")
	fmt.Println("  win-sshpass -p 'pass' -h example.com -local ./backup -remote /data/file.tar.gz -d")
	fmt.Println("\nSDK usage (as a Go library):")
	fmt.Println("  import \"github.com/chuccp/win-sshpass\"  // package sshpass")
	fmt.Println("  client, err := sshpass.NewClient(cfg)")
	fmt.Println("  client.Exec(\"ls\") / client.Shell() / client.SFTP()")
	fmt.Println("  sshpass.WithProgress(func(desc string, sent, total int64){ ... })")
}

// printVersion prints version info.
func printVersion() {
	fmt.Printf("win-sshpass version %s (Windows)\n", sshpass.Version)
}
