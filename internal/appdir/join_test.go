package appdir

import (
	"path/filepath"
	"testing"
)

// magpie's folders and the files named from them are joined once, not on
// every read of a file, and a folder changed since is seen at once.
func TestFoldersJoinedOnce(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	r := root(t)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(r, "cfg"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(r, "cache"))
	t.Cleanup(func() { UseExecutable("") })
	exe := touch(t, filepath.Join(r, "app", "magpie.exe"))
	UseExecutable(exe)
	data := filepath.Join(r, "app", "data")

	for name, f := range map[string]func(){
		"Config":   func() { Config() },
		"Cache":    func() { Cache() },
		"JoinFile": func() { JoinFile(Cache(), "openai", ".json") },
	} {
		f()
		if n := testing.AllocsPerRun(20, f); n > 0 {
			t.Errorf("%s with the folders unchanged: %.0f allocations", name, n)
		}
	}
	if got, want := JoinFile(Join(Cache(), "models"), "openai", ".json"), filepath.Join(data, "cache", "models", "openai.json"); got != want {
		t.Errorf("JoinFile = %q, want %q", got, want)
	}

	t.Setenv("XDG_CACHE_HOME", filepath.Join(r, "other"))
	if got, want := Cache(), filepath.Join(data, "cache"); got != want {
		t.Errorf("XDG changed the portable cache: %q, want %q", got, want)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv(homeVar(), filepath.Join(r, "home"))
	if got := Config(); got != data {
		t.Errorf("HOME changed the portable config: %q, want %q", got, data)
	}
	other := touch(t, filepath.Join(r, "other", "magpie.exe"))
	UseExecutable(other)
	if got, want := Cache(), filepath.Join(r, "other", "data", "cache"); got != want {
		t.Errorf("an executable changed since: %q, want %q", got, want)
	}
}
