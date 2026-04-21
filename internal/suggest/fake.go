package suggest

import (
	"fmt"
	"strings"

	"github.com/krithikr/munch/internal/protocol"
)

type Engine interface {
	Generate(prompt string) []protocol.Suggestion
}

type FakeEngine struct{}

func (FakeEngine) Generate(prompt string) []protocol.Suggestion {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil
	}

	suggestion := protocol.Suggestion{
		Command:              prompt,
		Description:          "Use the current prompt text as the command",
		Risk:                 "low",
		RequiresConfirmation: false,
	}

	lower := strings.ToLower(prompt)
	switch {
	case strings.Contains(lower, "sudo"), strings.Contains(lower, "rm "), strings.Contains(lower, "delete"):
		suggestion.Risk = "high"
		suggestion.RequiresConfirmation = true
		suggestion.Description = "Potentially destructive command inferred from the prompt"
	case strings.Contains(lower, "mkdir"), strings.Contains(lower, "touch"), strings.Contains(lower, "mv "), strings.Contains(lower, "cp "):
		suggestion.Risk = "medium"
		suggestion.RequiresConfirmation = true
		suggestion.Description = "Mutating command inferred from the prompt"
	case strings.Contains(lower, "list "), strings.Contains(lower, "show "), strings.Contains(lower, "find "), strings.Contains(lower, "search "):
		suggestion.Command = fmt.Sprintf("echo %q", prompt)
		suggestion.Description = "Placeholder read-only suggestion for the current prompt"
	}

	return []protocol.Suggestion{suggestion}
}
