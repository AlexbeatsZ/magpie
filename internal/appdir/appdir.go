// Package appdir owns the paths for magpie's files.
// Magpie always keeps its state beside the executable in data, caches in
// data/cache, logs in data/logs and the Windows webview in data/webview2.
// No marker, existing folder or installed configuration selects a mode.
// For a Mac app, data is beside the .app bundle; for an AppImage, it is
// beside the AppImage file. Symlinks are followed. The path is resolved
// once at startup so moving the executable aside during an update cannot
// move the running process's data.
//
// Files owned by other programs, system keychains, OS autostart and link
// registrations, and temporary files retain their owners' locations.
package appdir

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// DataName is the folder beside magpie that holds its state.
const DataName = "data"

var (
	once     sync.Once
	portable string // empty only with the explicit test override
)

// Portable is the data folder magpie keeps everything in.
func Portable() string {
	once.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			log.Panicf("appdir: finding the executable for portable data: %v", err)
		}
		portable = Resolve(exe)
	})
	return portable
}

// UseExecutable overrides the executable for tests. An empty path selects
// legacy profile paths explicitly, so sandboxed tests of existing data
// formats can keep their own HOME/XDG fixtures. Startup never selects them.
func UseExecutable(exe string) {
	once.Do(func() {})
	portable = ""
	if exe != "" {
		portable = Resolve(exe)
	}
}

// Resolve is the data folder beside exe, whether or not it exists.
// It never inspects data, marker files or a previous installation.
func Resolve(exe string) string {
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return absolute(filepath.Join(Beside(exe), DataName))
}

// Beside is the folder a portable magpie run from exe keeps its data folder
// in: the executable's own, the one holding the .app on a Mac, the one
// holding the .AppImage on Linux. APPIMAGE counts only for the AppImage
// magpie runs from (exe under its mount, APPDIR): every program an
// AppImage starts inherits both, a magpie started from an AppImage
// terminal too.
func Beside(exe string) string {
	if img := os.Getenv("APPIMAGE"); img != "" && runtime.GOOS == "linux" && within(exe, os.Getenv("APPDIR")) {
		return filepath.Dir(img)
	}
	dir := filepath.Dir(exe)
	// …/Magpie.app/Contents/MacOS/magpie
	if filepath.Base(dir) == "MacOS" && filepath.Base(filepath.Dir(dir)) == "Contents" {
		if app := filepath.Dir(filepath.Dir(dir)); strings.EqualFold(filepath.Ext(app), ".app") {
			return filepath.Dir(app)
		}
	}
	return dir
}

// within says whether path lies in the absolute folder dir.
func within(path, dir string) bool {
	if !filepath.IsAbs(dir) {
		return false
	}
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func absolute(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

// Config is the folder magpie keeps its settings, providers, usage and the
// rest of its own state in.
func Config() string {
	if p := Portable(); p != "" {
		return p
	}
	if x := os.Getenv("XDG_CONFIG_HOME"); rooted(x) {
		return Join(x, "magpie")
	}
	return Join(Join(mustHome(), ".config"), "magpie")
}

// Redirected is the folder magpie keeps its files in when it is opened
// from the Dock, the Start menu or a desktop launcher (~/.config/magpie),
// while this magpie reads them from another that XDG_CONFIG_HOME, set in
// the shell it was run from, names; "" when they are the same, or magpie
// is portable. A magpie run in such a shell sees none of what the app's
// window shows.
func Redirected() string {
	if Portable() != "" || os.Getenv("XDG_CONFIG_HOME") == "" {
		return ""
	}
	h, err := Home()
	if err != nil {
		return ""
	}
	if def := filepath.Join(h, ".config", "magpie"); filepath.Clean(def) != filepath.Clean(Config()) {
		return def
	}
	return ""
}

// Cache is the folder for what magpie can fetch again (the models.dev
// catalog, exchange rates, market lists).
func Cache() string {
	if p := Portable(); p != "" {
		return Join(p, "cache")
	}
	if x := os.Getenv("XDG_CACHE_HOME"); rooted(x) {
		return Join(x, "magpie")
	}
	return Join(Join(mustHome(), ".cache"), "magpie")
}

// Logs is the folder the desktop app writes its log in, where each system
// keeps an app's logs: ~/Library/Logs/magpie on a Mac (beside the launch
// agent's serve.log, `magpie daemon install`), %LOCALAPPDATA%\magpie\logs
// on Windows, and $XDG_STATE_HOME/magpie (~/.local/state/magpie) elsewhere;
// portable, data/logs. A user asked for magpie's log finds it there
// (rayem1w on X: the app opened from the Finder had written none).
func Logs() string {
	return logsOn(runtime.GOOS)
}

// logsOn is Logs as it is on goos, for tests of every system's folder.
func logsOn(goos string) string {
	if p := Portable(); p != "" {
		return filepath.Join(p, "logs")
	}
	switch goos {
	case "darwin":
		return filepath.Join(mustHome(), "Library", "Logs", "magpie")
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if !rooted(local) {
			local = filepath.Join(mustHome(), "AppData", "Local")
		}
		return filepath.Join(local, "magpie", "logs")
	}
	if x := os.Getenv("XDG_STATE_HOME"); rooted(x) {
		return filepath.Join(x, "magpie")
	}
	return filepath.Join(mustHome(), ".local", "state", "magpie")
}

// WebView is the folder the Windows webview (WebView2) keeps its profile
// in — its cache, cookies and the pages' storage — or "" for its own
// default. Installed, that default is %APPDATA%\<exe name>, as it always
// was, so nothing kept there is lost; portable, it is data\webview2, so a
// portable magpie leaves no magpie.exe folder in the user's AppData (#508).
// A Windows magpie installed from a download under another name
// (update.Settle) keeps the profile it had there: the WebViewNote beside
// it names that folder.
func WebView() string {
	if p := Portable(); p != "" {
		return filepath.Join(p, "webview2")
	}
	if runtime.GOOS != "windows" {
		return ""
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return noted(filepath.Join(filepath.Dir(exe), WebViewNote))
}

// WebViewNote is the file beside an installed Windows magpie naming the
// webview profile it keeps, the one it had where it was downloaded.
const WebViewNote = "webview-profile"

// noted is the absolute folder the note at path names, "" when there is
// no such note.
func noted(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if p := strings.TrimSpace(string(b)); filepath.IsAbs(p) {
		return p
	}
	return ""
}

// SystemCache is the cache folder the OS names for magpie
// (~/Library/Caches/magpie on a Mac, %LocalAppData%\magpie on Windows),
// which a few caches have always used; portable, it is Cache.
func SystemCache() (string, error) {
	if p := Portable(); p != "" {
		return filepath.Join(p, "cache"), nil
	}
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "magpie"), nil
}
