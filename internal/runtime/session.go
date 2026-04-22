package runtime

import (
	"fmt"

	munchctx "github.com/krithikr/munch/internal/context"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/suggest"
)

type State string

const (
	StateInitializing       State = "Initializing"
	StateReady              State = "Ready"
	StateLoadingSuggestions State = "LoadingSuggestions"
	StateShowingSuggestions State = "ShowingSuggestions"
	StateConfirmingAction   State = "ConfirmingAction"
	StateCompleting         State = "Completing"
	StateClosed             State = "Closed"
)

type Session struct {
	req           protocol.ShellInvocationRequest
	engine        suggest.Engine
	context       munchctx.Normalized
	safetyLevel   string
	state         State
	promptText    string
	promptVersion int
	suggestions   []protocol.Suggestion
	finalAction   *protocol.ShellInvocationResponse
}

func NewSession(req protocol.ShellInvocationRequest, engine suggest.Engine) *Session {
	return NewSessionWithSafetyLevel(req, engine, "balanced")
}

func NewSessionWithSafetyLevel(req protocol.ShellInvocationRequest, engine suggest.Engine, safetyLevel string) *Session {
	if engine == nil {
		engine = suggest.NewFakeEngine()
	}
	return &Session{
		req:         req,
		engine:      engine,
		context:     munchctx.CollectBootstrap(),
		safetyLevel: safetyLevel,
		state:       StateInitializing,
		promptText:  req.PromptText,
	}
}

func (s *Session) State() State {
	return s.state
}

func (s *Session) Suggestions() []protocol.Suggestion {
	out := make([]protocol.Suggestion, len(s.suggestions))
	copy(out, s.suggestions)
	return out
}

func (s *Session) Start() {
	if s.state != StateInitializing {
		return
	}
	s.state = StateReady
}

func (s *Session) UpdatePrompt(prompt string) {
	s.promptText = prompt
	s.promptVersion++
	s.state = StateLoadingSuggestions
}

func (s *Session) Generate() {
	suggestions, err := s.engine.Generate(s.promptText, s.context, s.safetyLevel)
	if err != nil {
		s.suggestions = nil
		s.state = StateShowingSuggestions
		return
	}
	s.suggestions = suggestions
	s.state = StateShowingSuggestions
}

func (s *Session) PrepareAction(action protocol.Action, command string) (protocol.ShellInvocationResponse, error) {
	if s.state == StateInitializing {
		return protocol.ShellInvocationResponse{}, fmt.Errorf("session not started")
	}

	resp := protocol.ShellInvocationResponse{
		SchemaVersion: protocol.SchemaVersion,
		RequestID:     s.req.RequestID,
		Action:        action,
	}

	switch action {
	case protocol.ActionCancel:
		s.state = StateCompleting
	case protocol.ActionInsert, protocol.ActionExecute:
		if command == "" {
			return protocol.ShellInvocationResponse{}, fmt.Errorf("command required for action %q", action)
		}
		resp.Command = command
		s.state = StateCompleting
	default:
		return protocol.ShellInvocationResponse{}, fmt.Errorf("unsupported action %q", action)
	}

	if err := resp.Validate(); err != nil {
		return protocol.ShellInvocationResponse{}, err
	}

	s.finalAction = &resp
	s.state = StateClosed
	return resp, nil
}
