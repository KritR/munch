package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krithikr/munch/internal/protocol"
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

	action := protocol.Action(os.Getenv("MUNCH_STUB_ACTION"))
	if action == "" {
		action = protocol.ActionCancel
	}

	resp := protocol.ShellInvocationResponse{
		SchemaVersion: protocol.SchemaVersion,
		RequestID:     req.RequestID,
		Action:        action,
	}

	switch action {
	case protocol.ActionCancel:
		// No-op.
	case protocol.ActionInsert, protocol.ActionExecute:
		command := os.Getenv("MUNCH_STUB_COMMAND")
		if command == "" {
			command = req.PromptText
		}
		resp.Command = command
	default:
		return fmt.Errorf("unsupported stub action: %s", action)
	}

	return protocol.EncodeResponse(os.Stdout, resp)
}
