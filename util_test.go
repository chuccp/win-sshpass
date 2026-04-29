package main

import (
	"fmt"
	"testing"
)

type testExitStatusError struct {
	code int
}

func (e testExitStatusError) Error() string {
	return "remote command failed"
}

func (e testExitStatusError) ExitStatus() int {
	return e.code
}

func TestExitCodeFromError(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", testExitStatusError{code: 7})

	code, ok := exitCodeFromError(err)
	if !ok {
		t.Fatal("expected exit status error to be detected")
	}
	if code != 7 {
		t.Fatalf("code = %d, want 7", code)
	}
}
