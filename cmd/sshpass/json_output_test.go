package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"
)

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = previous })

	fn()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previous
	data, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPrintJSONUsesStateFallbacks(t *testing.T) {
	previous := jsonState
	t.Cleanup(func() { jsonState = previous })
	jsonState.host = "root@example.com"
	jsonState.command = "uptime"
	jsonState.startTime = time.Now().Add(-10 * time.Millisecond)

	data := captureStdout(t, func() {
		printJSON(jsonResult{Success: true, ExitCode: 0, Stdout: "ok"})
	})

	var got jsonResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", data, err)
	}
	if !got.Success || got.Host != "root@example.com" || got.Command != "uptime" || got.Stdout != "ok" {
		t.Errorf("printJSON result = %+v", got)
	}
	if got.DurationMs <= 0 {
		t.Errorf("DurationMs = %d, want > 0", got.DurationMs)
	}
}

func TestPrintJSONPreservesExplicitValues(t *testing.T) {
	previous := jsonState
	t.Cleanup(func() { jsonState = previous })
	jsonState.host = "fallback-host"
	jsonState.command = "fallback-command"

	data := captureStdout(t, func() {
		printJSON(jsonResult{
			Success:    false,
			Host:       "explicit-host",
			Command:    "explicit-command",
			ExitCode:   7,
			Stderr:     "failed",
			Error:      "bad",
			DurationMs: 42,
		})
	})

	var got jsonResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", data, err)
	}
	if got.Host != "explicit-host" || got.Command != "explicit-command" || got.ExitCode != 7 || got.DurationMs != 42 {
		t.Errorf("printJSON result = %+v", got)
	}
}

func TestJSONSuccess(t *testing.T) {
	previous := jsonState
	t.Cleanup(func() { jsonState = previous })
	jsonState.host = "root@example.com"
	jsonState.command = "whoami"
	jsonInit()

	data := captureStdout(t, func() {
		jsonSuccess("root\n")
	})

	var got jsonResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", data, err)
	}
	if !got.Success || got.ExitCode != 0 || got.Stdout != "root\n" {
		t.Errorf("jsonSuccess result = %+v", got)
	}
}
