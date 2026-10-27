package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	sshpass "github.com/chuccp/win-sshpass"
)

func mcpRequestFor(method string, id int, params any) map[string]any {
	request := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if id >= 0 {
		request["id"] = id
	}
	if params != nil {
		request["params"] = params
	}
	return request
}

func decodeMCPResponse(t *testing.T, response *httptest.ResponseRecorder) mcpResponse {
	t.Helper()
	var result mcpResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid MCP response %q: %v", response.Body.String(), err)
	}
	return result
}

func TestMCPLifecycleAndTools(t *testing.T) {
	files := &fakeWebFileSystem{
		cwd: "/home/root",
		entries: []os.FileInfo{
			fakeFileInfo{name: "README.md", size: 12, mode: 0o644},
		},
	}
	client := &fakeFileClient{
		fakeWebExecutor: &fakeWebExecutor{
			exec: func(command string) (string, string, int, error) {
				return "mcp:" + command + "\n", "", 0, nil
			},
		},
		files: files,
	}
	app, _ := newWebTestApp(t, client, nil)
	app.auto = &sshpass.Config{Host: "example.com", User: "root", Port: "22", Password: "secret", PasswordSet: true}

	initResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(initResponse, webRequest(http.MethodPost, "/mcp", mcpRequestFor("initialize", 1, map[string]any{
		"protocolVersion": "2025-03-26",
	}), nil))
	if initResponse.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, body = %s", initResponse.Code, initResponse.Body.String())
	}
	sessionID := initResponse.Header().Get(mcpSessionHeader)
	if sessionID == "" {
		t.Fatal("initialize did not return an MCP session header")
	}
	initResult := decodeMCPResponse(t, initResponse)
	if initResult.Error != nil || initResult.Result == nil {
		t.Fatalf("initialize result = %+v", initResult)
	}

	notification := webRequest(http.MethodPost, "/mcp", mcpRequestFor("notifications/initialized", -1, nil), nil)
	notification.Header.Set(mcpSessionHeader, sessionID)
	notificationResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(notificationResponse, notification)
	if notificationResponse.Code != http.StatusAccepted {
		t.Errorf("notification status = %d, want 202", notificationResponse.Code)
	}

	listRequest := webRequest(http.MethodPost, "/mcp", mcpRequestFor("tools/list", 2, nil), nil)
	listRequest.Header.Set(mcpSessionHeader, sessionID)
	listResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(listResponse, listRequest)
	listResult := decodeMCPResponse(t, listResponse)
	if listResult.Error != nil || !strings.Contains(listResponse.Body.String(), "ssh_exec") ||
		!strings.Contains(listResponse.Body.String(), "sftp_list") {
		t.Errorf("tools/list response = %s", listResponse.Body.String())
	}

	execRequest := webRequest(http.MethodPost, "/mcp", mcpRequestFor("tools/call", 3, map[string]any{
		"name":      "ssh_exec",
		"arguments": map[string]any{"command": "whoami"},
	}), nil)
	execRequest.Header.Set(mcpSessionHeader, sessionID)
	execResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(execResponse, execRequest)
	if execResponse.Code != http.StatusOK || !strings.Contains(execResponse.Body.String(), "mcp:whoami") {
		t.Errorf("tools/call ssh_exec = %d %s", execResponse.Code, execResponse.Body.String())
	}

	fileListRequest := webRequest(http.MethodPost, "/mcp", mcpRequestFor("tools/call", 4, map[string]any{
		"name":      "sftp_list",
		"arguments": map[string]any{},
	}), nil)
	fileListRequest.Header.Set(mcpSessionHeader, sessionID)
	fileListResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(fileListResponse, fileListRequest)
	if fileListResponse.Code != http.StatusOK || !strings.Contains(fileListResponse.Body.String(), "README.md") {
		t.Errorf("tools/call sftp_list = %d %s", fileListResponse.Code, fileListResponse.Body.String())
	}
}

func TestMCPRequiresConfiguredConnection(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, webRequest(http.MethodPost, "/mcp", mcpRequestFor("initialize", 1, map[string]any{
		"protocolVersion": "2025-03-26",
	}), nil))
	result := decodeMCPResponse(t, response)
	if result.Error == nil || result.Error.Code != -32001 {
		t.Errorf("initialize without auto config = %+v", result)
	}
}

func TestMCPUnknownMethod(t *testing.T) {
	files := &fakeWebFileSystem{}
	client := &fakeFileClient{fakeWebExecutor: &fakeWebExecutor{}, files: files}
	app, _ := newWebTestApp(t, client, nil)
	app.auto = &sshpass.Config{Host: "example.com"}
	initResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(initResponse, webRequest(http.MethodPost, "/mcp", mcpRequestFor("initialize", 1, nil), nil))
	sessionID := initResponse.Header().Get(mcpSessionHeader)
	request := webRequest(http.MethodPost, "/mcp", mcpRequestFor("unknown/method", 2, nil), nil)
	request.Header.Set(mcpSessionHeader, sessionID)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	result := decodeMCPResponse(t, response)
	if result.Error == nil || result.Error.Code != -32601 {
		t.Errorf("unknown method response = %+v", result)
	}
}

func TestMCPDeleteSession(t *testing.T) {
	client := &fakeWebExecutor{}
	app, _ := newWebTestApp(t, client, nil)
	app.auto = &sshpass.Config{Host: "example.com"}

	initResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(initResponse, webRequest(http.MethodPost, "/mcp", mcpRequestFor("initialize", 1, nil), nil))
	sessionID := initResponse.Header().Get(mcpSessionHeader)
	if sessionID == "" {
		t.Fatal("missing MCP session id")
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/mcp", nil)
	deleteRequest.Header.Set(mcpSessionHeader, sessionID)
	deleteResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("DELETE /mcp status = %d", deleteResponse.Code)
	}
	if client.closed() != 1 {
		t.Errorf("client close count = %d, want 1", client.closed())
	}

	listRequest := webRequest(http.MethodPost, "/mcp", mcpRequestFor("tools/list", 2, nil), nil)
	listRequest.Header.Set(mcpSessionHeader, sessionID)
	listResponse := httptest.NewRecorder()
	app.Handler().ServeHTTP(listResponse, listRequest)
	result := decodeMCPResponse(t, listResponse)
	if result.Error == nil || result.Error.Code != -32003 {
		t.Errorf("tools/list after delete = %+v", result)
	}
}
