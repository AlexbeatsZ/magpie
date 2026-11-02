package appdir

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestDefaultPortableWithoutMarkers(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	r := root(t)
	h := filepath.Join(r, "home")
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(r, "cfg"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(r, "cache"))
	touch(t, filepath.Join(r, "cfg", "magpie", "providers.json"))
	for _, dir := range []string{
		filepath.Join(r, "app"), h,
		filepath.Join(h, "Downloads"), filepath.Join(h, "Desktop"),
		filepath.Join(h, "Documents"), filepath.Join(r, "Applications"),
		filepath.Join(r, "D", "Downloads"),
	} {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			exe := touch(t, filepath.Join(dir, "magpie.exe"))
			data := filepath.Join(dir, "data")
			if got := Resolve(exe); got != data {
				t.Fatalf("fresh executable: data = %q, want %q without a marker or data folder", got, data)
			}
			if _, err := os.Stat(data); !os.IsNotExist(err) {
				t.Fatalf("resolving the directory created data: %v", err)
			}
			// The old installation and unrelated contents never select a mode.
			touch(t, filepath.Join(data, "report.json"))
			if got := Resolve(exe); got != data {
				t.Fatalf("existing data: got %q, want %q", got, data)
			}
		})
	}
}

func TestDefaultPortableStartup(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	t.Cleanup(func() { UseExecutable("") })
	// Exercise normal startup rather than the explicit test override.
	once = sync.Once{}
	portable = ""
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data := Resolve(exe)
	if data == "" {
		t.Fatal("startup must have a data directory without pre-existing files")
	}
	for name, pair := range map[string][2]string{
		"portable": {Portable(), data},
		"config":   {Config(), data},
		"cache":    {Cache(), filepath.Join(data, "cache")},
		"logs":     {Logs(), filepath.Join(data, "logs")},
		"webview":  {WebView(), filepath.Join(data, "webview2")},
	} {
		if pair[0] != pair[1] {
			t.Errorf("%s = %q, want %q", name, pair[0], pair[1])
		}
	}
	if got, err := SystemCache(); err != nil || got != filepath.Join(data, "cache") {
		t.Errorf("system cache = %q, %v", got, err)
	}
	if got := Redirected(); got != "" {
		t.Errorf("portable startup reports a profile redirect: %q", got)
	}
}
