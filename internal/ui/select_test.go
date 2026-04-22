package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/krithikr/munch/internal/protocol"
)

func TestSelectorEnterReturnsSelectedSuggestion(t *testing.T) {
	model := selectorModel{
		input: textinput.New(),
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
	model := selectorModel{input: textinput.New()}

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := next.(selectorModel)
	if got.selection.Action != protocol.ActionCancel {
		t.Fatalf("unexpected action: %s", got.selection.Action)
	}
}

func TestSelectorMovesDown(t *testing.T) {
	model := selectorModel{
		input: textinput.New(),
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
		input: textinput.New(),
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
		input:      textinput.New(),
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

func TestGenerateResultIgnoresStaleVersion(t *testing.T) {
	model := selectorModel{
		input:   textinput.New(),
		version: 3,
		suggestions: []protocol.Suggestion{
			{Command: "current"},
		},
	}

	next, _ := model.Update(generateResultMsg{
		version: 2,
		suggestions: []protocol.Suggestion{
			{Command: "stale"},
		},
	})
	got := next.(selectorModel)
	if got.suggestions[0].Command != "current" {
		t.Fatalf("expected stale result to be ignored, got %#v", got.suggestions)
	}
}

func TestPromptEditSchedulesNewVersionAndLoading(t *testing.T) {
	input := textinput.New()
	input.Focus()
	model := selectorModel{
		input:   input,
		version: 1,
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatal("expected debounce command")
	}
	got := next.(selectorModel)
	if got.version != 2 {
		t.Fatalf("expected version to increment, got %d", got.version)
	}
	if !got.loading {
		t.Fatal("expected loading to be set immediately on prompt edit")
	}
	if got.input.Value() != "a" {
		t.Fatalf("unexpected prompt value: %q", got.input.Value())
	}
}
