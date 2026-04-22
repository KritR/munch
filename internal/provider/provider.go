package provider

import "github.com/krithikr/munch/internal/protocol"

type GenerationRequest struct {
	SystemPrompt    string
	UserPrompt      string
	PromptText      string
	SuggestionCount int
}

type Response struct {
	Suggestions []protocol.Suggestion
}

type Client interface {
	Name() string
	Generate(GenerationRequest) (Response, error)
}
