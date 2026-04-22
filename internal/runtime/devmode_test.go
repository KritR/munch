package runtime

import (
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestResolveDevActionAutoInsertFirst(t *testing.T) {
	action, command, ok := ResolveDevAction(DevModeAutoInsertFirst, []protocol.Suggestion{
		{Command: "echo hi"},
	}, "fallback")
	if !ok {
		t.Fatal("expected dev action to resolve")
	}
	if action != protocol.ActionInsert {
		t.Fatalf("unexpected action: %s", action)
	}
	if command != "echo hi" {
		t.Fatalf("unexpected command: %q", command)
	}
}

func TestResolveDevActionAutoInsertFirstFallsBack(t *testing.T) {
	action, command, ok := ResolveDevAction(DevModeAutoInsertFirst, nil, "fallback")
	if !ok {
		t.Fatal("expected dev action to resolve")
	}
	if action != protocol.ActionInsert || command != "fallback" {
		t.Fatalf("unexpected result: action=%s command=%q", action, command)
	}
}
