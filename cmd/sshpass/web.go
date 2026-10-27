package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	sshpass "github.com/chuccp/win-sshpass"
)

const (
	webCookieName       = "win_sshpass_web"
	webSessionTTL       = 30 * time.Minute
	webSessionCookieAge = 12 * time.Hour
	webMaxRequestBytes  = 1 << 20
	webMaxUploadBytes   = 1 << 30
)

type webExecutor interface {
	ExecCapture(command string) (stdout, stderr string, exitCode int, err error)
	Close() error
}

type webDialFunc func(*sshpass.Config) (webExecutor, error)

func dialWebClient(config *sshpass.Config) (webExecutor, error) {
	client, err := sshpass.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &sshpassWebClient{client: client}, nil
}

type webSession struct {
	client   webExecutor
	host     string
	user     string
	lastUsed time.Time
	mu       sync.Mutex
}

type webApp struct {
	templates *webTemplates
	dial      webDialFunc
	now       func() time.Time
	auto      *sshpass.Config

	sessionsMu sync.Mutex
	sessions   map[string]*webSession
}

func newWebApp(dial webDialFunc) *webApp {
	if dial == nil {
		dial = dialWebClient
	}
	return &webApp{
		templates: mustLoadWebTemplates(),
		dial:      dial,
		now:       time.Now,
		sessions:  make(map[string]*webSession),
	}
}

func (a *webApp) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /api/status", a.handleStatus)
	mux.HandleFunc("POST /api/connect", a.handleConnect)
	mux.HandleFunc("POST /api/exec", a.handleExec)
	mux.HandleFunc("POST /api/disconnect", a.handleDisconnect)
	mux.HandleFunc("GET /api/files", a.handleFiles)
	mux.HandleFunc("GET /api/download", a.handleDownload)
	mux.HandleFunc("POST /api/upload", a.handleUpload)
	mux.HandleFunc("POST /api/mkdir", a.handleMkdir)
	mux.HandleFunc("POST /api/delete", a.handleDelete)
	mux.HandleFunc("GET /api/terminal", a.handleTerminal)
	mux.HandleFunc("POST /mcp", a.handleMCP)
	mux.HandleFunc("DELETE /mcp", a.handleMCPDelete)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(webAssetFS)))
	return a.securityHeaders(mux)
}

func (a *webApp) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:; img-src 'self' data:; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

type webIndexData struct {
	Address string
}

func (a *webApp) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := a.templates.execute(w, webIndexData{Address: r.Host}); err != nil {
		http.Error(w, "failed to render page", http.StatusInternalServerError)
	}
}

type webConnectRequest struct {
	Host           string `json:"host"`
	User           string `json:"user"`
	Port           string `json:"port"`
	Password       string `json:"password"`
	KeyPath        string `json:"keyPath"`
	ProxyURL       string `json:"proxyURL"`
	StrictHostKey  bool   `json:"strictHostKey"`
	UseAgent       bool   `json:"useAgent"`
	Timeout        int    `json:"timeout"`
	ConnectTimeout int    `json:"connectTimeout"`
	Retries        int    `json:"retries"`
}

