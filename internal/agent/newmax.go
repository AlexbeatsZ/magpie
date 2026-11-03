package agent

// NewMax (newmax.cc, 牛马AI; jwaitq on Discord) is a desktop agent app on
// the Claude Agent SDK. It keeps its settings in its own SQLite database,
// newmax.db in its data folder: $NEWMAX_USER_DATA_DIR, else ~/.newmax on
// every system (configureAppDataPaths in its main bundle, v1.1.22). The
// row `app` of its table settings (key TEXT PRIMARY KEY, value TEXT NOT
// NULL) is JSON.stringify of all its settings, and their providers are a
// list in it, keys in the clear:
//
//	{…, "providerId":"default", "model":"default",
//	 "providers":[…, {"id":"magpie-gateway","name":"magpie",
//	   "baseUrl":"http://127.0.0.1:3425","apiKey":"magpie-newmax",
//	   "apiKeys":["magpie-newmax"],"apiKeyRotationEnabled":false,
//	   "apiKeyField":"ANTHROPIC_AUTH_TOKEN","apiFormat":"anthropic",
//	   "models":[…],"enabledModels":[…],"enabled":true,"canDelete":true}],
//	 "providerOrder":["default","magpie-gateway"], …}
//
// That entry is the one NewMax's own "import from magpie" writes for a
// provider of providers.json (mapProvider, applyCcSwitchProviderImport):
// an Anthropic base URL without /v1 (it posts to baseUrl + /v1/messages,
// through the Claude Code its SDK runs, as ANTHROPIC_BASE_URL and
// ANTHROPIC_AUTH_TOKEN), at most 20 models (MAX_MODELS: a primary and 19
// back-ups; it lists enabledModels, the first 20, in its model picker), and
// the id added to providerOrder. A provider is listed in the model picker
// whether or not providerOrder has it, so the model is picked there and the
// field is whether magpie's models are in it (ListsModels). The id is
// magpie's own: NewMax's built-in "Magpie" preset is "magpie", and the ones
// its import makes are "magpie-<id>%3A<protocol>", so neither is touched.
//
// NewMax reads the row when it starts and writes it whole from what it
// holds whenever a setting changes, so a change magpie makes while it runs
// is lost the next time it saves: magpie refuses one then, where it can
// tell (Running can't on Windows, where the notice says so instead). The
// row is edited in place, in one transaction: magpie's entry and its id in
// providerOrder go in and out, and the rest of the row stays byte for byte.
// The user's name for magpie's entry, its on/off and the fields NewMax adds
// to it (capabilities it tested) stay through a sync.

import (
	"path/filepath"

	"github.com/yetone/magpie/internal/appdir"
	_ "github.com/yetone/magpie/internal/sqlitedb"
)

// newMaxApp is NewMax's macOS bundle.
var newMaxApp = desktopApp{id: "cc.newmax.desktop", names: []string{"NewMax.app"}}

func newMax(home string) *Agent {
	dir := newMaxDir(home)
	path := filepath.Join(dir, "newmax.db")

	return &Agent{
		ID: "newmax", Name: "NewMax", Icon: "generic", Aliases: []string{"newmax-ai", "newmaxai"},
		Dir: dir, Path: path,
		detect: func() bool { return isDir(dir) || macBundle(newMaxApp) },
	}
}

// newMaxDir is NewMax's data folder: $NEWMAX_USER_DATA_DIR when it names
// one, else ~/.newmax, on every system.
func newMaxDir(home string) string {
	if d := appdir.Getenv("NEWMAX_USER_DATA_DIR"); d != "" && filepath.IsAbs(d) {
		return d
	}
	return filepath.Join(home, ".newmax")
}
