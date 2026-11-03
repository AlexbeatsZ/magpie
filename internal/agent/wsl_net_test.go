package agent

import (
	"strings"
	"testing"
)

// The probe takes WSL's networking mode from wslinfo; an empty answer or
// an old wslinfo's usage text says nothing.
func TestParseProbeNet(t *testing.T) {
	for out, want := range map[string]string{
		"home:/home/me\nnet:mirrored\n":                    "mirrored",
		"home:/home/me\nnet:NAT\r\n":                       "nat",
		"home:/home/me\nnet:\n":                            "",
		"home:/home/me\nnet:Usage: wslinfo [Options...]\n": "",
		"home:/home/me\n":                                  "",
	} {
		if d := parseProbe("U", out); d == nil || d.Net != want {
			t.Errorf("%q: %+v", out, d)
		}
	}
	if !strings.Contains(wslProbeScript, "wslinfo --networking-mode") {
		t.Error("probe doesn't ask wslinfo")
	}
}
