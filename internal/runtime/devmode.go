package runtime

import "github.com/krithikr/munch/internal/protocol"

type DevMode string

const (
	DevModeNone            DevMode = "none"
	DevModeAutoInsertFirst DevMode = "auto-insert-first"
)

func ResolveDevAction(mode DevMode, suggestions []protocol.Suggestion, fallback string) (protocol.Action, string, bool) {
	switch mode {
	case DevModeAutoInsertFirst:
		command := fallback
		if len(suggestions) > 0 && suggestions[0].Command != "" {
			command = suggestions[0].Command
		}
		return protocol.ActionInsert, command, true
	default:
		return "", "", false
	}
}
