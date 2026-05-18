package main

import (
	"fmt"
	"strings"

	"github.com/pkg/sftp"
)

// runSCP executes the scp command (file transfer over SSH)
func runSCP(config *Config, args []string) error {
	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// set up operation timeout (timer resets on each data transfer)
	resetTimeout, stopTimer := setupOperationTimeout(func() { client.Close() }, config.Timeout)
	defer stopTimer()

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

	// create a single SFTP client for all transfers
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	if isUpload {
		// upload each local file/directory to the remote path
		for _, lf := range localFiles {
			if err := uploadFile(sftpClient, lf, remotePath, resetTimeout); err != nil {
				return err
			}
		}
		return nil
	}

	// download: downloadFile handles local directory creation internally
	for _, lf := range localFiles {
		if err := downloadFile(sftpClient, remotePath, lf, resetTimeout); err != nil {
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

	// establish SSH connection
	client, err := SSHClient(config)
	if err != nil {
		return err
	}
	defer client.Close()

	// set up operation timeout (timer resets on each data transfer)
	resetTimeout, stopTimer := setupOperationTimeout(func() { client.Close() }, config.Timeout)
	defer stopTimer()

	// create a single SFTP client for all transfers
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	if isUpload {
		// local to remote (upload)
		_, _, remotePath := parseUserHostPath(remoteArg)
		rPath := cleanRemotePath(remotePath)
		for _, lf := range localArgs {
			if err := uploadFile(sftpClient, lf, rPath, resetTimeout); err != nil {
				return err
			}
		}
		return nil
	}
	// remote to local (download)
	_, _, remotePath := parseUserHostPath(remoteArg)
	rPath := cleanRemotePath(remotePath)
	// download: downloadFile handles local directory creation internally
	for _, lf := range localArgs {
		if err := downloadFile(sftpClient, rPath, lf, resetTimeout); err != nil {
			return err
		}
	}
	return nil
}
