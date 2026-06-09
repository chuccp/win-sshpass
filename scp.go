package main

import (
	"fmt"
	"strings"
)

// runSCP executes the scp command (file transfer over SSH)
func runSCP(config *Config, args []string) error {
	conn, err := connectSFTP(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	// parse scp arguments to determine source and target
	var remotePath string
	var remoteIdx int
	var remoteCount int
	var nonFlagArgs []string

	// collect non-flag arguments
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// parse source and target
	var localFiles []string
	for i, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// remote path user@host:path (supports IPv6)
			_, _, remotePath = parseUserHostPath(arg)
			remoteIdx = i
			remoteCount++
		} else if arg != "scp" {
			// local file
			localFiles = append(localFiles, arg)
		}
	}

	if remoteCount > 1 {
		return fmt.Errorf("scp does not support multiple remote paths")
	}

	if len(localFiles) == 0 || remotePath == "" {
		return fmt.Errorf("failed to parse scp arguments: %v", args)
	}

	// determine upload or download: remote path in last argument means upload
	isUpload := (remoteIdx == len(nonFlagArgs)-1)

	// clean remote path (handle Git Bash // prefix and path conversion)
	remotePath = cleanRemotePath(remotePath)

	if isUpload {
		for _, lf := range localFiles {
			if err := conn.Upload(lf, remotePath); err != nil {
				return err
			}
		}
		return nil
	}

	for _, lf := range localFiles {
		if err := conn.Download(remotePath, lf); err != nil {
			return err
		}
	}
	return nil
}

// runRsync executes rsync command (file sync over SSH)
func runRsync(config *Config, args []string) error {
	// simple implementation: parse source and target
	var remoteArg string
	var remoteCount int
	var localArgs []string
	var isUpload bool

	for _, arg := range args {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			remoteArg = arg
			remoteCount++
		} else if !strings.HasPrefix(arg, "-") {
			localArgs = append(localArgs, arg)
		}
	}

	if remoteCount > 1 {
		return fmt.Errorf("rsync does not support multiple remote paths")
	}

	if remoteArg == "" || len(localArgs) == 0 {
		return fmt.Errorf("failed to parse rsync arguments")
	}

	// determine direction: if remote is first non-flag arg, it's download
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			isUpload = false
		} else {
			isUpload = true
		}
		break
	}

	conn, err := connectSFTP(config)
	if err != nil {
		return err
	}
	defer conn.Close()

	if isUpload {
		// local to remote (upload)
		_, _, remotePath := parseUserHostPath(remoteArg)
		rPath := cleanRemotePath(remotePath)
		for _, lf := range localArgs {
			if err := conn.Upload(lf, rPath); err != nil {
				return err
			}
		}
		return nil
	}
	// remote to local (download)
	_, _, remotePath := parseUserHostPath(remoteArg)
	rPath := cleanRemotePath(remotePath)
	for _, lf := range localArgs {
		if err := conn.Download(rPath, lf); err != nil {
			return err
		}
	}
	return nil
}
