package safety

import (
	"testing"

	"github.com/krithikr/munch/internal/protocol"
)

func TestApplyMarksHighRiskAsConfirmRequired(t *testing.T) {
	got := Apply(LevelBalanced, []protocol.Suggestion{
		{Command: "rm -rf build", Description: "Remove build dir"},
	})
	if got[0].Risk != "high" {
		t.Fatalf("unexpected risk: %s", got[0].Risk)
	}
	if !got[0].RequiresConfirmation {
		t.Fatal("expected confirmation requirement")
	}
	if got[0].ConfirmationReason == "" {
		t.Fatal("expected confirmation reason")
	}
}

func TestApplyKeepsLowRiskWithoutConfirmation(t *testing.T) {
	got := Apply(LevelBalanced, []protocol.Suggestion{
		{Command: "rg -n TODO .", Description: "Search TODOs"},
	})
	if got[0].Risk != "low" {
		t.Fatalf("unexpected risk: %s", got[0].Risk)
	}
	if got[0].RequiresConfirmation {
		t.Fatal("did not expect confirmation requirement")
	}
}
