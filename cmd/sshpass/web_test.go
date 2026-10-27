package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sshpass "github.com/chuccp/win-sshpass"
)

type fakeWebExecutor struct {
	mu         sync.Mutex
	commands   []string
	closeCount int
	exec       func(string) (string, string, int, error)
}

type fakeFileClient struct {
	*fakeWebExecutor
	files webFileSystem
}

func (f *fakeFileClient) FileSystem() (webFileSystem, error) {
	if f.files == nil {
		return nil, errors.New("no fake file system")
	}
	return f.files, nil
}

type fakeFileInfo struct {
	name    string
	dir     bool
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() os.FileMode  { return f.mode }
func (f fakeFileInfo) ModTime() time.Time { return f.modTime }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }

type fakeWebFileSystem struct {
	cwd        string
	entries    []os.FileInfo
	contents   []byte
	statInfo   os.FileInfo
	created    string
	createdBuf bytes.Buffer
	mkdirs     []string
	removed    []string
	closeCount int
}

func (f *fakeWebFileSystem) Getwd() (string, error) { return f.cwd, nil }
func (f *fakeWebFileSystem) ReadDir(string) ([]os.FileInfo, error) {
	return f.entries, nil
}
func (f *fakeWebFileSystem) Open(string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.contents)), nil
}
func (f *fakeWebFileSystem) Create(p string) (io.WriteCloser, error) {
	f.created = p
	f.createdBuf.Reset()
	return nopWriteCloser{&f.createdBuf}, nil
}
func (f *fakeWebFileSystem) Stat(string) (os.FileInfo, error) {
	if f.statInfo == nil {
		return nil, errors.New("not found")
	}
	return f.statInfo, nil
}
func (f *fakeWebFileSystem) MkdirAll(p string) error {
	f.mkdirs = append(f.mkdirs, p)
	return nil
}
func (f *fakeWebFileSystem) Remove(p string) error {
	f.removed = append(f.removed, p)
	return nil
}
func (f *fakeWebFileSystem) RemoveDirectory(p string) error {
	f.removed = append(f.removed, p)
	return nil
}
func (f *fakeWebFileSystem) Close() error {
	f.closeCount++
	return nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func (f *fakeWebExecutor) ExecCapture(command string) (string, string, int, error) {
	f.mu.Lock()
	f.commands = append(f.commands, command)
	exec := f.exec
	f.mu.Unlock()
	if exec != nil {
		return exec(command)
	}
	return "ok\n", "", 0, nil
}

func (f *fakeWebExecutor) Close() error {
	f.mu.Lock()
	f.closeCount++
	f.mu.Unlock()
	return nil
}

func (f *fakeWebExecutor) closed() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closeCount
}

func newWebTestApp(t *testing.T, client webExecutor, dialErr error) (*webApp, **sshpass.Config) {
	t.Helper()
	var connectedConfig *sshpass.Config
	app := newWebApp(func(config *sshpass.Config) (webExecutor, error) {
		connectedConfig = config
		if dialErr != nil {
			return nil, dialErr
		}
		return client, nil
	})
	return app, &connectedConfig
}

func webRequest(method, path string, body any, cookie *http.Cookie) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	request := httptest.NewRequest(method, path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	return request
}

func connectWebTest(t *testing.T, app *webApp) *http.Cookie {
	t.Helper()
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/connect", map[string]any{
		"host":           "example.com",
		"user":           "alice",
		"port":           "2222",
		"password":       "secret",
		"keyPath":        "/tmp/key",
		"proxyURL":       "socks5://127.0.0.1:1080",
		"timeout":        120,
		"connectTimeout": 5,
		"retries":        2,
		"useAgent":       true,
		"strictHostKey":  true,
	}, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("connect status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("connect cookies = %v, want one session cookie", cookies)
	}
	return cookies[0]
}

func TestWebIndexTemplate(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:6326/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	body := response.Body.String()
	for _, want := range []string{"Connect to a server", "SFTP Files", "SSH Console", "uploadProgress", "127.0.0.1:6326"} {
		if !strings.Contains(body, want) {
			t.Errorf("page does not contain %q", want)
		}
	}
	if got := response.Header().Get("Content-Security-Policy"); got == "" {
		t.Error("Content-Security-Policy header is missing")
	}
	if got := response.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q, want DENY", got)
	}
}

