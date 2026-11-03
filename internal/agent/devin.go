package agent

// Devin keeps its settings in ~/.config/devin/config.json (%APPDATA%\devin on
// Windows), the model under agent.model. Devin has no endpoint magpie can
// stand in for — its models are its own — so the options are the families
// `devin models list --format json` reports, not the magpie catalog, offered
// as the provider offers them: each family one model (and its fast run one
// more, claude-opus-5-5-fast), with an effort beside it. Devin's config has
// no effort of its own — a variant id is the effort it runs at — so the pair
// is written as the one id: a family's id with no effort (it follows the
// family's newest), its variant at the effort picked (claude-opus-5-5-high,
// claude-opus-5-5-high-fast) otherwise, and read back the same way.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

func devin(_, cfg string) *Agent {
	dir := filepath.Join(cfg, "devin")
	if runtime.GOOS == "windows" {
		if app := appdir.Getenv("APPDATA"); app != "" {
			dir = filepath.Join(app, "devin")
		}
	}
	path := filepath.Join(dir, "config.json")

	return &Agent{
		ID: "devin", Name: "Devin", Icon: "devin",
		Bin: "devin", Dir: dir, Path: path,
	}
}
