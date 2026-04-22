package prompting

import (
	"strings"
	"testing"

	munchctx "github.com/krithikr/munch/internal/context"
)

func TestRenderUserPromptIncludesPromptAndRepoSummary(t *testing.T) {
	ctx := munchctx.Normalized{
		CWD:     "/tmp/project",
		History: []string{"git status"},
		InstalledTools: map[string]bool{
			"rg": true,
		},
		Repo: munchctx.RepoSummary{
			Type:   "git",
			Branch: "main",
			Dirty:  true,
		},
	}

	rendered := RenderUserPrompt(ctx, "find logs", 3)
	for _, want := range []string{"find logs", "/tmp/project", "git status", "branch: main", "Requested suggestion count"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered context missing %q", want)
		}
	}
}
