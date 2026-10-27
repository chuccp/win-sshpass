package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"sort"
	"strings"

	sshpass "github.com/chuccp/win-sshpass"
)

const (
	mcpSessionHeader   = "Mcp-Session-Id"
	mcpProtocolHeader  = "MCP-Protocol-Version"
	mcpProtocolVersion = "2025-03-26"
	mcpReadLimit       = 1 << 20
	mcpWriteLimit      = 16 << 20
)

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpToolResult struct {
	Content           []mcpContent `json:"content"`
	StructuredContent any          `json:"structuredContent,omitempty"`
	IsError           bool         `json:"isError,omitempty"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

var mcpTools = []mcpTool{
	{
		Name:        "ssh_exec",
		Description: "Run a command in the connected SSH account and return stdout, stderr, and the exit code.",
		InputSchema: objectSchema(map[string]any{
			"command": stringSchema("Command to execute"),
		}, []string{"command"}),
	},
	{
		Name:        "sftp_list",
		Description: "List files and directories at a remote SFTP path.",
		InputSchema: objectSchema(map[string]any{
			"path": stringSchema("Remote directory path; omit to use the login home directory"),
		}, nil),
	},
	{
		Name:        "sftp_read",
		Description: "Read a UTF-8 text file from the remote SFTP filesystem.",
		InputSchema: objectSchema(map[string]any{
			"path":      stringSchema("Remote file path"),
			"max_bytes": map[string]any{"type": "integer", "minimum": 1, "maximum": mcpReadLimit, "default": 65536},
		}, []string{"path"}),
	},
	{
		Name:        "sftp_write",
		Description: "Create or overwrite a remote SFTP file using text or base64 content.",
		InputSchema: objectSchema(map[string]any{
			"path":     stringSchema("Remote file path"),
			"content":  stringSchema("File content"),
			"encoding": map[string]any{"type": "string", "enum": []string{"text", "base64"}, "default": "text"},
		}, []string{"path", "content"}),
	},
	{
		Name:        "sftp_mkdir",
		Description: "Create a remote directory, including missing parents.",
		InputSchema: objectSchema(map[string]any{
			"path": stringSchema("Remote directory path"),
		}, []string{"path"}),
	},
	{
		Name:        "sftp_delete",
		Description: "Delete a remote file or empty directory.",
		InputSchema: objectSchema(map[string]any{
			"path": stringSchema("Remote file or empty directory path"),
		}, []string{"path"}),
	},
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func (a *webApp) handleMCP(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	var request mcpRequest
	if err := decodeWebJSON(w, r, &request); err != nil {
		writeMCPResponse(w, nil, nil, &mcpError{Code: -32700, Message: err.Error()})
		return
	}
	if request.JSONRPC != "2.0" || request.Method == "" {
		writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32600, Message: "invalid JSON-RPC request"})
		return
	}

	switch request.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(request.Params, &params)
		if a.auto == nil {
			writeMCPResponse(w, request.ID, nil, &mcpError{
				Code:    -32001,
				Message: "MCP requires SSH connection settings on the win-sshpass -web command",
			})
			return
		}
		sessionID, session, err := a.newSession(a.auto)
		if err != nil {
			writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32002, Message: "SSH connection failed", Data: err.Error()})
			return
		}
		a.sessionsMu.Lock()
		a.sessions[sessionID] = session
		a.sessionsMu.Unlock()
		w.Header().Set(mcpSessionHeader, sessionID)
		w.Header().Set(mcpProtocolHeader, mcpProtocolVersion)
		version := params.ProtocolVersion
		if version != "2024-11-05" && version != mcpProtocolVersion {
			version = mcpProtocolVersion
		}
		writeMCPResponse(w, request.ID, map[string]any{
			"protocolVersion": version,
			"capabilities": map[string]any{
				"tools": map[string]any{"listChanged": false},
			},
			"serverInfo": map[string]any{
				"name":    "win-sshpass-web",
				"version": sshpass.Version,
			},
			"instructions": "Tools operate on the SSH connection configured when the web server started.",
		}, nil)
		return
	case "notifications/initialized", "notifications/cancelled":
		w.WriteHeader(http.StatusAccepted)
		return
	case "ping":
		writeMCPResponse(w, request.ID, map[string]any{}, nil)
		return
	case "tools/list":
		if a.mcpSession(r) == nil {
			writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32003, Message: "MCP SSH session not found"})
			return
		}
		writeMCPResponse(w, request.ID, map[string]any{"tools": mcpTools}, nil)
		return
	case "tools/call":
		session := a.mcpSession(r)
		if session == nil {
			writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32003, Message: "MCP SSH session not found"})
			return
		}
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil || params.Name == "" {
			writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32602, Message: "invalid tools/call parameters"})
			return
		}
		result := a.callMCPTool(session, params.Name, params.Arguments)
		writeMCPResponse(w, request.ID, result, nil)
		return
	default:
		writeMCPResponse(w, request.ID, nil, &mcpError{Code: -32601, Message: "method not found"})
	}
}

func (a *webApp) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		webError(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	id := strings.TrimSpace(r.Header.Get(mcpSessionHeader))
	if id == "" {
		id = webSessionCookie(r)
	}
	if id == "" {
		webError(w, "MCP session id is required", http.StatusBadRequest)
		return
	}
	session := a.removeSession(id)
	if session == nil {
		webError(w, "MCP SSH session not found", http.StatusNotFound)
		return
	}
	session.client.Close()
	w.WriteHeader(http.StatusNoContent)
}

func (a *webApp) mcpSession(r *http.Request) *webSession {
	if id := strings.TrimSpace(r.Header.Get(mcpSessionHeader)); id != "" {
		return a.sessionByID(id)
	}
	return a.sessionFromRequest(r)
}

func writeMCPResponse(w http.ResponseWriter, id json.RawMessage, result any, rpcErr *mcpError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("MCP-Protocol-Version", mcpProtocolVersion)
	if id == nil {
		id = json.RawMessage("null")
	}
	response := mcpResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   rpcErr,
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func mcpToolError(message string) mcpToolResult {
	return mcpToolResult{
		Content: []mcpContent{{Type: "text", Text: message}},
		IsError: true,
	}
}

func (a *webApp) callMCPTool(session *webSession, name string, rawArgs json.RawMessage) mcpToolResult {
	switch name {
	case "ssh_exec":
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || strings.TrimSpace(args.Command) == "" {
			return mcpToolError("command is required")
		}
		session.mu.Lock()
		stdout, stderr, exitCode, err := session.client.ExecCapture(args.Command)
		session.mu.Unlock()
		if err != nil {
			return mcpToolError("command failed: " + err.Error())
		}
		text := fmt.Sprintf("exit_code: %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout, stderr)
		return mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: text}},
			StructuredContent: map[string]any{
				"exit_code": exitCode,
				"stdout":    stdout,
				"stderr":    stderr,
			},
			IsError: exitCode != 0,
		}
	case "sftp_list":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return mcpToolError("invalid arguments: " + err.Error())
		}
		files, err := a.mcpFileSystem(session)
		if err != nil {
			return mcpToolError(err.Error())
		}
		defer files.Close()
		current := strings.TrimSpace(args.Path)
		if current == "" {
			current, _ = files.Getwd()
			if current == "" || current == "." {
				current = "/"
			}
		}
		current = path.Clean(current)
		infos, err := files.ReadDir(current)
		if err != nil {
			return mcpToolError("cannot list directory: " + err.Error())
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
		data, _ := json.MarshalIndent(entries, "", "  ")
		return mcpToolResult{
			Content:           []mcpContent{{Type: "text", Text: string(data)}},
			StructuredContent: map[string]any{"path": current, "entries": entries},
		}
	case "sftp_read":
		var args struct {
			Path     string `json:"path"`
			MaxBytes int64  `json:"max_bytes"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || strings.TrimSpace(args.Path) == "" {
			return mcpToolError("path is required")
		}
		if args.MaxBytes <= 0 || args.MaxBytes > mcpReadLimit {
			args.MaxBytes = mcpReadLimit
		}
		files, err := a.mcpFileSystem(session)
		if err != nil {
			return mcpToolError(err.Error())
		}
		defer files.Close()
		file, err := files.Open(path.Clean(args.Path))
		if err != nil {
			return mcpToolError("cannot open file: " + err.Error())
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, args.MaxBytes+1))
		if err != nil {
			return mcpToolError("cannot read file: " + err.Error())
		}
		truncated := int64(len(data)) > args.MaxBytes
		if truncated {
			data = data[:args.MaxBytes]
		}
		return mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: string(data)}},
			StructuredContent: map[string]any{
				"path":      path.Clean(args.Path),
				"truncated": truncated,
				"bytes":     len(data),
			},
		}
	case "sftp_write":
		var args struct {
			Path     string `json:"path"`
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || strings.TrimSpace(args.Path) == "" {
			return mcpToolError("path is required")
		}
		data := []byte(args.Content)
		if args.Encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(args.Content)
			if err != nil {
				return mcpToolError("invalid base64 content: " + err.Error())
			}
			data = decoded
		}
		if len(data) > mcpWriteLimit {
			return mcpToolError(fmt.Sprintf("content exceeds %d bytes", mcpWriteLimit))
		}
		files, err := a.mcpFileSystem(session)
		if err != nil {
			return mcpToolError(err.Error())
		}
		defer files.Close()
		remotePath := path.Clean(args.Path)
		if err := files.MkdirAll(path.Dir(remotePath)); err != nil {
			return mcpToolError("cannot create parent directory: " + err.Error())
		}
		file, err := files.Create(remotePath)
		if err != nil {
			return mcpToolError("cannot create file: " + err.Error())
		}
		written, writeErr := file.Write(data)
		closeErr := file.Close()
		if writeErr != nil {
			return mcpToolError("cannot write file: " + writeErr.Error())
		}
		if closeErr != nil {
			return mcpToolError("cannot close file: " + closeErr.Error())
		}
		return mcpToolResult{
			Content:           []mcpContent{{Type: "text", Text: fmt.Sprintf("wrote %d bytes to %s", written, remotePath)}},
			StructuredContent: map[string]any{"path": remotePath, "bytes": written},
		}
	case "sftp_mkdir":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || strings.TrimSpace(args.Path) == "" {
			return mcpToolError("path is required")
		}
		files, err := a.mcpFileSystem(session)
		if err != nil {
			return mcpToolError(err.Error())
		}
		defer files.Close()
		remotePath := path.Clean(args.Path)
		if err := files.MkdirAll(remotePath); err != nil {
			return mcpToolError("cannot create directory: " + err.Error())
		}
		return mcpToolResult{
			Content:           []mcpContent{{Type: "text", Text: "created " + remotePath}},
			StructuredContent: map[string]any{"path": remotePath},
		}
	case "sftp_delete":
		var args struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil || strings.TrimSpace(args.Path) == "" {
			return mcpToolError("path is required")
		}
		remotePath := path.Clean(args.Path)
		if remotePath == "." || remotePath == "/" {
			return mcpToolError("refusing to delete this path")
		}
		files, err := a.mcpFileSystem(session)
		if err != nil {
			return mcpToolError(err.Error())
		}
		defer files.Close()
		info, err := files.Stat(remotePath)
		if err != nil {
			return mcpToolError("cannot stat path: " + err.Error())
		}
		if info.IsDir() {
			err = files.RemoveDirectory(remotePath)
		} else {
			err = files.Remove(remotePath)
		}
		if err != nil {
			return mcpToolError("cannot delete path: " + err.Error())
		}
		return mcpToolResult{
			Content:           []mcpContent{{Type: "text", Text: "deleted " + remotePath}},
			StructuredContent: map[string]any{"path": remotePath},
		}
	default:
		return mcpToolError("unknown tool: " + name)
	}
}

func (a *webApp) mcpFileSystem(session *webSession) (webFileSystem, error) {
	provider, ok := session.client.(webFileProvider)
	if !ok {
		return nil, fmt.Errorf("SFTP is unavailable for this connection")
	}
	return provider.FileSystem()
}
