package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/krithikr/munch/internal/config"
	munchctx "github.com/krithikr/munch/internal/context"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/provider"
	"github.com/krithikr/munch/internal/provider/cerebras"
	"github.com/krithikr/munch/internal/runtime"
	"github.com/krithikr/munch/internal/suggest"
	"github.com/krithikr/munch/internal/ui"
)

func RunSession(req protocol.ShellInvocationRequest, configPath string, devMode runtime.DevMode) (protocol.ShellInvocationResponse, error) {
	slog.Debug("decoded request", "request_id", req.RequestID, "shell", req.Shell, "prompt_text", req.PromptText)

	cfg, warnings, err := config.Load(configPath)
	if err != nil {
		return protocol.ShellInvocationResponse{}, err
	}
	for _, warning := range warnings {
		slog.Warn("config warning", "warning", string(warning))
	}

	engine := buildEngine(cfg)
	slog.Debug("selected engine", "engine", engine.Name())

	ctx := munchctx.NewCollector().Collect(req.Shell)
	session := runtime.NewSessionWithContextAndSafetyLevel(req, engine, ctx, cfg.Safety.Level)
	session.Start()

	action := protocol.Action(os.Getenv("MUNCH_STUB_ACTION"))
	if action == "" {
		if devMode != runtime.DevModeNone {
			session.UpdatePrompt(req.PromptText)
			if err := session.GenerateWithError(); err != nil {
				slog.Warn("generation failed", "error", err)
			} else {
				slog.Debug("generation succeeded")
			}
			slog.Debug("generated suggestions", "count", len(session.Suggestions()))
		}

		if autoAction, autoCommand, ok := runtime.ResolveDevAction(devMode, session.Suggestions(), req.PromptText); ok {
			slog.Debug("resolved dev action", "action", autoAction, "command", autoCommand)
			resp, err := session.PrepareAction(autoAction, autoCommand)
			if err != nil {
				return protocol.ShellInvocationResponse{}, err
			}
			slog.Debug("returning response", "action", resp.Action, "command", resp.Command)
			return resp, nil
		}

		selection, err := ui.SelectSuggestion(req.PromptText, engine, ctx, cfg.Safety.Level)
		if err != nil {
			return protocol.ShellInvocationResponse{}, err
		}
		slog.Debug("resolved interactive selection", "action", selection.Action, "command", selection.Command)
		resp, err := session.PrepareAction(selection.Action, selection.Command)
		if err != nil {
			return protocol.ShellInvocationResponse{}, err
		}
		slog.Debug("returning response", "action", resp.Action, "command", resp.Command)
		return resp, nil
	}

	var command string
	switch action {
	case protocol.ActionCancel:
	case protocol.ActionInsert, protocol.ActionExecute:
		session.UpdatePrompt(req.PromptText)
		if err := session.GenerateWithError(); err != nil {
			slog.Warn("generation failed", "error", err)
		} else {
			slog.Debug("generation succeeded")
		}
		slog.Debug("generated suggestions", "count", len(session.Suggestions()))
		command = os.Getenv("MUNCH_STUB_COMMAND")
		if command == "" {
			command = suggest.FirstCommand(session.Suggestions(), req.PromptText)
		}
	default:
		return protocol.ShellInvocationResponse{}, errUnsupportedStubAction(action)
	}

	resp, err := session.PrepareAction(action, command)
	if err != nil {
		return protocol.ShellInvocationResponse{}, err
	}
	slog.Debug("returning response", "action", resp.Action, "command", resp.Command)
	return resp, nil
}

func buildEngine(cfg config.Config) suggest.Engine {
	client := buildProviderClient(cfg)
	if client == nil {
		return suggest.NewFakeEngine()
	}
	return suggest.NewEngine(client, cfg.UI.VisibleSuggestionCount)
}

func buildProviderClient(cfg config.Config) provider.Client {
	if !cfg.HasProviderConfig() {
		return nil
	}

	apiKey := os.Getenv(cfg.Provider.APIKeyEnv)
	if apiKey == "" {
		slog.Warn("provider API key env var is unset; falling back to fake provider", "env_var", cfg.Provider.APIKeyEnv)
		return nil
	}

	return cerebras.Client{
		BaseURL:    cfg.Provider.BaseURL,
		APIKey:     apiKey,
		Model:      cfg.Provider.Model,
		Timeout:    time.Duration(cfg.Provider.TimeoutMS) * time.Millisecond,
		MaxRetries: cfg.Provider.MaxRetries,
	}
}

func errUnsupportedStubAction(action protocol.Action) error {
	return &unsupportedStubActionError{action: action}
}

type unsupportedStubActionError struct {
	action protocol.Action
}

func (e *unsupportedStubActionError) Error() string {
	return "unsupported stub action: " + string(e.action)
}
