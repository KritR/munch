package context

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

func Bootstrap() Normalized {
	return Normalized{
		CWD:     ".",
		History: nil,
		InstalledTools: map[string]bool{
			"rg":   true,
			"fd":   true,
			"jq":   false,
			"git":  true,
			"bat":  false,
			"fish": true,
			"zsh":  true,
		},
		Repo: RepoSummary{
			Type:   "git",
			Branch: "main",
			Dirty:  false,
		},
	}
}
