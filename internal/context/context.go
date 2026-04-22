package context

import (
	"os"
	"os/exec"
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

func CollectBootstrap() Normalized {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	return Normalized{
		CWD:     cwd,
		History: nil,
		InstalledTools: map[string]bool{
			"rg":   hasTool("rg"),
			"fd":   hasTool("fd"),
			"jq":   hasTool("jq"),
			"git":  hasTool("git"),
			"bat":  hasTool("bat"),
			"fish": hasTool("fish"),
			"zsh":  hasTool("zsh"),
		},
		Repo: RepoSummary{
			Type:   repoType(cwd),
			Branch: "",
			Dirty:  false,
		},
	}
}

func hasTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func repoType(cwd string) string {
	dir := cwd
	for {
		if _, err := os.Stat(dir + "/.git"); err == nil {
			return "git"
		}
		parent := dirParent(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func dirParent(path string) string {
	parent := path
	for len(parent) > 1 && parent[len(parent)-1] == '/' {
		parent = parent[:len(parent)-1]
	}
	idx := len(parent) - 1
	for idx >= 0 && parent[idx] != '/' {
		idx--
	}
	if idx <= 0 {
		return parent[:1]
	}
	return parent[:idx]
}
