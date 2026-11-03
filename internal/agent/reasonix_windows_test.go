//go:build windows

package agent

import (
	"testing"
)

func TestReasonixDesktopRequiresMajorVersionTwo(t *testing.T) {
	for _, tc := range []struct {
		name, version string
		want          bool
	}{
		{"Reasonix Studio", "1.0.3", false},
		{"Reasonix Studio", "", false},
		{"Reasonix Studio 2.24.0", "2.24.0", true},
		{"Reasonix Studio", "v2.24.0", true},
		{"Reasonix Studio Host", "1.18.0", false},
		{"Other app", "2.24.0", false},
	} {
		if got := reasonixStudioRelease(tc.name, tc.version); got != tc.want {
			t.Errorf("%q version %q detected=%v", tc.name, tc.version, got)
		}
	}
}
