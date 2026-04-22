package fish

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
		Shell:          protocol.ShellFish,
		OriginalBuffer: os.Getenv("ORIGINAL_BUFFER"),
		PromptText:     os.Getenv("PROMPT_TEXT"),
		CursorPosition: cursor,
	}
	return req, req.Validate()
}

func ResponseAssignments(resp protocol.ShellInvocationResponse) string {
	var b strings.Builder
	b.WriteString("set -g -- MUNCH_ACTION ")
	b.WriteString(fishQuote(string(resp.Action)))
	b.WriteString("; ")
	b.WriteString("set -g -- MUNCH_COMMAND ")
	b.WriteString(fishQuote(resp.Command))
	return b.String()
}

func fishQuote(s string) string {
	if s == "" {
		return "''"
	}

	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
	)
	return "'" + replacer.Replace(s) + "'"
}
