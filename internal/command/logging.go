package command

import (
	"io"
	"log/slog"
	"os"
)

func InitLogger() {
	var out io.Writer = io.Discard
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
