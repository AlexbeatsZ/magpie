package agent

import (
	"strings"
	"testing"
)

// The WSL probe asks after OmO too.
func TestWSLProbeFindsOmo(t *testing.T) {
	for _, want := range []string{`[ -d "$HOME/.omo" ] && echo dir:.omo`, `p=$(command -v omo 2>/dev/null || wslbin omo) && echo "bin:omo $p"`} {
		if !strings.Contains(wslProbeScript, want) {
			t.Errorf("probe lacks %q", want)
		}
	}
	if d := parseProbe("U", "home:/home/me\nbin:omo\n"); d == nil || !wslFound(*d) {
		t.Fatalf("%+v", d)
	}
}
