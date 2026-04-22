package context

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krithikr/munch/internal/protocol"
)

type RepoSummary struct {
	Type   string
	Branch string
	Dirty  bool
}

type Normalized struct {
	CWD            string
	History        []string
	InstalledTools map[string]bool
	Repo           RepoSummary
}

type Collector struct {
	LookPath func(string) (string, error)
	Environ  func() []string
	Getwd    func() (string, error)
	ReadFile func(string) ([]byte, error)
	Command  func(name string, arg ...string) *exec.Cmd
}

func NewCollector() Collector {
	return Collector{
		LookPath: exec.LookPath,
		Environ:  os.Environ,
		Getwd:    os.Getwd,
		ReadFile: os.ReadFile,
		Command:  exec.Command,
	}
}

func CollectBootstrap() Normalized {
	return NewCollector().Collect(protocol.ShellZsh)
}

func (c Collector) Collect(shell protocol.Shell) Normalized {
	cwd, err := c.Getwd()
	if err != nil {
		cwd = "."
	}

	return Normalized{
		CWD:     cwd,
		History: c.collectHistory(shell),
		InstalledTools: map[string]bool{
			"rg":   c.hasTool("rg"),
			"fd":   c.hasTool("fd"),
			"jq":   c.hasTool("jq"),
			"git":  c.hasTool("git"),
			"bat":  c.hasTool("bat"),
			"fish": c.hasTool("fish"),
			"zsh":  c.hasTool("zsh"),
		},
		Repo: c.collectRepoSummary(cwd),
	}
}

func (c Collector) hasTool(name string) bool {
	_, err := c.LookPath(name)
	return err == nil
}

func (c Collector) collectRepoSummary(cwd string) RepoSummary {
	if !c.hasTool("git") {
		return RepoSummary{}
	}

	cmd := c.Command("git", "-C", cwd, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return RepoSummary{}
	}

	repo := RepoSummary{Type: "git"}

	branchCmd := c.Command("git", "-C", cwd, "branch", "--show-current")
	if branchOut, err := branchCmd.Output(); err == nil {
		repo.Branch = strings.TrimSpace(string(branchOut))
	}

	dirtyCmd := c.Command("git", "-C", cwd, "status", "--porcelain")
	if dirtyOut, err := dirtyCmd.Output(); err == nil {
		repo.Dirty = strings.TrimSpace(string(dirtyOut)) != ""
	}

	return repo
}

func (c Collector) collectHistory(shell protocol.Shell) []string {
	path := c.historyPath(shell)
	if path == "" {
		return nil
	}

	raw, err := c.ReadFile(path)
	if err != nil {
		return nil
	}

	switch shell {
	case protocol.ShellFish:
		return lastEntries(parseFishHistory(raw), 10)
	default:
		return lastEntries(parseZshHistory(raw), 10)
	}
}

func (c Collector) historyPath(shell protocol.Shell) string {
	env := envMap(c.Environ())
	home := env["HOME"]

	switch shell {
	case protocol.ShellFish:
		if xdg := env["XDG_CONFIG_HOME"]; xdg != "" {
			return filepath.Join(xdg, "fish", "fish_history")
		}
		if home != "" {
			return filepath.Join(home, ".local", "share", "fish", "fish_history")
		}
	case protocol.ShellZsh:
		if hist := env["HISTFILE"]; hist != "" {
			return hist
		}
		if home != "" {
			return filepath.Join(home, ".zsh_history")
		}
	}

	return ""
}

func envMap(items []string) map[string]string {
	out := make(map[string]string, len(items))
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			out[key] = value
		}
	}
	return out
}

func parseZshHistory(raw []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	var entries []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ": ") {
			if idx := strings.Index(line, ";"); idx >= 0 && idx+1 < len(line) {
				line = line[idx+1:]
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries
}

func parseFishHistory(raw []byte) []string {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	var entries []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- cmd:") {
			entry := strings.TrimSpace(strings.TrimPrefix(line, "- cmd:"))
			if entry != "" {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

func lastEntries(entries []string, limit int) []string {
	if len(entries) <= limit {
		return entries
	}
	return append([]string(nil), entries[len(entries)-limit:]...)
}
