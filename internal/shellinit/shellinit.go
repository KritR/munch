package shellinit

import (
	"embed"
	"fmt"
	"strings"
)

const (
	defaultZshKey  = "^G"
	defaultFishKey = `\cg`
)

//go:embed scripts/*
var scripts embed.FS

type Options struct {
	Key    string
	NoBind bool
}

func DefaultKey(shell string) string {
	switch shell {
	case "zsh":
		return defaultZshKey
	case "fish":
		return defaultFishKey
	default:
		return ""
	}
}

func Script(shell string, opts Options) (string, error) {
	switch shell {
	case "zsh":
		return renderZsh(opts)
	case "fish":
		return renderFish(opts)
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

func renderZsh(opts Options) (string, error) {
	script, err := readScript("scripts/munch.zsh")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(script, "{{BINDINGS}}", zshBindings(opts)), nil
}

func renderFish(opts Options) (string, error) {
	script, err := readScript("scripts/munch.fish")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(script, "{{BINDINGS}}", fishBindings(opts)), nil
}

func readScript(path string) (string, error) {
	b, err := scripts.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func zshBindings(opts Options) string {
	if opts.NoBind {
		return ""
	}
	key := opts.Key
	if key == "" {
		key = defaultZshKey
	}
	return fmt.Sprintf("  bindkey %s __munch_zle_widget", shellQuote(key))
}

func fishBindings(opts Options) string {
	if opts.NoBind {
		return ""
	}
	key := opts.Key
	if key == "" {
		key = defaultFishKey
	}
	key = strings.TrimSpace(key)
	return fmt.Sprintf("    bind %s __munch_widget\n    bind -M insert %s __munch_widget\n    bind -M default %s __munch_widget", key, key, key)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
