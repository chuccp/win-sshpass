package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	sshpass "github.com/chuccp/win-sshpass"
)

// jsonResult is the structured output emitted when the -json flag is used.
// It is designed for easy consumption by AI agents and automation tools.
type jsonResult struct {
	Success    bool   `json:"success"`
	Host       string `json:"host"`
	Command    string `json:"command,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// jsonState holds the global context needed to emit JSON results.
var jsonState struct {
	enabled   bool
	startTime time.Time
	host      string
	command   string
}

// jsonEnabled reports whether JSON output mode is active.
func jsonEnabled() bool {
	return jsonState.enabled
}

// jsonSetHost records the target host for inclusion in JSON results.
func jsonSetHost(host string) {
	jsonState.host = host
}

// jsonSetCommand records the command being executed for inclusion in JSON results.
func jsonSetCommand(cmd string) {
	jsonState.command = cmd
}

// jsonDurationMs returns the elapsed milliseconds since JSON mode was initialized.
func jsonDurationMs() int64 {
	return time.Since(jsonState.startTime).Milliseconds()
}

// jsonInit marks the start of a JSON-mode operation.
func jsonInit() {
	jsonState.startTime = time.Now()
}

// printJSON writes a jsonResult to stdout as pretty-printed JSON.
func printJSON(r jsonResult) {
	if r.Host == "" {
		r.Host = jsonState.host
	}
	if r.Command == "" {
		r.Command = jsonState.command
	}
	if r.DurationMs == 0 {
		r.DurationMs = jsonDurationMs()
	}
	data, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(data))
}

// jsonSuccess emits a successful JSON result and returns.
func jsonSuccess(stdout string) {
	printJSON(jsonResult{
		Success:  true,
		ExitCode: 0,
		Stdout:   stdout,
	})
}

// jsonFail emits a failed JSON result and exits with code 1.
// exitCode is the remote/operation exit code (-1 for connection failures).
func jsonFail(errMsg string, exitCode int) {
	printJSON(jsonResult{
		Success:  false,
		ExitCode: exitCode,
		Error:    errMsg,
	})
	os.Exit(1)
}

// jsonFailFromError is a convenience wrapper that extracts the exit code from
// err (if possible) and emits a JSON failure result.
func jsonFailFromError(err error, fallbackExitCode int) {
	msg := err.Error()
	code := fallbackExitCode
	if c, ok := exitCodeFromErr(err); ok {
		code = c
	}
	jsonFail(msg, code)
}

// exitCodeFromErr delegates to the SDK's ExitCodeFromError.
func exitCodeFromErr(err error) (int, bool) {
	return sshpass.ExitCodeFromError(err)
}
