package safety

import (
	"strings"

	"github.com/krithikr/munch/internal/protocol"
)

type Level string

const (
	LevelLow      Level = "low"
	LevelBalanced Level = "balanced"
	LevelStrict   Level = "strict"
)

func Apply(level Level, suggestions []protocol.Suggestion) []protocol.Suggestion {
	out := make([]protocol.Suggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		risk, reason := classify(suggestion.Command, suggestion.Risk)
		suggestion.Risk = risk
		suggestion.RequiresConfirmation = requiresConfirmation(level, risk, suggestion.Command)
		suggestion.ConfirmationReason = ""
		if suggestion.RequiresConfirmation {
			suggestion.ConfirmationReason = reason
		}
		out = append(out, suggestion)
	}
	return out
}

func classify(command string, advisoryRisk string) (string, string) {
	cmd := strings.ToLower(command)

	switch {
	case strings.Contains(cmd, "sudo"):
		return "high", "This command uses sudo."
	case strings.Contains(cmd, "rm -rf"), strings.Contains(cmd, "rm -r"), strings.Contains(cmd, " rm "):
		return "high", "This command may delete files."
	case strings.Contains(cmd, "git reset --hard"):
		return "high", "This command may discard repository changes."
	case strings.Contains(cmd, "git clean"):
		return "high", "This command may remove untracked files."
	case strings.Contains(cmd, "dd "):
		return "high", "This command may overwrite data."
	}

	switch {
	case strings.Contains(cmd, "brew install"),
		strings.Contains(cmd, "apt install"),
		strings.Contains(cmd, "npm install -g"),
		strings.Contains(cmd, "pip install"),
		strings.Contains(cmd, "sed -i"),
		strings.Contains(cmd, "mkdir"),
		strings.Contains(cmd, "cp "),
		strings.Contains(cmd, "mv "),
		strings.Contains(cmd, "touch "),
		strings.Contains(cmd, "git add"),
		strings.Contains(cmd, "git commit"),
		strings.Contains(cmd, " >"),
		strings.Contains(cmd, ">>"):
		return "medium", "This command modifies local state."
	}

	switch advisoryRisk {
	case "high":
		return "high", "This command may have destructive effects."
	case "medium":
		return "medium", "This command may modify local state."
	default:
		return "low", ""
	}
}

func requiresConfirmation(level Level, risk string, command string) bool {
	switch level {
	case LevelLow:
		return risk == "high"
	case LevelStrict:
		if risk == "medium" || risk == "high" {
			return true
		}
		return mutates(command)
	default:
		return risk == "medium" || risk == "high"
	}
}

func mutates(command string) bool {
	cmd := strings.ToLower(command)
	switch {
	case strings.Contains(cmd, "mkdir"),
		strings.Contains(cmd, "cp "),
		strings.Contains(cmd, "mv "),
		strings.Contains(cmd, "touch "),
		strings.Contains(cmd, "git add"),
		strings.Contains(cmd, "git commit"),
		strings.Contains(cmd, "sed -i"),
		strings.Contains(cmd, " >"),
		strings.Contains(cmd, ">>"),
		strings.Contains(cmd, "sudo"),
		strings.Contains(cmd, "rm "):
		return true
	default:
		return false
	}
}