func TestWebStaticAssets(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	for _, name := range []string{"xterm.min.js", "xterm.css", "addon-fit.min.js"} {
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/"+name, nil))
		if response.Code != http.StatusOK {
			t.Errorf("%s status = %d", name, response.Code)
		}
		if response.Body.Len() < 500 {
			t.Errorf("%s size = %d, want embedded asset", name, response.Body.Len())
		}
	}
}

func TestWebConnectAndCookie(t *testing.T) {
	client := &fakeWebExecutor{}
	app, connectedConfig := newWebTestApp(t, client, nil)
	cookie := connectWebTest(t, app)

	if cookie.Name != webCookieName || cookie.Value == "" {
		t.Errorf("cookie = %+v", cookie)
	}
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie security flags = HttpOnly:%v SameSite:%v", cookie.HttpOnly, cookie.SameSite)
	}
	config := *connectedConfig
	if config.Host != "example.com" || config.User != "alice" || config.Port != "2222" {
		t.Errorf("connected config = %+v", config)
	}
	if !config.PasswordSet || config.Password != "secret" || config.KeyPath != "/tmp/key" {
		t.Errorf("auth config = %+v", config)
	}
	if config.Timeout != 120 || config.ConnectTimeout != 5 || config.Retries != 2 {
		t.Errorf("timeout config = %+v", config)
	}
	if !config.UseAgent || !config.StrictHostKey || config.ProxyURL != "socks5://127.0.0.1:1080" {
		t.Errorf("connection options = %+v", config)
	}
}

func TestWebConnectExpandsKeyPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	client := &fakeWebExecutor{}
	app, connectedConfig := newWebTestApp(t, client, nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/connect", map[string]any{
		"host":    "example.com",
		"keyPath": "~/.ssh/id_ed25519",
	}, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	want := filepath.Join(home, ".ssh", "id_ed25519")
	if got := (*connectedConfig).KeyPath; got != want {
		t.Errorf("KeyPath = %q, want %q", got, want)
	}
}

func TestExpandWebPathNonHomePathUnchanged(t *testing.T) {
	path := filepath.Join("relative", "key")
	if got := expandWebPath(path); got != path {
		t.Errorf("expandWebPath(%q) = %q", path, got)
	}
}

