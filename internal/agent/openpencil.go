package agent

// OpenPencil (ZSeven-W/openpencil, the openpencil-desktop app; @JC_be_better
// on X) is a design canvas whose chat runs models through "builtin agents"
// of the user's own: an Anthropic or OpenAI-compatible endpoint, a key and
// a list of models, each listed in its chat's model picker. It keeps them
// in one settings.json in the platform's config folder, dirs::config_dir()
// joined with openpencil: ~/Library/Application Support/openpencil on a
// Mac, $XDG_CONFIG_HOME/openpencil (~/.config/openpencil) on Linux,
// %APPDATA%\openpencil on Windows (op-editor-host-core's settings_io.rs,
// settings_path). Its fields (SettingsPayload, BuiltinAgentPayload,
// settings_payload.rs, main at 40902a3f and v0.8.4):
//
//	{"version":1, …,
//	 "builtin_agents":[{"id":"magpie","preset":"custom","display_name":"magpie",
//	   "kind":"openai-compat","api_key":"magpie-openpencil",
//	   "model":"<first model>","models":["<model>",…],
//	   "base_url":"http://127.0.0.1:3425/v1","enabled":true}], …}
//
// version is required (no serde default): a file without it fails to parse,
// is set aside as settings.json.corrupt-<time>, and OpenPencil then refuses
// to save at all; a version other than 1 has the whole file ignored. So
// magpie writes version 1 in a file it makes, and leaves alone a file whose
// version it doesn't know. kind openai-compat posts to base_url +
// /chat/completions, so base_url is the gateway's /v1. models is the list
// (at most 64, written only when there are two or more, the first again as
// model); v0.8.4 knows only model and reads past models, so it lists the
// first model alone.
//
// OpenPencil reads the file once, when it starts, and writes the whole of
// it back from what it holds when a setting changes and when it quits
// (app_handler.rs, settings_io::save). A change magpie makes while it runs
// is written over when it quits, so magpie refuses one then, where it can
// tell (Running can't on Windows, where the notice says so instead).
//
// The model is picked in OpenPencil's own chat picker
// (builtin:magpie:<model>), not in a file, so what magpie sets is whether
// its models are in that picker: one entry of builtin_agents, id magpie
// (OpenPencil numbers its own builtin-N, and reads a picker value up to its
// second colon). The user's other entries and settings stay, and so do the
// name and the on/off they give magpie's entry in OpenPencil's settings.
// Its User-Agent isn't known, so its requests are told apart by their key.
//
// open-pencil/open-pencil, the other app of the name, keeps its providers in
// its webview's localStorage, no file magpie can write: it reaches magpie
// through an agent magpie wires (Pi, an ACP agent) or a base URL typed in.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

// openPencilApp is OpenPencil's macOS bundle, as its release builds it
// (scripts/bundle-macos.sh).
var openPencilApp = desktopApp{id: "com.zseven-w.openpencil", names: []string{"OpenPencil.app"}}

func openPencil(home, cfg string) *Agent {
	dir := openPencilDir(home, cfg)
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "openpencil", Name: "OpenPencil", Icon: "generic", Aliases: []string{"open-pencil", "openpencil-desktop"},
		Dir: dir, Path: path,
		// its settings folder once it has run, ~/.openpencil once its op
		// CLI or MCP server has, else its app. No command: op is its CLI's
		// name, and other tools' too
		detect: func() bool {
			return isDir(dir) || isDir(filepath.Join(home, ".openpencil")) || macBundle(openPencilApp) || openPencilWindowsApp()
		},
	}
}

// openPencilDir is the folder of OpenPencil's settings.json, as Rust's
// dirs::config_dir() finds the platform's config folder.
func openPencilDir(home, cfg string) string {
	return openPencilDirOn(runtime.GOOS, home, cfg, appdir.Getenv("APPDATA"))
}

// openPencilDirOn is openPencilDir on the system goos, with the
// XDG_CONFIG_HOME folder cfg and %APPDATA% appData.
func openPencilDirOn(goos, home, cfg, appData string) string {
	switch goos {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "openpencil")
	case "windows":
		d := appData
		if d == "" {
			d = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(d, "openpencil")
	}
	return filepath.Join(cfg, "openpencil")
}

// openPencilWindowsApp says OpenPencil's installer has put it where it
// installs it, %ProgramFiles%\OpenPencil (scripts/package-windows.nsi).
func openPencilWindowsApp() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	for _, v := range []string{"ProgramW6432", "ProgramFiles"} {
		if d := appdir.Getenv(v); d != "" && isFile(filepath.Join(d, "OpenPencil", "openpencil-desktop.exe")) {
			return true
		}
	}
	return false
}
