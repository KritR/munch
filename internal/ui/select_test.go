package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestSelectFromTTYCancelsOnEmptyInput(t *testing.T) {
	selection, err := selectFromTTY("find todos", []protocol.Suggestion{
		{Command: "rg -n TODO .", Description: "Search TODOs"},
	}, strings.NewReader("\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("selectFromTTY() error = %v", err)
	}
	if selection.Action != protocol.ActionCancel {
		t.Fatalf("unexpected action: %s", selection.Action)
	}
}

func TestSelectFromTTYSelectsNumber(t *testing.T) {
	selection, err := selectFromTTY("find todos", []protocol.Suggestion{
		{Command: "rg -n TODO .", Description: "Search TODOs"},
		{Command: "grep -R TODO .", Description: "Fallback grep"},
	}, strings.NewReader("2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("selectFromTTY() error = %v", err)
	}
	if selection.Action != protocol.ActionInsert {
		t.Fatalf("unexpected action: %s", selection.Action)
	}
	if selection.Command != "grep -R TODO ." {
		t.Fatalf("unexpected command: %q", selection.Command)
	}
}
