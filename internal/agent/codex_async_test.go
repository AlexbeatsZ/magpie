package agent

import (
	"testing"
)

// The versions that take the table: 0.163.0-alpha.8 on.
func TestCodexAsyncBlocked(t *testing.T) {
	for v, blocked := range map[string]bool{
		"0.160.0": true, "0.162.1": true, "0.163.0-alpha.5": true, "": true,
		"0.163.0-alpha.8": false, "0.163.0-alpha.10.2": false, "0.163.0": false, "0.164.0": false,
	} {
		if got := len(codexAsyncBlocked([]CodexReader{{Kind: "cli", Version: v}})) > 0; got != blocked {
			t.Errorf("%q blocked = %v, want %v", v, got, blocked)
		}
	}
	if len(codexAsyncBlocked(nil)) == 0 {
		t.Error("no Codex found counts as one that takes the table")
	}
}