func (r webConnectRequest) config() (*sshpass.Config, error) {
	config := sshpass.NewConfig()
	config.Host = strings.TrimSpace(r.Host)
	if strings.TrimSpace(r.User) != "" {
		config.User = strings.TrimSpace(r.User)
	}
	if strings.TrimSpace(r.Port) != "" {
		config.Port = strings.TrimSpace(r.Port)
	}
	config.Password = r.Password
	config.PasswordSet = true
	config.KeyPath = expandWebPath(strings.TrimSpace(r.KeyPath))
	config.ProxyURL = strings.TrimSpace(r.ProxyURL)
	config.StrictHostKey = r.StrictHostKey
	config.UseAgent = r.UseAgent
	config.Timeout = max(r.Timeout, 0)
	config.ConnectTimeout = max(r.ConnectTimeout, 0)
	config.Retries = max(r.Retries, 0)
	config.Normalize()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func expandWebPath(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

func (a *webApp) handleConnect(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	var request webConnectRequest
	if err := decodeWebJSON(w, r, &request); err != nil {
		webError(w, err.Error(), http.StatusBadRequest)
		return
	}
	config, err := request.config()
	if err != nil {
		webError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := a.createSession(w, r, config); err != nil {
		webError(w, "SSH connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	writeWebJSON(w, http.StatusOK, map[string]any{
		"ok":   true,
		"host": config.Host,
		"user": config.User,
	})
}

func (a *webApp) handleStatus(w http.ResponseWriter, r *http.Request) {
	session := a.sessionFromRequest(r)
	if session == nil && a.auto != nil {
		var err error
		session, err = a.createSession(w, r, a.auto)
		if err != nil {
			writeWebJSON(w, http.StatusOK, map[string]any{
				"connected": false,
				"error":     "automatic SSH connection failed: " + err.Error(),
			})
			return
		}
	}
	if session == nil {
		writeWebJSON(w, http.StatusOK, map[string]any{"connected": false})
		return
	}
	writeWebJSON(w, http.StatusOK, map[string]any{
		"connected": true,
		"host":      session.host,
		"user":      session.user,
	})
}

func (a *webApp) createSession(w http.ResponseWriter, r *http.Request, config *sshpass.Config) (*webSession, error) {
	sessionID, session, err := a.newSession(config)
	if err != nil {
		return nil, err
	}
	if oldID := webSessionCookie(r); oldID != "" {
		if old := a.removeSession(oldID); old != nil {
			old.client.Close()
		}
	}
	a.sessionsMu.Lock()
	a.sessions[sessionID] = session
	a.sessionsMu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     webCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(webSessionCookieAge.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	return session, nil
}

func (a *webApp) newSession(config *sshpass.Config) (string, *webSession, error) {
	client, err := a.dial(config)
	if err != nil {
		return "", nil, err
	}
	sessionID, err := newWebSessionID()
	if err != nil {
		client.Close()
		return "", nil, errors.New("failed to create web session")
	}
	session := &webSession{
		client:   client,
		host:     config.Host,
		user:     config.User,
		lastUsed: a.now(),
	}
	return sessionID, session, nil
}

func (a *webApp) sessionByID(id string) *webSession {
	if id == "" {
		return nil
	}
	now := a.now()
	a.sessionsMu.Lock()
	session := a.sessions[id]
	var expired *webSession
	if session != nil && now.Sub(session.lastUsed) > webSessionTTL {
		delete(a.sessions, id)
		expired = session
		session = nil
	}
	if session != nil {
		session.lastUsed = now
	}
	a.sessionsMu.Unlock()
	if expired != nil {
		expired.client.Close()
	}
	return session
}

func (a *webApp) handleExec(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	session := a.sessionFromRequest(r)
	if session == nil {
		webError(w, "not connected", http.StatusUnauthorized)
		return
	}
	var request struct {
		Command string `json:"command"`
	}
	if err := decodeWebJSON(w, r, &request); err != nil {
		webError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.Command) == "" {
		webError(w, "command is required", http.StatusBadRequest)
		return
	}

	session.mu.Lock()
	stdout, stderr, exitCode, execErr := session.client.ExecCapture(request.Command)
	session.mu.Unlock()
	if execErr != nil {
		if id := webSessionCookie(r); id != "" {
			if removed := a.removeSession(id); removed != nil {
				removed.client.Close()
			}
		}
		webError(w, "command failed: "+execErr.Error(), http.StatusBadGateway)
		return
	}
	writeWebJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"stdout":    stdout,
		"stderr":    stderr,
		"exit_code": exitCode,
	})
}

func (a *webApp) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	if id := webSessionCookie(r); id != "" {
		if session := a.removeSession(id); session != nil {
			session.client.Close()
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     webCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	writeWebJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *webApp) sessionFromRequest(r *http.Request) *webSession {
	return a.sessionByID(webSessionCookie(r))
}

func (a *webApp) removeSession(id string) *webSession {
	a.sessionsMu.Lock()
	session := a.sessions[id]
	delete(a.sessions, id)
	a.sessionsMu.Unlock()
	return session
}

func (a *webApp) closeSessions() {
	a.sessionsMu.Lock()
	sessions := a.sessions
	a.sessions = make(map[string]*webSession)
	a.sessionsMu.Unlock()
	for _, session := range sessions {
		session.client.Close()
	}
}

func webSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie(webCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func newWebSessionID() (string, error) {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(data[:]), nil
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return parsed.Host == r.Host
}

func decodeWebJSON(w http.ResponseWriter, r *http.Request, target any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return fmt.Errorf("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, webMaxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON request: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON object")
	}
	return nil
}

func webError(w http.ResponseWriter, message string, status int) {
	writeWebJSON(w, status, map[string]any{
		"ok":    false,
		"error": message,
	})
}

func writeWebJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func prepareWebConfig(base, override *sshpass.Config, agentForward bool) (*sshpass.Config, error) {
	config := sshpass.NewConfig()
	config.MergeConfig(base, override)
	if strings.TrimSpace(config.Host) == "" {
		return nil, errors.New("SSH host is required; use -h <host> or -f <config-file> with -web")
	}
	config.ApplyUserDefault()
	config.Normalize()
	if !config.HasPassword() && config.KeyPath == "" {
		config.UseAgent = true
	}
	if agentForward {
		config.AgentForward = true
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func runWebServer(address string, autoConfig *sshpass.Config) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", address, err)
	}
	app := newWebApp(dialWebClient)
	app.auto = autoConfig
	defer app.closeSessions()

	server := &http.Server{
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	fmt.Fprintf(os.Stderr, "win-sshpass Web UI: http://%s\n", listener.Addr())

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("web server shutdown failed: %w", err)
		}
		return nil
	}
}