func TestWebStatusAndExec(t *testing.T) {
	client := &fakeWebExecutor{
		exec: func(command string) (string, string, int, error) {
			switch command {
			case "uname -a":
				return "Linux\n", "", 0, nil
			case "exit 7":
				return "", "failed\n", 7, nil
			default:
				return "", "", 0, nil
			}
		},
	}
	app, _ := newWebTestApp(t, client, nil)
	cookie := connectWebTest(t, app)

	statusResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(statusResponse, webRequest(http.MethodGet, "/api/status", nil, cookie))
	var status map[string]any
	if err := json.Unmarshal(statusResponse.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status["connected"] != true || status["host"] != "example.com" || status["user"] != "alice" {
		t.Errorf("status = %+v", status)
	}

	execResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(execResponse, webRequest(http.MethodPost, "/api/exec", map[string]any{
		"command": "uname -a",
	}, cookie))
	if execResponse.Code != http.StatusOK {
		t.Fatalf("exec status = %d, body = %s", execResponse.Code, execResponse.Body.String())
	}
	var result struct {
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
	}
	if err := json.Unmarshal(execResponse.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Stdout != "Linux\n" || result.ExitCode != 0 {
		t.Errorf("exec result = %+v", result)
	}

	exitResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(exitResponse, webRequest(http.MethodPost, "/api/exec", map[string]any{
		"command": "exit 7",
	}, cookie))
	if err := json.Unmarshal(exitResponse.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if exitResponse.Code != http.StatusOK || result.ExitCode != 7 || result.Stderr != "failed\n" {
		t.Errorf("non-zero exit response = %d/%+v", exitResponse.Code, result)
	}
}

func TestWebAutoConnectStatus(t *testing.T) {
	client := &fakeWebExecutor{}
	app, connectedConfig := newWebTestApp(t, client, nil)
	app.auto = &sshpass.Config{
		Host:        "auto.example",
		User:        "root",
		Port:        "2222",
		Password:    "secret",
		PasswordSet: true,
	}

	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodGet, "/api/status", nil, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["connected"] != true || result["host"] != "auto.example" || result["user"] != "root" {
		t.Errorf("auto-connect status = %+v", result)
	}
	if *connectedConfig == nil || (*connectedConfig).Host != "auto.example" {
		t.Errorf("auto config = %+v", *connectedConfig)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value == "" {
		t.Fatalf("auto-connect cookies = %+v", cookies)
	}

	execResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(execResponse, webRequest(http.MethodPost, "/api/exec", map[string]any{
		"command": "whoami",
	}, cookies[0]))
	if execResponse.Code != http.StatusOK {
		t.Fatalf("exec after auto-connect = %d, body = %s", execResponse.Code, execResponse.Body.String())
	}
}

func TestWebAutoConnectFailureFallsBackToForm(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, errors.New("authentication failed"))
	app.auto = &sshpass.Config{Host: "auto.example"}

	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodGet, "/api/status", nil, nil))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["connected"] != false || !strings.Contains(result["error"].(string), "authentication failed") {
		t.Errorf("auto-connect failure = %+v", result)
	}
}

func TestWebFileList(t *testing.T) {
	modTime := time.Unix(1_700_000_100, 0)
	files := &fakeWebFileSystem{
		cwd: "/home/root",
		entries: []os.FileInfo{
			fakeFileInfo{name: "zeta.txt", size: 5, mode: 0o644, modTime: modTime},
			fakeFileInfo{name: "src", dir: true, mode: os.ModeDir | 0o755, modTime: modTime},
		},
	}
	provider := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, provider, nil)
	cookie := connectWebTest(t, app)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodGet, "/api/files", nil, cookie))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var result struct {
		Path    string           `json:"path"`
		Entries []webRemoteEntry `json:"entries"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Path != "/home/root" || len(result.Entries) != 2 {
		t.Fatalf("file list = %+v", result)
	}
	if !result.Entries[0].IsDir || result.Entries[0].Name != "src" {
		t.Errorf("directory should sort first: %+v", result.Entries)
	}
	if files.closeCount != 1 {
		t.Errorf("file system close count = %d, want 1", files.closeCount)
	}
}

func TestWebFileDownload(t *testing.T) {
	files := &fakeWebFileSystem{
		contents: []byte("remote content"),
		statInfo: fakeFileInfo{name: "report.txt", size: 14, mode: 0o644},
	}
	provider := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, provider, nil)
	cookie := connectWebTest(t, app)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodGet, "/api/download?path="+url.QueryEscape("/home/root/report.txt"), nil, cookie))
	if response.Code != http.StatusOK || response.Body.String() != "remote content" {
		t.Errorf("download = %d %q", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "report.txt") {
		t.Errorf("Content-Disposition = %q", got)
	}
}

func TestWebFileUpload(t *testing.T) {
	files := &fakeWebFileSystem{}
	provider := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, provider, nil)
	cookie := connectWebTest(t, app)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("path", "/home/root")
	filePart, err := writer.CreateFormFile("file", "upload.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = filePart.Write([]byte("uploaded content"))
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if files.created != "/home/root/upload.txt" || files.createdBuf.String() != "uploaded content" {
		t.Errorf("uploaded path/content = %q/%q", files.created, files.createdBuf.String())
	}
}

func TestWebUploadRejectsTraversalFilename(t *testing.T) {
	files := &fakeWebFileSystem{}
	provider := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, provider, nil)
	cookie := connectWebTest(t, app)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("path", "/home/root")
	filePart, err := writer.CreateFormFile("file", "..")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = filePart.Write([]byte("escape"))
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", response.Code)
	}
	if files.created != "" {
		t.Errorf("unsafe filename created remote path %q", files.created)
	}
}

func TestWebFileMkdirAndDelete(t *testing.T) {
	files := &fakeWebFileSystem{
		statInfo: fakeFileInfo{name: "old", dir: true, mode: os.ModeDir | 0o755},
	}
	provider := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, provider, nil)
	cookie := connectWebTest(t, app)

	mkdirResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(mkdirResponse, webRequest(http.MethodPost, "/api/mkdir", map[string]any{
		"path": "/home/root/new-dir",
	}, cookie))
	if mkdirResponse.Code != http.StatusOK || len(files.mkdirs) != 1 || files.mkdirs[0] != "/home/root/new-dir" {
		t.Errorf("mkdir = %d %+v", mkdirResponse.Code, files.mkdirs)
	}

	deleteResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteResponse, webRequest(http.MethodPost, "/api/delete", map[string]any{
		"path": "/home/root/old",
	}, cookie))
	if deleteResponse.Code != http.StatusOK || len(files.removed) != 1 || files.removed[0] != "/home/root/old" {
		t.Errorf("delete = %d %+v", deleteResponse.Code, files.removed)
	}
}

func TestWebExecRequiresSession(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/exec", map[string]any{
		"command": "whoami",
	}, nil))
	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", response.Code)
	}
}

func TestPrepareWebConfig(t *testing.T) {
	t.Run("CLI connection settings", func(t *testing.T) {
		override := &sshpass.Config{
			Host:           "127.0.0.1",
			User:           "root",
			Port:           "22260",
			Password:       "testpass",
			PasswordSet:    true,
			ConnectTimeout: 5,
			Retries:        1,
		}
		config, err := prepareWebConfig(nil, override, false)
		if err != nil {
			t.Fatal(err)
		}
		if config == nil || config.Host != "127.0.0.1" || config.Port != "22260" {
			t.Fatalf("config = %+v", config)
		}
		if config.User != "root" || !config.HasPassword() {
			t.Errorf("auth config = %+v", config)
		}
	})

	t.Run("missing host is rejected", func(t *testing.T) {
		config, err := prepareWebConfig(nil, &sshpass.Config{Password: "secret", PasswordSet: true}, false)
		if err == nil || !strings.Contains(err.Error(), "SSH host is required") {
			t.Fatalf("error = %v, want required host error", err)
		}
		if config != nil {
			t.Errorf("config = %+v, want nil on validation failure", config)
		}
	})

	t.Run("agent fallback and forwarding", func(t *testing.T) {
		config, err := prepareWebConfig(nil, &sshpass.Config{Host: "example.com"}, true)
		if err != nil {
			t.Fatal(err)
		}
		if config == nil || !config.UseAgent || !config.AgentForward {
			t.Errorf("config = %+v", config)
		}
	})
}

func TestWebConnectErrors(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		client := &fakeWebExecutor{}
		app, connectedConfig := newWebTestApp(t, client, nil)
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/connect", map[string]any{
			"user": "root",
		}, nil))
		if response.Code != http.StatusBadRequest {
			t.Errorf("status = %d, body = %s", response.Code, response.Body.String())
		}
		if *connectedConfig != nil {
			t.Error("dialer was called for invalid config")
		}
	})

	t.Run("dial failure", func(t *testing.T) {
		client := &fakeWebExecutor{}
		app, _ := newWebTestApp(t, client, errors.New("connection refused"))
		response := httptest.NewRecorder()
		app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/connect", map[string]any{
			"host": "example.com",
		}, nil))
		if response.Code != http.StatusBadGateway {
			t.Errorf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})
}

func TestWebExecConnectionFailureCleansSession(t *testing.T) {
	client := &fakeWebExecutor{
		exec: func(string) (string, string, int, error) {
			return "", "", -1, errors.New("connection lost")
		},
	}
	app, _ := newWebTestApp(t, client, nil)
	cookie := connectWebTest(t, app)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/exec", map[string]any{
		"command": "whoami",
	}, cookie))
	if response.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", response.Code)
	}
	if client.closed() != 1 {
		t.Errorf("client close count = %d, want 1", client.closed())
	}
}

func TestWebDisconnectClearsSession(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	cookie := connectWebTest(t, app)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/api/disconnect", map[string]any{}, cookie))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if client.closed() != 1 {
		t.Errorf("client close count = %d, want 1", client.closed())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Errorf("disconnect cookies = %+v, want expired cookie", cookies)
	}
}

func TestWebCrossOriginRejected(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	request := webRequest(http.MethodPost, "/api/connect", map[string]any{
		"host": "example.com",
	}, nil)
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", response.Code)
	}
}

func TestWebSessionExpires(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	now := time.Unix(1_700_000_000, 0)
	app.now = func() time.Time { return now }
	cookie := connectWebTest(t, app)
	now = now.Add(webSessionTTL + time.Second)

	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodGet, "/api/status", nil, cookie))
	var result map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["connected"] != false {
		t.Errorf("status after expiry = %+v", result)
	}
	if client.closed() != 1 {
		t.Errorf("expired client close count = %d, want 1", client.closed())
	}
}

func TestWebJSONContentTypeRequired(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/connect", strings.NewReader(`{"host":"example.com"}`))
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", response.Code)
	}
}
