package fake

import (
	"fmt"
	"strings"

	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/provider"
)

type Client struct{}

func (Client) Name() string {
	return "fake"
}

func (Client) Generate(req provider.GenerationRequest) (provider.Response, error) {
	prompt := strings.TrimSpace(req.PromptText)
	if prompt == "" {
		prompt = strings.TrimSpace(req.UserPrompt)
	}
	if prompt == "" {
		return provider.Response{}, nil
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
		suggestion.UsesTools = []string{"echo"}
	}

	resp := provider.Response{
		Suggestions: []protocol.Suggestion{suggestion},
	}
	if req.SuggestionCount > 0 && len(resp.Suggestions) > req.SuggestionCount {
		resp.Suggestions = resp.Suggestions[:req.SuggestionCount]
	}
	return resp, nil
}
