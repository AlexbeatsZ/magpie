package agent

import (
	"runtime"
	"slices"
	"testing"
)

// Its updates: npm's package, or for install.sh's tarball in ~/.penguin
// its own `penguin update --yes`, which Windows' Penguin refuses.
func TestPenguinUpdate(t *testing.T) {
	self := cliSpecs["penguin"].self
	got := self("/home/me/.local/bin/penguin", "/home/me/.penguin/bin/penguin")
	if runtime.GOOS == "windows" {
		if got != nil {
			t.Fatalf("windows: %q", got)
		}
	} else if !slices.Equal(got, []string{"/home/me/.local/bin/penguin", "update", "--yes"}) {
		t.Fatalf("tarball: %q", got)
	}
	if got := self("/usr/local/bin/penguin", "/usr/local/lib/node_modules/@prismshadow/penguin-cli/dist/bin.js"); got != nil {
		t.Fatalf("npm taken for the tarball: %q", got)
	}
	if !slices.Equal(cliSpecs["penguin"].npm, []string{"@prismshadow/penguin-cli"}) {
		t.Fatal(cliSpecs["penguin"].npm)
	}
	for goos, want := range map[string]string{"darwin": "curl -fsSL https://penguin.ooo/install.sh | sh", "linux": "curl -fsSL https://penguin.ooo/install.sh | sh",
		"windows": "irm https://penguin.ooo/install.ps1 | iex"} {
		if c := vendorInstall["penguin"](goos); len(c) != 1 || c[0].Command != want {
			t.Fatalf("%s: %+v", goos, c)
		}
	}
}
