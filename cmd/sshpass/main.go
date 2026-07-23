package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	sshpass "github.com/chuccp/win-sshpass"
)

// repeatableFlag is a flag.Value that collects multiple occurrences of the
// same flag (e.g. -L a -L b → ["a", "b"]).
type repeatableFlag []string

func (f *repeatableFlag) String() string     { return strings.Join(*f, ",") }
func (f *repeatableFlag) Set(v string) error { *f = append(*f, v); return nil }

// parseForwardSpec parses an SSH forward spec into a listen address and a
// target address. Supported formats (matching OpenSSH):
//
//	port:host:hostport            → 127.0.0.1:port, host:hostport
//	bind:port:host:hostport       → bind:port, host:hostport
func parseForwardSpec(spec string) (listen, target string, err error) {
	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 3: // port:host:hostport
		listen = "127.0.0.1:" + parts[0]
		target = parts[1] + ":" + parts[2]
	case 4: // bind:port:host:hostport
		listen = parts[0] + ":" + parts[1]
		target = parts[2] + ":" + parts[3]
	default:
		return "", "", fmt.Errorf("invalid forward spec %q (expected [bind:]port:host:hostport)", spec)
	}
	return listen, target, nil
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
	timeout := flag.Int("t", 0, "total operation timeout in seconds (0 = no limit)")
	connectTimeout := flag.Int("ct", 10, "TCP connection timeout in seconds")
	retries := flag.Int("retry", 3, "total connection attempts (default 3)")
	resume := flag.Bool("resume", false, "resume interrupted file transfer from breakpoint")
	proxyURL := flag.String("proxy", "", "proxy URL for SSH connection (socks5://[user:pass@]host:port, socks4://, http://, https://)")
	keyAlgo := flag.String("algo", "ed25519", "key algorithm for keygen subcommand (ed25519 or rsa)")
	keyComment := flag.String("comment", "", "comment embedded in the generated public key (default: user@host)")
	keyOutPath := flag.String("out", "", "output path for the generated private key (keygen subcommand; default: ~/.ssh/id_ed25519 or ~/.ssh/id_rsa)")
	showVersion := flag.Bool("v", false, "show version")
	showHelp := flag.Bool("help", false, "show help")
	jsonFlag := flag.Bool("json", false, "output results as JSON (for AI agents and automation)")
	var localForwards, remoteForwards repeatableFlag
	flag.Var(&localForwards, "L", "local port forward [bind:]port:host:hostport (e.g. -L 8080:db:3306)")
	flag.Var(&remoteForwards, "R", "remote port forward [bind:]port:host:hostport (e.g. -R 9090:localhost:8080)")
	flag.Parse()

	// initialize JSON state
	jsonState.enabled = *jsonFlag
	jsonInit()

	// CLI-side UI adapters: progress bar (stderr) and zenity file dialogs.
	// The SDK itself ships no UI; these are injected through options.
	cliOpts := []sshpass.Option{
		sshpass.WithProgress(newCLIProgress(os.Stderr).progress),
		sshpass.WithFileSelector(cliFileSelector{}),
		sshpass.WithSignalHandler(),
	}
	if *resume {
		cliOpts = append(cliOpts, sshpass.WithResume())
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
			fatalError("%v", err)
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
		ProxyURL:       *proxyURL,
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
			jsonSetCommand(fmt.Sprintf("hash %s %s", remainingArgs[1], remainingArgs[2]))
			result, err := sshpass.HashFile(remainingArgs[2], remainingArgs[1])
			if err != nil {
				fatalError("%v", err)
			}
			if jsonEnabled() {
				jsonSuccess(result)
				return
			}
			fmt.Println(result)
			return
		case "verify":
			if len(remainingArgs) != 4 {
				fatalError("Usage: sshpass verify <algorithm> <hash> <file>\nAlgorithms: md5, sha1, sha256, sha512")
			}
			jsonSetCommand(fmt.Sprintf("verify %s %s %s", remainingArgs[1], remainingArgs[2], remainingArgs[3]))
			ok, err := sshpass.VerifyFile(remainingArgs[3], remainingArgs[1], remainingArgs[2])
			if err != nil {
				fatalError("%v", err)
			}
			if jsonEnabled() {
				if ok {
					jsonSuccess("OK")
				} else {
					jsonFail("verification FAILED", 1)
				}
				return
			}
			if ok {
				fmt.Println("OK")
			} else {
				fmt.Println("FAILED")
			}
			return
		}
	}

	// --- keygen subcommand (SSH key generation, local only) ---
	// keygen generates a key pair locally. Deployment to a remote server is
	// NOT automated — users deploy the public key manually.
	if len(remainingArgs) > 0 && remainingArgs[0] == "keygen" {
		runKeygen(remainingArgs[1:], keygenGlobalFlags{
			algo:    *keyAlgo,
			comment: *keyComment,
			outPath: *keyOutPath,
		})
		return
	}

	// detect command type
	cmdType := sshpass.DetectCommandType(remainingArgs)

	// handle based on command type
	switch cmdType {
	case sshpass.CommandSCP:
		if len(localForwards) > 0 || len(remoteForwards) > 0 {
			fatalError("port forwarding (-L/-R) is not supported with scp")
		}
		scpParsed, scpArgs := sshpass.ParseSCPArgs(remainingArgs)
		cfgConfig := sshpass.NewConfig()
		cfgConfig.MergeConfig(config, scpParsed) // config file as src, scp-parsed as override
		cfgConfig.MergeConfig(nil, cliOverride)  // CLI as final override
		cfgConfig.ApplyUserDefault()
		cfgConfig.Normalize()
		jsonSetHost(fmt.Sprintf("%s@%s", cfgConfig.User, cfgConfig.Host))
		client, err := sshpass.NewClient(cfgConfig, cliOpts...)
		if err != nil {
			fatalError("SCP connection failed: %v", err)
		}
		defer client.Close()
		if err := sshpass.RunSCP(client, scpArgs); err != nil {
			fatalError("SCP failed: %v", err)
		}
		if jsonEnabled() {
			jsonSuccess("SCP transfer completed")
		}
		return

	case sshpass.CommandRsync:
		if len(localForwards) > 0 || len(remoteForwards) > 0 {
			fatalError("port forwarding (-L/-R) is not supported with rsync")
		}
		rsyncParsed, rsyncArgs := sshpass.ParseRsyncArgs(remainingArgs)
		cfgConfig := sshpass.NewConfig()
		cfgConfig.MergeConfig(config, rsyncParsed) // config file as src, rsync-parsed as override
		cfgConfig.MergeConfig(nil, cliOverride)    // CLI as final override
		cfgConfig.ApplyUserDefault()
		cfgConfig.Normalize()
		jsonSetHost(fmt.Sprintf("%s@%s", cfgConfig.User, cfgConfig.Host))
		client, err := sshpass.NewClient(cfgConfig, cliOpts...)
		if err != nil {
			fatalError("Rsync connection failed: %v", err)
		}
		defer client.Close()
		if err := sshpass.RunRsync(client, rsyncArgs); err != nil {
			fatalError("Rsync failed: %v", err)
		}
		if jsonEnabled() {
			jsonSuccess("Rsync transfer completed")
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
	jsonSetHost(fmt.Sprintf("%s@%s", config.User, config.Host))

	// validate config
	if err := config.Validate(); err != nil {
		fatalError("%v", err)
	}

	// file transfer path — uses client.SFTP which shares the Client's SSH
	// connection, timeout, and interrupt handling.
	if *localPath != "" && *remotePath != "" {
		if len(localForwards) > 0 || len(remoteForwards) > 0 {
			fatalError("port forwarding (-L/-R) is not supported with file transfer mode")
		}
		localPaths, err := sshpass.SplitPaths(*localPath, "local")
		if err != nil {
			fatalError("%v", err)
		}
		remotePaths, err := sshpass.SplitPaths(*remotePath, "remote")
		if err != nil {
			fatalError("%v", err)
		}
		for i := range remotePaths {
			remotePaths[i], err = sshpass.CleanRemotePath(remotePaths[i])
			if err != nil {
				fatalError("%v", err)
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

		var transferDesc string
		if *download {
			transferDesc = "download"
			for _, rPath := range remotePaths {
				for _, lp := range localPaths {
					if !jsonEnabled() {
						fmt.Printf("Downloading %s -> %s...\n", rPath, lp)
					}
					if err := conn.Download(rPath, lp); err != nil {
						fatalError("Download failed: %v", err)
					}
				}
			}
		} else {
			transferDesc = "upload"
			for _, lp := range localPaths {
				for _, rPath := range remotePaths {
					if !jsonEnabled() {
						fmt.Printf("Uploading %s -> %s...\n", lp, rPath)
					}
					if err := conn.Upload(lp, rPath); err != nil {
						fatalError("Upload failed: %v", err)
					}
				}
			}
		}
		if jsonEnabled() {
			jsonSuccess(fmt.Sprintf("%s completed successfully", transferDesc))
		} else {
			if *download {
				fmt.Println("Download successful!")
			} else {
				fmt.Println("Upload successful!")
			}
		}
		return
	} else if *localPath != "" || *remotePath != "" {
		fatalError("file transfer requires both -local and -remote arguments")
	}

	// command/shell path
	client, err := sshpass.NewClient(config, cliOpts...)
	if err != nil {
		fatalError("SSH connection failed: %v", err)
	}
	defer client.Close()

	// Set up port forwarding if requested.
	var forwarders []*sshpass.Forwarder
	for _, spec := range localForwards {
		listen, target, err := parseForwardSpec(spec)
		if err != nil {
			fatalError("%v", err)
		}
		f, err := client.LocalForward(listen, target)
		if err != nil {
			fatalError("local forward %s failed: %v", spec, err)
		}
		forwarders = append(forwarders, f)
		if !jsonEnabled() {
			fmt.Fprintf(os.Stderr, "local forward: %s -> %s\n", listen, target)
		}
	}
	for _, spec := range remoteForwards {
		listen, target, err := parseForwardSpec(spec)
		if err != nil {
			fatalError("%v", err)
		}
		f, err := client.RemoteForward(listen, target)
		if err != nil {
			fatalError("remote forward %s failed: %v", spec, err)
		}
		forwarders = append(forwarders, f)
		if !jsonEnabled() {
			fmt.Fprintf(os.Stderr, "remote forward: %s -> %s\n", listen, target)
		}
	}
	defer func() {
		for _, f := range forwarders {
			f.Close()
		}
	}()

	// execute command
	cmd := *command
	if cmd == "" {
		cmd = cmdToRun
	}
	jsonSetCommand(cmd)

	// If forwarding is active without a command, block until interrupted.
	if len(forwarders) > 0 && cmd == "" {
		if jsonEnabled() {
			jsonSuccess("Port forwarding established")
		}
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		return
	}

	// JSON mode: capture output and emit structured result
	if jsonEnabled() {
		if cmd == "" {
			jsonFail("JSON mode is not supported for interactive shell sessions", -1)
		}
		stdout, stderr, exitCode, execErr := client.ExecCapture(cmd)
		r := jsonResult{
			Host:     jsonState.host,
			Command:  cmd,
			ExitCode: exitCode,
			Stdout:   stdout,
			Stderr:   stderr,
		}
		if exitCode == 0 && execErr == nil {
			r.Success = true
		} else {
			r.Success = false
			if execErr != nil {
				// session/connection-level error
				r.Error = execErr.Error()
			} else {
				// command exited with non-zero code; stderr is already
				// in r.Stderr — provide a brief summary in error
				r.Error = fmt.Sprintf("command exited with code %d", exitCode)
			}
		}
		printJSON(r)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return
	}

	// Normal mode: stream output directly
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
// In JSON mode, it outputs a JSON error result to stdout instead.
func fatalError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if jsonEnabled() {
		jsonFail(msg, -1)
	}
	fmt.Fprintln(os.Stderr, msg)
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
	fmt.Println("  win-sshpass keygen [-algo <ed25519|rsa>] [-out <keypath>]  (generate key pair locally)")
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
	fmt.Println("  -resume            resume interrupted file transfer from breakpoint")
	fmt.Println("  -proxy <url>       connect via proxy (socks5://[user:pass@]host:port, socks4://, http://, https://)")
	fmt.Println("  -local <path>      local file path(s), comma-separated for multiple files")
	fmt.Println("  -remote <path>     remote file path (for upload/download)")
	fmt.Println("  -d                 download mode (remote to local)")
	fmt.Println("  -algo <type>       key algorithm for keygen (ed25519 or rsa, default: ed25519)")
	fmt.Println("  -comment <string>  comment for generated key (default: user@host)")
	fmt.Println("  -out <path>        output path for generated private key (keygen; default: ~/.ssh/id_ed25519)")
	fmt.Println("  -L [bind:]port:host:hostport  local port forward (e.g. -L 8080:db:3306)")
	fmt.Println("  -R [bind:]port:host:hostport  remote port forward (e.g. -R 9090:localhost:8080)")
	fmt.Println("  -json              output results as JSON (for AI agents and automation)")
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
	fmt.Println("  win-sshpass -p 'pass' -h example.com -local ./bigfile.iso -remote /data/bigfile.iso -resume")
	fmt.Println("  win-sshpass -p 'pass' -L 8080:db.internal:3306 ssh user@jumphost   # access db.internal:3306 via jumphost at localhost:8080")
	fmt.Println("  win-sshpass -p 'pass' -R 9090:localhost:8080 ssh user@server        # expose local :8080 at server:9090")
	fmt.Println("  win-sshpass -p 'pass' -proxy socks5://127.0.0.1:1080 ssh user@example.com")
	fmt.Println("  win-sshpass -p 'pass' -proxy http://user:pass@proxy.local:8080 ssh user@example.com")
	fmt.Println("  win-sshpass keygen                                  # generate ed25519 key to ~/.ssh/id_ed25519")
	fmt.Println("  win-sshpass keygen -algo rsa -out ~/.ssh/mykey      # generate RSA key to custom path")
	fmt.Println("  win-sshpass -json -p 'pass' ssh user@example.com 'whoami'  # JSON output for automation")
	fmt.Println("\nSDK usage (as a Go library):")
	fmt.Println("  import \"github.com/chuccp/win-sshpass\"  // package sshpass")
	fmt.Println("  client, err := sshpass.NewClient(cfg)")
	fmt.Println("  client.Exec(\"ls\") / client.Shell() / client.SFTP()")
	fmt.Println("  sshpass.WithProgress(func(desc string, sent, total int64){ ... })")
}

// printVersion prints version info.
func printVersion() {
	fmt.Printf("win-sshpass version %s (%s/%s)\n", sshpass.Version, runtime.GOOS, runtime.GOARCH)
}
