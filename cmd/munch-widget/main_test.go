package main

import (
	"os"
	"strings"
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestResponseAssignments(t *testing.T) {
	assignments := responseAssignments(protocol.ShellInvocationResponse{
		SchemaVersion: protocol.SchemaVersion,
		RequestID:     "req_1",
		Action:        protocol.ActionInsert,
		Command:       "echo 'hi'",
	})

	if !strings.Contains(assignments, "MUNCH_ACTION='insert'") {
		t.Fatalf("unexpected action assignments: %s", assignments)
	}
	if !strings.Contains(assignments, "MUNCH_COMMAND='echo '\"'\"'hi'\"'\"''") {
		t.Fatalf("unexpected command assignments: %s", assignments)
	}
}

func TestRequestFromEnv(t *testing.T) {
	t.Setenv("REQUEST_ID", "req_1")
	t.Setenv("ORIGINAL_BUFFER", "ls")
	t.Setenv("PROMPT_TEXT", "ls")
	t.Setenv("CURSOR_POSITION", "2")

	req, err := requestFromEnv(protocol.ShellZsh)
	if err != nil {
		t.Fatalf("requestFromEnv() error = %v", err)
	}
	if req.RequestID != "req_1" || req.PromptText != "ls" || req.CursorPosition != 2 {
		t.Fatalf("unexpected request: %#v", req)
	}
}

func TestRequestFromEnvInvalidCursor(t *testing.T) {
	t.Setenv("CURSOR_POSITION", "abc")
	if _, err := requestFromEnv(protocol.ShellZsh); err == nil {
		t.Fatal("expected cursor parse error")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
