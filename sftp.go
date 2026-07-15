package sshpass

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

// SFTPClient bundles an SFTP sub-client with the operation-timeout reset hook
// and progress callback so callers can perform uploads/downloads over a
// shared SSH connection managed by a Client.
type SFTPClient struct {
	sftpClient   *sftp.Client
	resetTimeout func()
	progress     ProgressFunc
	resume       bool
}

// Close tears down the SFTP sub-client. It does not close the underlying SSH
// connection; the owning Client is responsible for that.
func (c *SFTPClient) Close() error {
	return c.sftpClient.Close()
}

// SFTP returns the underlying *sftp.Client for advanced operations not covered
// by Upload/Download.
func (c *SFTPClient) SFTP() *sftp.Client {
	return c.sftpClient
}

// Upload uploads a local file or directory to the remote path.
func (c *SFTPClient) Upload(localPath, remotePath string) error {
	return uploadFile(c.sftpClient, localPath, remotePath, c.resetTimeout, c.progress, c.resume)
}

// Download downloads a remote file or directory to the local path.
func (c *SFTPClient) Download(remotePath, localPath string) error {
	return downloadFile(c.sftpClient, remotePath, localPath, c.resetTimeout, c.progress, c.resume)
}

// timeoutReader wraps an io.Reader and calls reset on each successful Read.
// This keeps the operation timeout alive as long as data flows from the remote.
type timeoutReader struct {
	r     io.Reader
	reset func()
}

func (tr *timeoutReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 && tr.reset != nil {
		tr.reset()
	}
	return n, err
}

// timeoutWriter wraps an io.Writer and calls reset on each successful Write.
// This keeps the operation timeout alive as long as data flows to the remote.
type timeoutWriter struct {
	w     io.Writer
	reset func()
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	n, err := tw.w.Write(p)
	if n > 0 && tw.reset != nil {
		tw.reset()
	}
	return n, err
}

// progressWriter wraps an io.Writer, counts bytes written, and invokes the
// progress callback (if set) on every chunk.
type progressWriter struct {
	w     io.Writer
	desc  string
	sent  int64
	total int64
	fn    ProgressFunc
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	if n > 0 {
		pw.sent += int64(n)
		if pw.fn != nil {
			pw.fn(pw.desc, pw.sent, pw.total)
		}
	}
	return n, err
}

// progressReader wraps an io.Reader, counts bytes read, and invokes the
// progress callback (if set) on every chunk.
type progressReader struct {
	r     io.Reader
	desc  string
	sent  int64
	total int64
	fn    ProgressFunc
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.sent += int64(n)
		if pr.fn != nil {
			pr.fn(pr.desc, pr.sent, pr.total)
		}
	}
	return n, err
}

// uploadFile uploads a file or directory to the remote server.
// resetTimeout is called on each data transfer to keep the operation timeout alive.
func uploadFile(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// check if local path is a file or directory
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to get local file info: %w", err)
	}

	if localInfo.IsDir() {
		return uploadDirectory(sftpClient, localPath, remotePath, resetTimeout, progress, resume)
	}
	return uploadSingleFile(sftpClient, localPath, remotePath, resetTimeout, progress, resume)
}

