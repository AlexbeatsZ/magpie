package agent

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
)

// Zed uses ~/.config on macOS, XDG on Linux and Roaming AppData on Windows.
func zed(home, cfg string) *Agent {
	bin := os.Getenv("MAGPIE_ZED_BIN")
	if bin == "" {
		bin = "zed"
	}
	processes := zedProcessNames()
	if custom := appdir.Getenv("MAGPIE_ZED_CONFIG_DIR"); custom != "" {
		return zedAtWith(custom, bin, processes)
	}
	switch runtime.GOOS {
	case "darwin":
		cfg = filepath.Join(home, ".config")
	case "windows":
		cfg = appdir.Getenv("APPDATA")
		if cfg == "" {
			cfg = filepath.Join(home, "AppData", "Roaming")
		}
		return zedAtWith(filepath.Join(cfg, "Zed"), bin, processes)
	}
	return zedAtWith(filepath.Join(cfg, "zed"), bin, processes)
}

func zedAtWith(dir, bin string, _ []string) *Agent {
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "zed", Name: "Zed", Icon: "zed", Bin: bin, Dir: dir, Path: path,
		UA: []string{"zed"},
	}
}

func zedProcessNames() []string {
	value := os.Getenv("MAGPIE_ZED_PROCESS_NAMES")
	if value == "" {
		return []string{`(^|/)(zed|zeditor|zed-editor)( |$)`}
	}
	var out []string
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, `(^|/)`+regexp.QuoteMeta(name)+`( |$)`)
		}
	}
	if len(out) == 0 {
		return []string{`(^|/)(zed|zeditor|zed-editor)( |$)`}
	}
	return out
}
