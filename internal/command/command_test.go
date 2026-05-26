package command

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunUnsupportedMode(t *testing.T) {
	err := Run(context.Background(), []string{"munch", "--mode", "bogus"}, IO{
		Stdin:  strings.NewReader(""),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected unsupported mode error")
	}
	if !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitZshUsesDefaultBinding(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"munch", "init", "zsh"}, IO{
		Stdout: &out,
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("run init zsh: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		`: "${MUNCH_BIN:=munch}"`,
		`widget_cmd=("$MUNCH_BIN" --mode zsh-bridge)`,
		`bindkey '^G' __munch_zle_widget`,
		`__munch_bind_zle`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected init output to contain %q\n%s", want, got)
		}
	}
}

func TestInitFishUsesDefaultBinding(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"munch", "init", "fish"}, IO{
		Stdout: &out,
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("run init fish: %v", err)
	}

	got := out.String()
	for _, want := range []string{
		`set -g MUNCH_BIN munch`,
		`set -l widget_cmd "$MUNCH_BIN" --mode fish-bridge`,
		`bind \cg __munch_widget`,
		`__munch_bind`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected init output to contain %q\n%s", want, got)
		}
	}
}

func TestInitKeyOverride(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"munch", "init", "zsh", "--key", "^X^M"}, IO{
		Stdout: &out,
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("run init zsh: %v", err)
	}
	if !strings.Contains(out.String(), `bindkey '^X^M' __munch_zle_widget`) {
		t.Fatalf("expected key override in init output\n%s", out.String())
	}
}

func TestInitNoBind(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), []string{"munch", "init", "fish", "--no-bind"}, IO{
		Stdout: &out,
		Stderr: &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("run init fish: %v", err)
	}

	got := out.String()
	if strings.Contains(got, "bind \\cg __munch_widget") {
		t.Fatalf("expected no binding in init output\n%s", got)
	}
	if !strings.Contains(got, "function __munch_widget") {
		t.Fatalf("expected function definition in init output\n%s", got)
	}
}
