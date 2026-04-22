package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/krithikr/munch/internal/protocol"
)

type Selection struct {
	Action  protocol.Action
	Command string
}

func SelectSuggestion(prompt string, suggestions []protocol.Suggestion) (Selection, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return Selection{}, err
	}
	defer tty.Close()

	return selectFromTTY(prompt, suggestions, tty, tty)
}

func selectFromTTY(prompt string, suggestions []protocol.Suggestion, input io.Reader, output io.Writer) (Selection, error) {
	if len(suggestions) == 0 {
		_, _ = fmt.Fprintf(output, "\n[munch] No suggestions for: %s\n", prompt)
		return Selection{Action: protocol.ActionCancel}, nil
	}

	_, _ = fmt.Fprintf(output, "\n[munch] Prompt: %s\n", prompt)
	for i, suggestion := range suggestions {
		_, _ = fmt.Fprintf(output, "  %d. %s\n     %s\n", i+1, suggestion.Command, suggestion.Description)
	}
	_, _ = fmt.Fprint(output, "[munch] Choose suggestion number or press Enter to cancel: ")

	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return Selection{}, err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return Selection{Action: protocol.ActionCancel}, nil
	}

	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(suggestions) {
		_, _ = fmt.Fprintf(output, "[munch] Invalid selection: %s\n", line)
		return Selection{Action: protocol.ActionCancel}, nil
	}

	return Selection{
		Action:  protocol.ActionInsert,
		Command: suggestions[idx-1].Command,
	}, nil
}
