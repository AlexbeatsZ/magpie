package agent

import (
	"strings"
	"testing"
)

// The probe prints a named omp profile as profile:omp:<name>, and only
// real directories under ~/.omp/profiles count. A profile row's place
// takes omp's version, which the probe keeps under the bare omp kind.
func TestParseProbeOmpProfiles(t *testing.T) {
	d := parseProbe("U", "home:/home/me\ndir:.omp\nver:omp omp/16.5.1\n"+
		"profile:omp:work\nprofile:omp:Bad\nprofile:omp:work.\nprofile:pi:x\nprofile:omp:work\n")
	if d == nil {
		t.Fatal("no distro")
	}
	if len(d.OmpProfiles) != 1 || d.OmpProfiles[0] != "work" {
		t.Fatalf("profiles %v", d.OmpProfiles)
	}
	if v := d.place("omp#work@wsl:U").version; v != "16.5.1" {
		t.Fatalf("profile place version %q", v)
	}
	if !strings.Contains(wslProbeScript, `for d in "$HOME"/.omp/profiles/*`) ||
		!strings.Contains(wslProbeScript, `echo "profile:omp:$(basename "$d")"`) {
		t.Fatalf("the probe doesn't list profiles: %s", wslProbeScript)
	}
}
