package agent

import (
	"testing"
)

// WSL puts Windows' PATH on the distro's, so `command -v pi` there finds
// Windows npm's pi under /mnt/c (#406). A command on one of Windows' drives
// — /mnt/<letter>/, or wherever /proc/mounts says a drive is — finds no
// agent, for every kind; one in the distro's own folders still does.
func TestWSLWindowsCommandIsNotTheDistros(t *testing.T) {
	d := parseProbe("Ubuntu-24.04", "home:/home/me\n"+
		"bin:pi /mnt/c/Users/me/AppData/Roaming/npm/pi\n"+
		"bin:codex /mnt/d/tools/npm/codex\n"+
		"bin:claude /win/c/Users/me/.local/bin/claude\n"+
		"win:/win/c\n"+
		"bin:omo /mnt/data/omo/bin/omo\n")
	if d == nil {
		t.Fatal("no distro")
	}
	for _, k := range []string{"bin:pi", "bin:codex", "bin:claude"} {
		if d.Has[k] {
			t.Errorf("%s taken from Windows: %v", k, d.Has)
		}
	}
	if !d.Has["bin:omo"] {
		t.Errorf("a distro's own /mnt/data command lost: %v", d.Has)
	}
	for _, k := range wslKinds {
		if k.id != "omo" && k.found(*d) {
			t.Errorf("%s found", k.id)
		}
	}

	// Codex installed in the distro (#251's case) is found as before, with
	// a space in its path kept
	d = parseProbe("U", "home:/home/me\nbin:codex /home/me/.nvm/versions/node/v22/bin/codex\nbin:pi /opt/my tools/pi\nwin:/mnt/c\n")
	if d == nil || !d.Has["bin:codex"] || !d.Has["bin:pi"] || !wslFound(*d) {
		t.Fatalf("%+v", d)
	}
	if !onWindowsDrive("/mnt/C/x", nil) || onWindowsDrive("/mnt/c", nil) || onWindowsDrive("/mnt/cd/x", nil) || onWindowsDrive("/usr/bin/pi", nil) {
		t.Fatal("onWindowsDrive")
	}
}
