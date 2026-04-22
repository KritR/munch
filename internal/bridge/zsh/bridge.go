package zsh

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/krithikr/munch/internal/protocol"
)

func RequestFromEnv() (protocol.ShellInvocationRequest, error) {
	cursor, err := strconv.Atoi(os.Getenv("CURSOR_POSITION"))
	if err != nil {
		return protocol.ShellInvocationRequest{}, fmt.Errorf("invalid CURSOR_POSITION: %w", err)
	}

	reqID := os.Getenv("REQUEST_ID")
	if reqID == "" {
		reqID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	req := protocol.ShellInvocationRequest{
		SchemaVersion:  protocol.SchemaVersion,
		RequestID:      reqID,
		Shell:          protocol.ShellZsh,
		OriginalBuffer: os.Getenv("ORIGINAL_BUFFER"),
		PromptText:     os.Getenv("PROMPT_TEXT"),
		CursorPosition: cursor,
	}
	return req, req.Validate()
}

func ResponseAssignments(resp protocol.ShellInvocationResponse) string {
	var b strings.Builder
	b.WriteString("MUNCH_ACTION=")
	b.WriteString(shellQuote(string(resp.Action)))
	b.WriteString("\n")
	b.WriteString("MUNCH_COMMAND=")
	b.WriteString(shellQuote(resp.Command))
	b.WriteString("\n")
	return b.String()
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
