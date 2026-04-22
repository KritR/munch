package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krithikr/munch/internal/config"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/provider"
	"github.com/krithikr/munch/internal/provider/cerebras"
	"github.com/krithikr/munch/internal/runtime"
	"github.com/krithikr/munch/internal/suggest"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("munch-widget failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("munch-widget", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	mode := fs.String("mode", "session", "run mode")
	configPath := fs.String("config", "", "path to config file")
	devAction := fs.String("dev-action", "none", "developer action override: none|auto-insert-first")
	if err := fs.Parse(args); err != nil {
		return err
	}

	switch *mode {
	case "session":
		return runSession(*configPath, runtime.DevMode(*devAction))
	default:
		return fmt.Errorf("unsupported mode: %s", *mode)
	}
}

func runSession(configPath string, devMode runtime.DevMode) error {
	req, err := protocol.DecodeRequest(os.Stdin)
	if err != nil {
		return err
	}

	cfg, warnings, err := config.Load(configPath)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		slog.Warn("config warning", "warning", string(warning))
	}

	session := runtime.NewSession(req, buildEngine(cfg))
	session.Start()
	session.UpdatePrompt(req.PromptText)
	session.Generate()

	action := protocol.Action(os.Getenv("MUNCH_STUB_ACTION"))
	if action == "" {
		if autoAction, autoCommand, ok := runtime.ResolveDevAction(devMode, session.Suggestions(), req.PromptText); ok {
			resp, err := session.PrepareAction(autoAction, autoCommand)
			if err != nil {
				return err
			}
			return protocol.EncodeResponse(os.Stdout, resp)
		}
		action = protocol.ActionCancel
	}

	var command string
	switch action {
	case protocol.ActionCancel:
	case protocol.ActionInsert, protocol.ActionExecute:
		command = os.Getenv("MUNCH_STUB_COMMAND")
		if command == "" {
			command = suggest.FirstCommand(session.Suggestions(), req.PromptText)
		}
	default:
		return fmt.Errorf("unsupported stub action: %s", action)
	}

	resp, err := session.PrepareAction(action, command)
	if err != nil {
		return err
	}
	return protocol.EncodeResponse(os.Stdout, resp)
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
