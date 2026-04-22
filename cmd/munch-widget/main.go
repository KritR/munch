package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/krithikr/munch/internal/app"
	fishbridge "github.com/krithikr/munch/internal/bridge/fish"
	zshbridge "github.com/krithikr/munch/internal/bridge/zsh"
	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/runtime"
)

func main() {
	initLogger()
	if err := run(os.Args[1:]); err != nil {
		slog.Error("munch-widget failed", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	var out io.Writer = os.Stderr
	if path := os.Getenv("MUNCH_LOG_FILE"); path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			out = f
		}
	}

	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
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
	case "fish-bridge":
		return runFishBridge(*configPath, runtime.DevMode(*devAction))
	case "zsh-bridge":
		return runZshBridge(*configPath, runtime.DevMode(*devAction))
	default:
		return fmt.Errorf("unsupported mode: %s", *mode)
	}
}

func runSession(configPath string, devMode runtime.DevMode) error {
	slog.Debug("starting session", "config_path", configPath, "dev_mode", devMode)
	req, err := protocol.DecodeRequest(os.Stdin)
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	return protocol.EncodeResponse(os.Stdout, resp)
}

func runZshBridge(configPath string, devMode runtime.DevMode) error {
	req, err := zshbridge.RequestFromEnv()
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, zshbridge.ResponseAssignments(resp))
	return err
}

func runFishBridge(configPath string, devMode runtime.DevMode) error {
	req, err := fishbridge.RequestFromEnv()
	if err != nil {
		return err
	}
	resp, err := app.RunSession(req, configPath, devMode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, fishbridge.ResponseAssignments(resp))
	return err
}
