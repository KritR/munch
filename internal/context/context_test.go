package context

import "testing"

func TestDirParent(t *testing.T) {
	tests := map[string]string{
		"/tmp/project": "/tmp",
		"/tmp":         "/",
		"/":            "/",
	}

	for input, want := range tests {
		if got := dirParent(input); got != want {
			t.Fatalf("dirParent(%q) = %q, want %q", input, got, want)
		}
	}
}
