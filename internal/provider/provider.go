package provider

import "github.com/krithikr/munch/internal/protocol"

type GenerationRequest struct {
	PromptText      string
	RenderedContext string
	SuggestionCount int
}

type Response struct {
	Suggestions []protocol.Suggestion
}

type Client interface {
	Generate(GenerationRequest) (Response, error)
}
