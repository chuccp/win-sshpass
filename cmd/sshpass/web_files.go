package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	sshpass "github.com/chuccp/win-sshpass"
	"github.com/pkg/sftp"
)

type webFileSystem interface {
	Getwd() (string, error)
	ReadDir(string) ([]os.FileInfo, error)
	Open(string) (io.ReadCloser, error)
	Create(string) (io.WriteCloser, error)
	Stat(string) (os.FileInfo, error)
	MkdirAll(string) error
	Remove(string) error
	RemoveDirectory(string) error
	Close() error
}

type webFileProvider interface {
	FileSystem() (webFileSystem, error)
}

type sshpassWebClient struct {
	client *sshpass.Client
}

func (c *sshpassWebClient) ExecCapture(command string) (string, string, int, error) {
	return c.client.ExecCapture(command)
}

func (c *sshpassWebClient) Close() error {
	return c.client.Close()
}

func (c *sshpassWebClient) FileSystem() (webFileSystem, error) {
	client, err := c.client.SFTP()
	if err != nil {
		return nil, err
	}
	return &sftpWebFileSystem{
		client: client.SFTP(),
		close:  client.Close,
	}, nil
}

type sftpWebFileSystem struct {
	client *sftp.Client
	close  func() error
}

func (f *sftpWebFileSystem) Getwd() (string, error)                  { return f.client.Getwd() }
func (f *sftpWebFileSystem) ReadDir(p string) ([]os.FileInfo, error) { return f.client.ReadDir(p) }
func (f *sftpWebFileSystem) Open(p string) (io.ReadCloser, error)    { return f.client.Open(p) }
func (f *sftpWebFileSystem) Create(p string) (io.WriteCloser, error) { return f.client.Create(p) }
func (f *sftpWebFileSystem) Stat(p string) (os.FileInfo, error)      { return f.client.Stat(p) }
func (f *sftpWebFileSystem) MkdirAll(p string) error                 { return f.client.MkdirAll(p) }
func (f *sftpWebFileSystem) Remove(p string) error                   { return f.client.Remove(p) }
func (f *sftpWebFileSystem) RemoveDirectory(p string) error {
	return f.client.RemoveDirectory(p)
}
func (f *sftpWebFileSystem) Close() error { return f.close() }

type webRemoteEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime int64  `json:"mod_time"`
}

func (a *webApp) fileSystemForRequest(w http.ResponseWriter, r *http.Request) (webFileSystem, *webSession, bool) {
	session := a.sessionFromRequest(r)
	if session == nil {
		webError(w, "not connected", http.StatusUnauthorized)
		return nil, nil, false
	}
	provider, ok := session.client.(webFileProvider)
	if !ok {
		webError(w, "file manager is unavailable for this connection", http.StatusNotImplemented)
		return nil, session, false
	}
	files, err := provider.FileSystem()
	if err != nil {
		webError(w, "cannot open SFTP file manager: "+err.Error(), http.StatusBadGateway)
		return nil, session, false
	}
	return files, session, true
}

