package main

import (
	"fmt"
	"os"
	"strings"
)

// runSCP executes the scp command (file transfer over SSH)
func runSCP(config *Config, args []string) error {
	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// parse scp arguments to determine source and target
	var remotePath string
	var isUpload bool
	var nonFlagArgs []string

	// collect non-flag arguments
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	// parse source and target
	var localFiles []string
	for _, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			// remote path user@host:path (supports IPv6)
			_, _, remotePath = parseUserHostPath(arg)
		} else if arg != "scp" {
			// local file
			localFiles = append(localFiles, arg)
		}
	}

	// determine upload or download: remote path in last argument means upload
	for i, arg := range nonFlagArgs {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			isUpload = (i == len(nonFlagArgs)-1)
			break
		}
	}

	if len(localFiles) == 0 || remotePath == "" {
		return fmt.Errorf("failed to parse scp arguments: %v", args)
	}

	// clean remote path (handle Git Bash // prefix and path conversion)
	remotePath = cleanRemotePath(remotePath)

	if isUpload {
		// ensure remote target directory exists
		if err := ensureRemoteDir(client, remotePath); err != nil {
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
		// upload each local file/directory to the remote path
		for _, lf := range localFiles {
			if err := uploadFile(client, lf, remotePath); err != nil {
				return err
			}
		}
		return nil
	}

	// download: ensure local target directory exists
	if len(localFiles) > 1 {
		if err := os.MkdirAll(localFiles[0], 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}
	for _, lf := range localFiles {
		if err := downloadFile(client, remotePath, lf); err != nil {
			return err
		}
	}
	return nil
}

// runRsync executes rsync command (file sync over SSH)
func runRsync(config *Config, args []string) error {
	// simple implementation: parse source and target
	var remoteArg string
	var localArgs []string
	var isUpload bool

	for _, arg := range args {
		if strings.Contains(arg, "@") && strings.Contains(arg, ":") {
			remoteArg = arg
		} else if !strings.HasPrefix(arg, "-") {
			localArgs = append(localArgs, arg)
		}
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

	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	if isUpload {
		// local to remote (upload)
		_, _, remotePath := parseUserHostPath(remoteArg)
		rPath := cleanRemotePath(remotePath)
		// ensure remote target directory exists
		if err := ensureRemoteDir(client, rPath); err != nil {
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
		for _, lf := range localArgs {
			if err := uploadFile(client, lf, rPath); err != nil {
				return err
			}
		}
		return nil
	}
	// remote to local (download)
	_, _, remotePath := parseUserHostPath(remoteArg)
	rPath := cleanRemotePath(remotePath)
	// ensure local target directory exists
	if len(localArgs) > 1 {
		if err := os.MkdirAll(localArgs[0], 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}
	for _, lf := range localArgs {
		if err := downloadFile(client, rPath, lf); err != nil {
			return err
		}
	}
	return nil
}
