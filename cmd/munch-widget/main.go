package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/runtime"
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
	if err := fs.Parse(args); err != nil {
		return err
	}

	switch *mode {
	case "session":
		return runSession()
	default:
		return fmt.Errorf("unsupported mode: %s", *mode)
	}
}

func runSession() error {
	req, err := protocol.DecodeRequest(os.Stdin)
	if err != nil {
		return err
	}

	session := runtime.NewSession(req, nil)
	session.Start()
	session.UpdatePrompt(req.PromptText)
	session.Generate()

	action := protocol.Action(os.Getenv("MUNCH_STUB_ACTION"))
	if action == "" {
		action = protocol.ActionCancel
	}

	var command string
	switch action {
	case protocol.ActionCancel:
	case protocol.ActionInsert, protocol.ActionExecute:
		command = os.Getenv("MUNCH_STUB_COMMAND")
		if command == "" {
			suggestions := session.Suggestions()
			if len(suggestions) > 0 {
				command = suggestions[0].Command
			} else {
				command = req.PromptText
			}
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
