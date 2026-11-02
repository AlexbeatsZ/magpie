package update

import (
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/appdir"
)

// A fresh download runs in place without creating an installed copy or
// changing PATH, even when neither data nor .portable exists yet.
func TestFreshDownloadIsPortable(t *testing.T) {
	f := newInstallFixture(t)
	t.Setenv("APPIMAGE", "")
	appdir.UseExecutable(f.exe)
	t.Cleanup(func() { appdir.UseExecutable("") })
	portableData = appdir.Portable
	if got, want := portableData(), filepath.Join(f.downloads, "data"); got != want {
		t.Fatalf("fresh download data = %q, want %q", got, want)
	}
	if to, err := settle(f.exe); err != nil || to != "" {
		t.Fatalf("fresh download opens an installed copy %q: %v", to, err)
	}
	if forwards(f.exe) || exists(f.to) || len(f.placed) != 0 {
		t.Fatalf("fresh download installed or changed PATH: target %v, PATH %v", exists(f.to), f.placed)
	}
}