func (a *webApp) handleFiles(w http.ResponseWriter, r *http.Request) {
	files, _, ok := a.fileSystemForRequest(w, r)
	if !ok {
		return
	}
	defer files.Close()

	current := strings.TrimSpace(r.URL.Query().Get("path"))
	if current == "" {
		var err error
		current, err = files.Getwd()
		if err != nil || current == "" || current == "." {
			current = "/"
		}
	}
	current = path.Clean(current)
	infos, err := files.ReadDir(current)
	if err != nil {
		webError(w, "cannot list directory: "+err.Error(), http.StatusBadGateway)
		return
	}
	entries := make([]webRemoteEntry, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, webRemoteEntry{
			Name:    info.Name(),
			Path:    path.Join(current, info.Name()),
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Unix(),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	writeWebJSON(w, http.StatusOK, map[string]any{
		"path":    current,
		"parent":  path.Dir(current),
		"entries": entries,
	})
}

func (a *webApp) handleDownload(w http.ResponseWriter, r *http.Request) {
	files, _, ok := a.fileSystemForRequest(w, r)
	if !ok {
		return
	}
	defer files.Close()

	remotePath := strings.TrimSpace(r.URL.Query().Get("path"))
	if remotePath == "" {
		webError(w, "path is required", http.StatusBadRequest)
		return
	}
	info, err := files.Stat(remotePath)
	if err != nil {
		webError(w, "cannot stat remote file: "+err.Error(), http.StatusBadGateway)
		return
	}
	if info.IsDir() {
		webError(w, "cannot download a directory", http.StatusBadRequest)
		return
	}
	file, err := files.Open(remotePath)
	if err != nil {
		webError(w, "cannot open remote file: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer file.Close()

	name := path.Base(remotePath)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file)
}

func (a *webApp) handleUpload(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	files, _, ok := a.fileSystemForRequest(w, r)
	if !ok {
		return
	}
	defer files.Close()

	r.Body = http.MaxBytesReader(w, r.Body, webMaxUploadBytes)
	reader, err := r.MultipartReader()
	if err != nil {
		webError(w, "multipart upload required", http.StatusBadRequest)
		return
	}
	var targetDir string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			webError(w, "cannot read upload: "+err.Error(), http.StatusBadRequest)
			return
		}
		switch part.FormName() {
		case "path":
			data, err := io.ReadAll(io.LimitReader(part, 64<<10))
			if err != nil {
				part.Close()
				webError(w, "cannot read upload path", http.StatusBadRequest)
				return
			}
			targetDir = strings.TrimSpace(string(data))
		case "file":
			if targetDir == "" {
				part.Close()
				webError(w, "upload path must precede file data", http.StatusBadRequest)
				return
			}
			name := safeRemoteName(part.FileName())
			if name == "" {
				part.Close()
				webError(w, "invalid upload filename", http.StatusBadRequest)
				return
			}
			remotePath := path.Join(targetDir, name)
			file, err := files.Create(remotePath)
			if err != nil {
				part.Close()
				webError(w, "cannot create remote file: "+err.Error(), http.StatusBadGateway)
				return
			}
			written, copyErr := io.Copy(file, part)
			closeErr := file.Close()
			part.Close()
			if copyErr != nil || closeErr != nil {
				_ = files.Remove(remotePath)
				if copyErr != nil {
					webError(w, "upload failed: "+copyErr.Error(), http.StatusBadGateway)
				} else {
					webError(w, "upload failed: "+closeErr.Error(), http.StatusBadGateway)
				}
				return
			}
			writeWebJSON(w, http.StatusOK, map[string]any{
				"ok":       true,
				"path":     remotePath,
				"size":     written,
				"filename": name,
			})
			return
		}
		part.Close()
	}
	webError(w, "no file data in upload", http.StatusBadRequest)
}

func safeRemoteName(name string) string {
	name = path.Base(strings.ReplaceAll(name, `\`, "/"))
	if name == "" || name == "." || name == ".." || name == "/" || strings.ContainsAny(name, "\x00\r\n") {
		return ""
	}
	return name
}

func (a *webApp) handleMkdir(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	var request struct {
		Path string `json:"path"`
	}
	if err := decodeWebJSON(w, r, &request); err != nil {
		webError(w, err.Error(), http.StatusBadRequest)
		return
	}
	request.Path = strings.TrimSpace(request.Path)
	if request.Path == "" {
		webError(w, "path is required", http.StatusBadRequest)
		return
	}
	files, _, ok := a.fileSystemForRequest(w, r)
	if !ok {
		return
	}
	defer files.Close()
	if err := files.MkdirAll(path.Clean(request.Path)); err != nil {
		webError(w, "cannot create directory: "+err.Error(), http.StatusBadGateway)
		return
	}
	writeWebJSON(w, http.StatusOK, map[string]any{"ok": true, "path": path.Clean(request.Path)})
}

func (a *webApp) handleDelete(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	var request struct {
		Path string `json:"path"`
	}
	if err := decodeWebJSON(w, r, &request); err != nil {
		webError(w, err.Error(), http.StatusBadRequest)
		return
	}
	request.Path = path.Clean(strings.TrimSpace(request.Path))
	if request.Path == "" || request.Path == "." || request.Path == "/" {
		webError(w, "refusing to delete this path", http.StatusBadRequest)
		return
	}
	files, _, ok := a.fileSystemForRequest(w, r)
	if !ok {
		return
	}
	defer files.Close()
	info, err := files.Stat(request.Path)
	if err != nil {
		webError(w, "cannot stat path: "+err.Error(), http.StatusBadGateway)
		return
	}
	if info.IsDir() {
		err = files.RemoveDirectory(request.Path)
	} else {
		err = files.Remove(request.Path)
	}
	if err != nil {
		webError(w, "cannot delete path: "+err.Error(), http.StatusBadGateway)
		return
	}
	writeWebJSON(w, http.StatusOK, map[string]any{"ok": true, "path": request.Path})
}
