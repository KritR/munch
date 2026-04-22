package context

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestParseZshHistory(t *testing.T) {
	raw := []byte(": 1713734000:0;git status\nrg TODO .\n")
	got := parseZshHistory(raw)
	if len(got) != 2 {
		t.Fatalf("unexpected history length: %d", len(got))
	}
	if got[0] != "git status" || got[1] != "rg TODO ." {
		t.Fatalf("unexpected parsed zsh history: %#v", got)
	}
}

func TestParseFishHistory(t *testing.T) {
	raw := []byte("- cmd: git status\n  when: 1713734000\n- cmd: rg TODO .\n")
	got := parseFishHistory(raw)
	if len(got) != 2 {
		t.Fatalf("unexpected history length: %d", len(got))
	}
	if got[0] != "git status" || got[1] != "rg TODO ." {
		t.Fatalf("unexpected parsed fish history: %#v", got)
	}
}

func TestCollectorUsesHistoryFileFromEnv(t *testing.T) {
	collector := Collector{
		LookPath: func(name string) (string, error) { return "/usr/bin/" + name, nil },
		Environ: func() []string {
			return []string{"HOME=/tmp/home", "HISTFILE=/tmp/custom_history"}
		},
		Getwd: func() (string, error) { return "/tmp/project", nil },
		ReadFile: func(path string) ([]byte, error) {
			if path != "/tmp/custom_history" {
				t.Fatalf("unexpected history path: %s", path)
			}
			return []byte(": 1:0;git status\n"), nil
		},
		Command: func(name string, arg ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "printf false")
		},
	}

	ctx := collector.Collect(protocol.ShellZsh)
	if len(ctx.History) != 1 || ctx.History[0] != "git status" {
		t.Fatalf("unexpected history: %#v", ctx.History)
	}
}

func TestLastEntries(t *testing.T) {
	got := lastEntries([]string{"1", "2", "3"}, 2)
	if strings.Join(got, ",") != "2,3" {
		t.Fatalf("unexpected last entries: %#v", got)
	}
}
