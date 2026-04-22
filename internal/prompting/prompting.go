package prompting

import (
	"fmt"
	"sort"
	"strings"

	munchctx "github.com/krithikr/munch/internal/context"
)

const CanonicalSystemPrompt = `You generate shell command suggestions for an interactive shell widget.

Your job is to return a small set of useful shell commands that help the user accomplish the requested task in their current environment.

Follow these rules:

1. Return only a valid JSON object matching the requested output shape.
2. Do not include any text before or after the JSON object.
3. Generate shell commands, not conversational answers.
4. Prefer commands that use tools known to be installed.
5. Prefer modern tools when they are available and appropriate.
6. Prefer read-only commands unless the user's task clearly implies mutation.
7. When useful, include a fallback command that uses more standard tooling.
8. Keep descriptions short and practical.
9. Keep assumptions short and include them only when they materially help the user understand the command.
10. Include an advisory risk classification for each suggestion using only: low, medium, high.
11. Do not decide final confirmation policy.
12. Do not output extra explanation, markdown, or prose outside the JSON object.

Output a JSON object with this shape:
{
  "suggestions": [
    {
      "command": "string",
      "description": "string",
      "risk": "low|medium|high",
      "assumptions": ["string"],
      "uses_tools": ["string"],
      "confidence": 0.0
    }
  ]
}

Return up to the requested number of suggestions. Prefer a smaller number of high-quality suggestions over padding with weak ones.`

func RenderContext(ctx munchctx.Normalized, promptText string, suggestionCount int) string {
	var b strings.Builder
	b.WriteString("System prompt:\n")
	b.WriteString(CanonicalSystemPrompt)
	b.WriteString("\n\nTask:\n")
	b.WriteString(promptText)
	b.WriteString("\n\nContext:\n")
	b.WriteString(fmt.Sprintf("- cwd: %s\n", ctx.CWD))
	b.WriteString("- history:\n")
	if len(ctx.History) == 0 {
		b.WriteString("  - <none>\n")
	} else {
		for _, entry := range ctx.History {
			b.WriteString(fmt.Sprintf("  - %s\n", entry))
		}
	}
	b.WriteString("- installed tools:\n")
	tools := sortedToolKeys(ctx.InstalledTools)
	for _, tool := range tools {
		b.WriteString(fmt.Sprintf("  - %s: %t\n", tool, ctx.InstalledTools[tool]))
	}
	b.WriteString("- repo summary:\n")
	b.WriteString(fmt.Sprintf("  - type: %s\n", ctx.Repo.Type))
	b.WriteString(fmt.Sprintf("  - branch: %s\n", ctx.Repo.Branch))
	b.WriteString(fmt.Sprintf("  - dirty: %t\n", ctx.Repo.Dirty))
	b.WriteString("\nRequested suggestion count:\n")
	b.WriteString(fmt.Sprintf("%d\n", suggestionCount))
	return b.String()
}

func sortedToolKeys(tools map[string]bool) []string {
	keys := make([]string, 0, len(tools))
	for k := range tools {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
