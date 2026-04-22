package main

import (
	"strings"
	"testing"
)

func TestRunUnsupportedMode(t *testing.T) {
	err := run([]string{"--mode", "bogus"})
	if err == nil {
		t.Fatal("expected unsupported mode error")
	}
	if !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}
