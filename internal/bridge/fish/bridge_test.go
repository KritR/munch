package fish

import (
	"strings"
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestRequestFromEnv(t *testing.T) {
	t.Setenv("REQUEST_ID", "req_1")
	t.Setenv("ORIGINAL_BUFFER", "ls")
	t.Setenv("PROMPT_TEXT", "ls")
	t.Setenv("CURSOR_POSITION", "2")

	req, err := RequestFromEnv()
	if err != nil {
		t.Fatalf("RequestFromEnv() error = %v", err)
	}
	if req.RequestID != "req_1" || req.PromptText != "ls" || req.CursorPosition != 2 || req.Shell != protocol.ShellFish {
		t.Fatalf("unexpected request: %#v", req)
	}
}

func TestRequestFromEnvGeneratesRequestID(t *testing.T) {
	t.Setenv("ORIGINAL_BUFFER", "pwd")
	t.Setenv("PROMPT_TEXT", "pwd")
	t.Setenv("CURSOR_POSITION", "3")

	req, err := RequestFromEnv()
	if err != nil {
		t.Fatalf("RequestFromEnv() error = %v", err)
	}
	if req.RequestID == "" {
		t.Fatal("expected generated request ID")
	}
	if !strings.HasPrefix(req.RequestID, "req_") {
		t.Fatalf("expected generated request ID with req_ prefix, got %q", req.RequestID)
	}
}

func TestRequestFromEnvInvalidCursor(t *testing.T) {
	t.Setenv("CURSOR_POSITION", "abc")
	if _, err := RequestFromEnv(); err == nil {
		t.Fatal("expected cursor parse error")
	}
}

func TestResponseAssignments(t *testing.T) {
	assignments := ResponseAssignments(protocol.ShellInvocationResponse{
		SchemaVersion: protocol.SchemaVersion,
		RequestID:     "req_1",
		Action:        protocol.ActionInsert,
		Command:       "echo 'hi' \\\\done",
	})

	if !strings.Contains(assignments, "set -g -- MUNCH_ACTION 'insert'") {
		t.Fatalf("unexpected action assignments: %s", assignments)
	}
	if !strings.Contains(assignments, "set -g -- MUNCH_COMMAND 'echo \\'hi\\' \\\\\\\\done'") {
		t.Fatalf("unexpected command assignments: %s", assignments)
	}
}

func TestResponseAssignmentsEmptyCommand(t *testing.T) {
	assignments := ResponseAssignments(protocol.ShellInvocationResponse{
		SchemaVersion: protocol.SchemaVersion,
		RequestID:     "req_1",
		Action:        protocol.ActionCancel,
	})

	if !strings.Contains(assignments, "set -g -- MUNCH_COMMAND ''") {
		t.Fatalf("expected empty command to be fish-quoted, got %s", assignments)
	}
}
