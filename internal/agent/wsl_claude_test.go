package agent

import (
	"strings"
	"testing"
)

// The probe asks after Claude Code as after Codex and Pi, and reads what
// it says.
func TestWSLProbeFindsClaude(t *testing.T) {
	for _, want := range []string{`[ -d "$HOME/.claude" ] && echo dir:.claude`, `p=$(command -v claude 2>/dev/null || wslbin claude) && echo "bin:claude $p"`,
		`[ -d "$HOME/.pi" ] && echo dir:.pi`, `[ -d "$HOME/.codex" ] && echo dir:.codex`} {
		if !strings.Contains(wslProbeScript, want) {
			t.Errorf("probe lacks %q:\n%s", want, wslProbeScript)
		}
	}
	d := parseProbe("Debian", "home:/home/user\ndir:.claude\nbin:claude\n")
	if d == nil || !d.Has["dir:.claude"] || !d.Has["bin:claude"] || !wslFound(*d) {
		t.Fatalf("%+v", d)
	}
}
