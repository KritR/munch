package runtime

import "github.com/krithikr/munch/internal/protocol"

func (s *Session) GenerateWithError() error {
	suggestions, err := s.engine.Generate(s.promptText, s.context, s.safetyLevel)
	if err != nil {
		s.suggestions = nil
		s.state = StateShowingSuggestions
		return err
	}
	s.suggestions = suggestions
	s.state = StateShowingSuggestions
	return nil
}

func (s *Session) SuggestionsOrNil() []protocol.Suggestion {
	return s.suggestions
}
