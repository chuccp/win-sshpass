package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

const (
	defaultTerminalColumns = 100
	defaultTerminalRows    = 30
	maxTerminalInputBytes  = 64 << 10
)

type webTerminal interface {
	Write([]byte) (int, error)
	Resize(columns, rows int) error
	Wait() error
	Close() error
}

type webTerminalProvider interface {
	OpenTerminal(output io.Writer, columns, rows int) (webTerminal, error)
}

func (c *sshpassWebClient) OpenTerminal(output io.Writer, columns, rows int) (webTerminal, error) {
	session, err := c.client.SSHClient().NewSession()
	if err != nil {
		return nil, err
	}
	if columns < 2 || columns > 1000 {
		columns = defaultTerminalColumns
	}
	if rows < 2 || rows > 500 {
		rows = defaultTerminalRows
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, columns, modes); err != nil {
		session.Close()
		return nil, err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, err
	}
	session.Stdout = output
	session.Stderr = output
	if err := session.Shell(); err != nil {
		stdin.Close()
		session.Close()
		return nil, err
	}
	return &sshWebTerminal{session: session, stdin: stdin}, nil
}

type sshWebTerminal struct {
	session   *ssh.Session
	stdin     io.WriteCloser
	closeOnce sync.Once
	closeErr  error
}

func (t *sshWebTerminal) Write(data []byte) (int, error) {
	return t.stdin.Write(data)
}

func (t *sshWebTerminal) Resize(columns, rows int) error {
	if columns < 2 || columns > 1000 || rows < 2 || rows > 500 {
		return fmt.Errorf("invalid terminal size %dx%d", columns, rows)
	}
	return t.session.WindowChange(rows, columns)
}

func (t *sshWebTerminal) Wait() error {
	return t.session.Wait()
}

func (t *sshWebTerminal) Close() error {
	t.closeOnce.Do(func() {
		t.stdin.Close()
		t.closeErr = t.session.Close()
	})
	return t.closeErr
}

type webSocketWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *webSocketWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.Write(data)
}

type terminalInputMessage struct {
	Type    string `json:"type"`
	Data    string `json:"data"`
	Columns int    `json:"cols"`
	Rows    int    `json:"rows"`
}

func (a *webApp) handleTerminal(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	session := a.sessionFromRequest(r)
	if session == nil {
		http.Error(w, "not connected", http.StatusUnauthorized)
		return
	}
	provider, ok := session.client.(webTerminalProvider)
	if !ok {
		http.Error(w, "interactive terminal is unavailable for this connection", http.StatusNotImplemented)
		return
	}

	websocket.Handler(func(conn *websocket.Conn) {
		writer := &webSocketWriter{conn: conn}
		terminal, err := provider.OpenTerminal(writer, defaultTerminalColumns, defaultTerminalRows)
		if err != nil {
			_, _ = writer.Write([]byte(fmt.Sprintf("\r\n[terminal error] %v\r\n", err)))
			return
		}
		defer terminal.Close()

		go func() {
			_ = terminal.Wait()
			_, _ = writer.Write([]byte("\r\n[SSH session closed]\r\n"))
			conn.Close()
		}()

		for {
			var message terminalInputMessage
			if err := websocket.JSON.Receive(conn, &message); err != nil {
				return
			}
			switch message.Type {
			case "input":
				if len(message.Data) > maxTerminalInputBytes {
					return
				}
				if _, err := terminal.Write([]byte(message.Data)); err != nil {
					return
				}
			case "resize":
				_ = terminal.Resize(message.Columns, message.Rows)
			}
		}
	}).ServeHTTP(w, r)
}
