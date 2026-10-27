package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type fakeTerminalClient struct {
	*fakeWebExecutor
	terminal *fakeWebTerminal
}

func (f *fakeTerminalClient) OpenTerminal(output io.Writer, columns, rows int) (webTerminal, error) {
	f.terminal.output = output
	f.terminal.columns = columns
	f.terminal.rows = rows
	_, _ = output.Write([]byte("shell-ready\r\n"))
	return f.terminal, nil
}

type fakeWebTerminal struct {
	mu        sync.Mutex
	output    io.Writer
	input     bytes.Buffer
	columns   int
	rows      int
	resizeCol int
	resizeRow int
	closeOnce sync.Once
	closed    chan struct{}
	closeHits int
}

func newFakeWebTerminal() *fakeWebTerminal {
	return &fakeWebTerminal{closed: make(chan struct{})}
}

func (t *fakeWebTerminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.input.Write(data)
}

func (t *fakeWebTerminal) Resize(columns, rows int) error {
	t.mu.Lock()
	t.resizeCol, t.resizeRow = columns, rows
	t.mu.Unlock()
	return nil
}

func (t *fakeWebTerminal) Wait() error {
	<-t.closed
	return nil
}

func (t *fakeWebTerminal) Close() error {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		t.closeHits++
		t.mu.Unlock()
		close(t.closed)
	})
	return nil
}

func TestWebTerminalWebSocket(t *testing.T) {
	fakeTerminal := newFakeWebTerminal()
	provider := &fakeTerminalClient{
		fakeWebExecutor: &fakeWebExecutor{},
		terminal:        fakeTerminal,
	}
	app, _ := newWebTestApp(t, provider, nil)
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	connectResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(connectResponse, webRequest(http.MethodPost, "/api/connect", map[string]any{
		"host": "example.com",
	}, nil))
	cookies := connectResponse.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("connect cookies = %v", cookies)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/terminal"
	config, err := websocket.NewConfig(wsURL, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	config.Header.Set("Cookie", cookies[0].String())
	conn, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer conn.Close()

	var banner []byte
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err := websocket.Message.Receive(conn, &banner); err != nil {
		t.Fatalf("receive banner: %v", err)
	}
	if string(banner) != "shell-ready\r\n" {
		t.Errorf("banner = %q", banner)
	}
	if err := websocket.JSON.Send(conn, terminalInputMessage{Type: "input", Data: "uname -a\r"}); err != nil {
		t.Fatal(err)
	}
	if err := websocket.JSON.Send(conn, terminalInputMessage{Type: "resize", Columns: 132, Rows: 41}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		fakeTerminal.mu.Lock()
		input := fakeTerminal.input.String()
		resizeCol, resizeRow := fakeTerminal.resizeCol, fakeTerminal.resizeRow
		fakeTerminal.mu.Unlock()
		if input == "uname -a\r" && resizeCol == 132 && resizeRow == 41 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("terminal input/size not received: input=%q size=%dx%d", input, resizeCol, resizeRow)
		}
		time.Sleep(10 * time.Millisecond)
	}

	conn.Close()
	select {
	case <-fakeTerminal.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("terminal was not closed after websocket disconnect")
	}
}
