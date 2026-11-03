package agent

import (
	"testing"

	"github.com/yetone/magpie/internal/proc"
)

// TestCodexWindowsStoreAppVersion: on Windows the Codex app is the Microsoft
// Store's OpenAI.Codex package (#1334, laacmilan: C:\Program Files\
// WindowsApps\OpenAI.Codex_26.1002.7124.0_x64__2p2nqsd0c76g0), read from the
// newest main package registered for the user.
func TestCodexWindowsStoreAppVersion(t *testing.T) {
	was, wasStore := appFolders, windowsStorePackages
	appFolders = func() []string { return nil }
	windowsStorePackages = func() []string {
		return []string{
			"OpenAI.Codex_not-a-version_x64__2p2nqsd0c76g0",
			"OpenAI.Codex_26.928.2636.0_x64__2p2nqsd0c76g0",
			"OpenAI.Codex_26.1002.7124.0_arm64__2p2nqsd0c76g0",
			"OpenAI.Codex_99.0.0.0_x64__wrong-publisher",
			// a resource package left behind, as the Windows box's registry
			// keeps Clipchamp's after its main package moved on
			"OpenAI.Codex_27.0.0.0_neutral_split.scale-125_2p2nqsd0c76g0",
		}
	}
	t.Cleanup(func() { appFolders, windowsStorePackages = was, wasStore })

	codex := &Agent{ID: "codex", Name: "Codex", Bin: "codex"}
	if v := codex.AppVersion(); v != "26.1002.7124.0" {
		t.Fatalf("Codex Store app: %q", v)
	}
	if !codex.HasApp() {
		t.Fatal("Codex Store app not detected")
	}
	if v := (&Agent{ID: "codex@wsl:Ubuntu", WSL: "Ubuntu", Bin: "codex"}).AppVersion(); v != "" {
		t.Fatalf("WSL twin: %q", v)
	}
}

// A test binary testenv isolated never reads the machine's Store apps, in
// whatever package it is: on a Windows machine with the Codex app installed,
// library's tests saw it and codex's CLI was never gone (#2254, ggbdpq).
func TestStorePackagesNotTheMachinesInTests(t *testing.T) {
	if !proc.Isolated() {
		t.Fatal("agent's tests run outside testenv's home")
	}
	if got := storePackages(); len(got) != 0 {
		t.Fatalf("a test read this machine's Store packages: %d of them, %v", len(got), got[:min(len(got), 3)])
	}
}