// uploadSingleFile uploads a single file
func uploadSingleFile(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// check if remote path is a directory
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err == nil && remoteFileInfo.IsDir() {
		remotePath = joinRemotePath(remotePath, localBaseName(localPath))
	} else if err != nil && strings.HasSuffix(remotePath, "/") {
		// path doesn't exist but has trailing slash: treat as directory target
		remotePath = joinRemotePath(remotePath, localBaseName(localPath))
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// get file size
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()
	desc := fmt.Sprintf("Uploading %s", localBaseName(localPath))

	// ensure remote directory exists
	remoteDir := remoteDirName(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// determine transfer mode: resume or fresh start
	var offset int64
	var remoteFile *sftp.File

	if resume {
		if info, statErr := sftpClient.Stat(remotePath); statErr == nil {
			remoteSize := info.Size()
			if remoteSize >= fileSize {
				// already fully uploaded — nothing to do
				if progress != nil {
					progress(desc, fileSize, fileSize)
				}
				return nil
			}
			// partial file exists — seek local and append to remote
			offset = remoteSize
			if _, seekErr := localFile.Seek(offset, io.SeekStart); seekErr != nil {
				return fmt.Errorf("failed to seek local file for resume: %w", seekErr)
			}
			remoteFile, err = sftpClient.OpenFile(remotePath, os.O_WRONLY|os.O_APPEND)
			if err != nil {
				return fmt.Errorf("failed to open remote file for append: %w", err)
			}
		} else {
			// remote file does not exist — create fresh
			remoteFile, err = sftpClient.Create(remotePath)
			if err != nil {
				return fmt.Errorf("failed to create remote file: %w", err)
			}
		}
	} else {
		remoteFile, err = sftpClient.Create(remotePath)
		if err != nil {
			return fmt.Errorf("failed to create remote file: %w", err)
		}
	}
	defer remoteFile.Close()

	// build the write pipeline: progress counter -> timeout reset -> remote file
	var w io.Writer = remoteFile
	if progress != nil {
		progress(desc, offset, fileSize)
		w = &progressWriter{w: remoteFile, desc: desc, sent: offset, total: fileSize, fn: progress}
	}
	remoteWriter := &timeoutWriter{w: w, reset: resetTimeout}
	if _, err := io.Copy(remoteWriter, localFile); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// uploadDirectory uploads an entire directory
func uploadDirectory(sftpClient *sftp.Client, localPath, remotePath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// get local directory base name
	localBase := localBaseName(localPath)

	// ensure remote directory exists
	remoteDir := joinRemotePath(remotePath, localBase)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// walk local directory
	return filepath.Walk(localPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// calculate relative path
		relPath, err := filepath.Rel(localPath, filePath)
		if err != nil {
			return err
		}

		// convert Windows relative path to Unix style
		relPath = toSlash(relPath)

		// remote full path
		remoteFullPath := joinRemotePath(remoteDir, relPath)

		if info.IsDir() {
			// create remote directory
			return sftpClient.MkdirAll(remoteFullPath)
		}

		// upload file
		return uploadSingleFile(sftpClient, filePath, remoteFullPath, resetTimeout, progress, resume)
	})
}

// downloadFile downloads a file or directory from the remote server.
// resetTimeout is called on each data transfer to keep the operation timeout alive.
func downloadFile(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// check if remote path is a file or directory
	remoteInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}

	if remoteInfo.IsDir() {
		return downloadDirectory(sftpClient, remotePath, localPath, resetTimeout, progress, resume)
	}
	return downloadSingleFile(sftpClient, remotePath, localPath, resetTimeout, progress, resume)
}

// downloadSingleFile downloads a single file
func downloadSingleFile(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// check if local path is a directory
	localFileInfo, err := os.Stat(localPath)
	if err == nil && localFileInfo.IsDir() {
		localPath = joinLocalPath(localPath, remoteBaseName(remotePath))
	}

	// ensure local directory exists
	localDir := localDirName(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// get remote file size
	fileInfo, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()
	desc := fmt.Sprintf("Downloading %s", remoteBaseName(remotePath))

	// determine transfer mode: resume or fresh start
	var offset int64
	var localFile *os.File

	if resume {
		if info, statErr := os.Stat(localPath); statErr == nil {
			localSize := info.Size()
			if localSize >= fileSize {
				// already fully downloaded — nothing to do
				if progress != nil {
					progress(desc, fileSize, fileSize)
				}
				return nil
			}
			// partial file exists — seek remote and append to local
			offset = localSize
			if _, seekErr := remoteFile.Seek(offset, io.SeekStart); seekErr != nil {
				return fmt.Errorf("failed to seek remote file for resume: %w", seekErr)
			}
			localFile, err = os.OpenFile(localPath, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("failed to open local file for append: %w", err)
			}
		} else {
			// local file does not exist — create fresh
			localFile, err = os.Create(localPath)
			if err != nil {
				return fmt.Errorf("failed to create local file: %w", err)
			}
		}
	} else {
		localFile, err = os.Create(localPath)
		if err != nil {
			return fmt.Errorf("failed to create local file: %w", err)
		}
	}
	defer localFile.Close()

	// build the read pipeline: remote file -> progress counter -> timeout reset
	var r io.Reader = remoteFile
	if progress != nil {
		progress(desc, offset, fileSize)
		r = &progressReader{r: remoteFile, desc: desc, sent: offset, total: fileSize, fn: progress}
	}
	remoteReader := &timeoutReader{r: r, reset: resetTimeout}
	if _, err := io.Copy(localFile, remoteReader); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// flush data to disk for reliability
	if err := localFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// downloadDirectory downloads an entire directory
func downloadDirectory(sftpClient *sftp.Client, remotePath, localPath string, resetTimeout func(), progress ProgressFunc, resume bool) error {
	// get remote directory name (trim trailing / to avoid remoteBaseName returning empty string)
	remoteBase := remoteBaseName(strings.TrimSuffix(remotePath, "/"))

	// ensure local directory exists
	localDir := joinLocalPath(localPath, remoteBase)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// ensure remote path ends with / for relative path calculation
	remotePath = strings.TrimSuffix(remotePath, "/") + "/"

	// walk remote directory
	walker := sftpClient.Walk(remotePath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		remoteFilePath := walker.Path()

		// calculate relative path (remove remote base path)
		relPath := strings.TrimPrefix(remoteFilePath, remotePath)
		if relPath == "" {
			continue
		}

		// local full path
		localFullPath := joinLocalPath(localDir, relPath)

		info := walker.Stat()
		if info.IsDir() {
			// create local directory
			if err := os.MkdirAll(localFullPath, 0755); err != nil {
				return err
			}
		} else {
			// download file
			if err := downloadSingleFile(sftpClient, remoteFilePath, localFullPath, resetTimeout, progress, resume); err != nil {
				return err
			}
		}
	}

	return nil
}
