package agent

import (
	"testing"

	"github.com/yetone/magpie/internal/usage"
)

func TestStepCodeRegistered(t *testing.T) {
	byID, err := Find("stepcode")
	if err != nil || byID == nil {
		t.Fatalf("Find(\"stepcode\") failed: %v", err)
	}
	if byID.ID != "stepcode" {
		t.Fatalf("Find(\"stepcode\").ID = %q, want \"stepcode\"", byID.ID)
	}

	byBin, err := Find("step")
	if err != nil || byBin == nil {
		t.Fatalf("Find(\"step\") failed: %v", err)
	}
	if byBin.ID != "stepcode" {
		t.Fatalf("Find(\"step\").ID = %q, want \"stepcode\"", byBin.ID)
	}
	// StepCode 0.1.3's requests say "step (darwin 25.2.0; arm64)", read off
	// a real turn through the gateway; they count as StepCode's
	for _, ua := range []string{"step (darwin 25.2.0; arm64)", "step (linux 6.8.0; x64)"} {
		if got := usage.AgentOf(ua); got != "stepcode" {
			t.Fatalf("usage.AgentOf(%q) = %q, want stepcode", ua, got)
		}
	}
}
