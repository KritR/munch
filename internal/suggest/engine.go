package suggest

import (
	munchctx "github.com/krithikr/munch/internal/context"
	"github.com/krithikr/munch/internal/prompting"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/provider"
	fakeprovider "github.com/krithikr/munch/internal/provider/fake"
)

type Engine interface {
	Generate(prompt string, ctx munchctx.Normalized) []protocol.Suggestion
}

type ProviderBackedEngine struct {
	Client          provider.Client
	SuggestionCount int
}

func NewFakeEngine() Engine {
	return ProviderBackedEngine{
		Client:          fakeprovider.Client{},
		SuggestionCount: 5,
	}
}

func (e ProviderBackedEngine) Generate(prompt string, ctx munchctx.Normalized) []protocol.Suggestion {
	client := e.Client
	if client == nil {
		client = fakeprovider.Client{}
	}

	count := e.SuggestionCount
	if count <= 0 {
		count = 5
	}

	rendered := prompting.RenderContext(ctx, prompt, count)
	resp, err := client.Generate(provider.GenerationRequest{
		PromptText:      prompt,
		RenderedContext: rendered,
		SuggestionCount: count,
	})
	if err != nil {
		return nil
	}

	return resp.Suggestions
}

func FirstCommand(suggestions []protocol.Suggestion, fallback string) string {
	if len(suggestions) == 0 || suggestions[0].Command == "" {
		return fallback
	}
	return suggestions[0].Command
}
