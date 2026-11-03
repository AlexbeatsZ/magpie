package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// lodyMarks are entries Lody makes in its home folder (~/.lody, or
// ~/.lody-oss for its OSS build) once it has run: a folder of that name
// another tool made, with none of them in it, isn't Lody.
var lodyMarks = []string{"workspace-catalog.json", "loro-repo", "agent-binaries", "code-collab-v2"}

// lodyNames are the names Lody's builds go by: its app bundle on a Mac and
// its data folder under the OS's config folder (Electron's userData:
// ~/Library/Application Support/Lody, %APPDATA%\Lody, ~/.config/Lody).
var lodyNames = []string{"Lody", "Lody OSS", "Lody Nightly"}

// lodyBundle reports whether an app bundle in folders is one of Lody's:
// ai.lody.desktop (or another ai.lody. build), or a plist that isn't XML,
// which says nothing either way. A Lody.app of another vendor's isn't.
func lodyBundle(folders []string) bool {
	for _, dir := range folders {
		for _, name := range lodyNames {
			b, err := os.ReadFile(filepath.Join(dir, name+".app", "Contents", "Info.plist"))
			if err != nil {
				continue
			}
			if id := plistIDRe.FindSubmatch(b); id == nil || strings.HasPrefix(strings.TrimSpace(string(id[1])), "ai.lody.") {
				return true
			}
		}
	}
	return false
}

func lody(home string) *Agent {
	homes := []string{filepath.Join(home, ".lody"), filepath.Join(home, ".lody-oss")}
	var dirs []string
	if cfg, err := os.UserConfigDir(); err == nil {
		for _, name := range lodyNames {
			dirs = append(dirs, filepath.Join(cfg, name))
		}
	}
	return &Agent{
		ID: "lody", Name: "Lody", Icon: "lody", UA: []string{"lody"},
		detect: func() bool {
			for _, dir := range homes {
				for _, m := range lodyMarks {
					if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
						return true
					}
				}
			}
			for _, dir := range dirs {
				if isDir(dir) {
					return true
				}
			}
			return runtime.GOOS == "darwin" && lodyBundle(macAppFolders(home))
		},
	}
}
