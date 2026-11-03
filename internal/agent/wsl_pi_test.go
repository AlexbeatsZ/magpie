package agent

import (
	"strings"
	"testing"
)

// The probe asks after Pi as after Codex, and reads what it says.
func TestWSLProbeFindsPi(t *testing.T) {
	for _, want := range []string{`[ -d "$HOME/.pi" ] && echo dir:.pi`, `p=$(command -v pi 2>/dev/null || wslbin pi) && echo "bin:pi $p"`,
		`[ -d "$HOME/.codex" ] && echo dir:.codex`, `p=$(command -v codex 2>/dev/null || wslbin codex) && echo "bin:codex $p"`} {
		if !strings.Contains(wslProbeScript, want) {
			t.Errorf("probe lacks %q:\n%s", want, wslProbeScript)
		}
	}
	d := parseProbe("U", "home:/home/me\ndir:.pi\nbin:pi\n")
	if d == nil || !d.Has["dir:.pi"] || !d.Has["bin:pi"] || !wslFound(*d) {
		t.Fatalf("%+v", d)
	}
}
