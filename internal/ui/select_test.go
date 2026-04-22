package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/krithikr/munch/internal/protocol"
)

func TestSelectorEnterReturnsSelectedSuggestion(t *testing.T) {
	model := selectorModel{
		prompt: "find todos",
		suggestions: []protocol.Suggestion{
			{Command: "rg -n TODO .", Description: "Search TODOs"},
			{Command: "grep -R TODO .", Description: "Fallback grep"},
		},
		selected: 1,
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	got := next.(selectorModel)
	if got.selection.Action != protocol.ActionInsert {
		t.Fatalf("unexpected action: %s", got.selection.Action)
	}
	if got.selection.Command != "grep -R TODO ." {
		t.Fatalf("unexpected command: %q", got.selection.Command)
	}
}

func TestSelectorEscCancels(t *testing.T) {
	model := selectorModel{
		prompt: "find todos",
		suggestions: []protocol.Suggestion{
			{Command: "rg -n TODO .", Description: "Search TODOs"},
		},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := next.(selectorModel)
	if got.selection.Action != protocol.ActionCancel {
		t.Fatalf("unexpected action: %s", got.selection.Action)
	}
}

func TestSelectorMovesDown(t *testing.T) {
	model := selectorModel{
		suggestions: []protocol.Suggestion{
			{Command: "one"},
			{Command: "two"},
		},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	got := next.(selectorModel)
	if got.selected != 1 {
		t.Fatalf("unexpected selected index: %d", got.selected)
	}
}

func TestSelectorEnterRequiresConfirmationFirst(t *testing.T) {
	model := selectorModel{
		prompt: "delete build",
		suggestions: []protocol.Suggestion{
			{Command: "rm -rf build", Description: "Delete build", RequiresConfirmation: true, ConfirmationReason: "This command may delete files."},
		},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(selectorModel)
	if !got.confirming {
		t.Fatal("expected confirming state")
	}
	if got.selection.Action != "" {
		t.Fatalf("unexpected selection action: %s", got.selection.Action)
	}
}

func TestSelectorConfirmingEnterConfirms(t *testing.T) {
	model := selectorModel{
		prompt:     "delete build",
		confirming: true,
		suggestions: []protocol.Suggestion{
			{Command: "rm -rf build", Description: "Delete build", RequiresConfirmation: true},
		},
	}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(selectorModel)
	if got.selection.Action != protocol.ActionInsert {
		t.Fatalf("unexpected action: %s", got.selection.Action)
	}
}
