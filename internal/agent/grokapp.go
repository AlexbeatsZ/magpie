package agent

// Grok App (RongleCat/grok-app, Tauri) is a desktop front end for the grok
// CLI, which it starts over ACP. Shared mode, the default, uses ~/.grok,
// the same home as Grok Build. Independent mode, which Use on a custom
// provider turns on, uses agent-home. The Library writes that home's
// AGENTS.md, [mcp_servers] and skills/ (and Grok App still reads
// ~/.agents/skills). Its models are its own, though: its composer lists
// only Grok Build's official catalog
// (models_cache.json), never a [model."<id>"] table of config.toml, and a
// model of the user's choosing is a "provider channel" of its Settings ›
// Providers page, a [model.<id>] table with the App's own keys in
// agent-home/config.toml under its data folder. That file is the CLI's
// config.toml while a channel is on: [models] default naming a channel is
// what the App takes as the route, and it then starts the CLI with
// GROK_HOME at agent-home, switching itself to its independent mode
// (grok-app's providers.rs, ensure_independent_for_custom_route).
//
// So magpie adds itself as the channel "magpie", written as the App writes
// one (upsert_custom_provider): the gateway as base_url with Grok App's own
// key, chat completions, app_models a JSON list of magpie's models with each
// one's context window, images and effort levels, so Accounts › Custom
// providers lists it and the composer can offer the models and their
// efforts. It does not set [models] default. That key is the route, and
// making it "magpie" takes the App over: the App then drops its own sign-in
// and the session no longer runs. The route stays whatever the App had.
// Pressing Use on the provider in the App is what switches the route, and
// that choice is the App's. A magpie that already set the route puts the
// old one back once, on the next sync or connect. A model the user picks in
// the App's composer since is theirs and stays through a sync.

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
)

// grokAppOS is the system grokAppDir looks for the App's data on; a test
// sets it.
var grokAppOS = runtime.GOOS

// grokAppDir is the App's data folder: $GROK_APP_HOME, else where the
// directories crate's ProjectDirs("com", "grokapp", "grok-app") puts its
// data (grok-app's paths.rs app_data_root) — on a Mac ~/Library/Application
// Support/com.grokapp.grok-app, on Windows %APPDATA%\grokapp\grok-app\data,
// on Linux $XDG_DATA_HOME/grok-app. The App's README names ~/.grok-app on
// Linux and %APPDATA%\grokapp\grok-app on Windows, and the App falls back to
// ~/.grok-app (%APPDATA%\grok-app) itself, so those are looked for too; the
// first that is there wins, else the App's own default.
func grokAppDir(home, goos string) string {
	if d := strings.TrimSpace(appdir.Getenv("GROK_APP_HOME")); d != "" {
		return d
	}
	var dirs []string
	switch goos {
	case "darwin":
		dirs = []string{filepath.Join(home, "Library", "Application Support", "com.grokapp.grok-app"), filepath.Join(home, ".grok-app")}
	case "windows":
		app := appdir.Getenv("APPDATA")
		if app == "" {
			app = filepath.Join(home, "AppData", "Roaming")
		}
		dirs = []string{filepath.Join(app, "grokapp", "grok-app", "data"), filepath.Join(app, "grokapp", "grok-app"), filepath.Join(app, "grok-app")}
	default:
		data := appdir.Getenv("XDG_DATA_HOME")
		if data == "" {
			data = filepath.Join(home, ".local", "share")
		}
		dirs = []string{filepath.Join(data, "grok-app"), filepath.Join(home, ".grok-app")}
	}
	for _, d := range dirs {
		if isDir(d) {
			return d
		}
	}
	return dirs[0]
}

func grokApp(home string) *Agent {
	dir := grokAppDir(home, grokAppOS)
	path := filepath.Join(dir, "agent-home", "config.toml")

	return &Agent{
		ID: "grok-app", Name: "Grok App", Icon: "grok", Aliases: []string{"grokapp"},
		Dir: dir, Path: path,
	}
}
