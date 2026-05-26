package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/krithikr/munch/internal/command"
)

func main() {
	command.InitLogger()
	if err := command.Run(context.Background(), os.Args, command.IO{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}); err != nil {
		slog.Error("munch failed", "error", err)
		os.Exit(1)
	}
}
