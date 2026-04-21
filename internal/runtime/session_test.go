package runtime

import (
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestSessionLifecycleWithGeneratedSuggestion(t *testing.T) {
	req := protocol.ShellInvocationRequest{
		SchemaVersion:  protocol.SchemaVersion,
		RequestID:      "req_1",
		Shell:          protocol.ShellZsh,
		OriginalBuffer: "ls",
		PromptText:     "ls",
		CursorPosition: 2,
	}

	session := NewSession(req, nil)
	if session.State() != StateInitializing {
		t.Fatalf("unexpected initial state: %s", session.State())
	}

	session.Start()
	if session.State() != StateReady {
		t.Fatalf("unexpected ready state: %s", session.State())
	}

	session.UpdatePrompt("echo hi")
	if session.State() != StateLoadingSuggestions {
		t.Fatalf("unexpected loading state: %s", session.State())
	}

	session.Generate()
	if session.State() != StateShowingSuggestions {
		t.Fatalf("unexpected showing state: %s", session.State())
	}

	suggestions := session.Suggestions()
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	resp, err := session.PrepareAction(protocol.ActionInsert, suggestions[0].Command)
	if err != nil {
		t.Fatalf("PrepareAction() error = %v", err)
	}
	if resp.Action != protocol.ActionInsert {
		t.Fatalf("unexpected action: %s", resp.Action)
	}
	if session.State() != StateClosed {
		t.Fatalf("unexpected final state: %s", session.State())
	}
}

func TestSessionPrepareCancel(t *testing.T) {
	req := protocol.ShellInvocationRequest{
		SchemaVersion:  protocol.SchemaVersion,
		RequestID:      "req_1",
		Shell:          protocol.ShellZsh,
		OriginalBuffer: "",
		PromptText:     "",
		CursorPosition: 0,
	}

	session := NewSession(req, nil)
	session.Start()

	resp, err := session.PrepareAction(protocol.ActionCancel, "")
	if err != nil {
		t.Fatalf("PrepareAction() error = %v", err)
	}
	if resp.Action != protocol.ActionCancel {
		t.Fatalf("unexpected cancel action: %s", resp.Action)
	}
}

func TestPrepareActionRequiresCommandForInsert(t *testing.T) {
	req := protocol.ShellInvocationRequest{
		SchemaVersion:  protocol.SchemaVersion,
		RequestID:      "req_1",
		Shell:          protocol.ShellZsh,
		OriginalBuffer: "",
		PromptText:     "",
		CursorPosition: 0,
	}

	session := NewSession(req, nil)
	session.Start()

	if _, err := session.PrepareAction(protocol.ActionInsert, ""); err == nil {
		t.Fatal("expected missing command error")
	}
}
