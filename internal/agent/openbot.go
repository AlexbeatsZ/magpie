package agent

// OpenBot (nightly-labs/openbot, openbot.run; jiezhengj on Discord) is an
// Electron workspace that runs Codex, Claude, OpenCode and other agents,
// and lists the user's own OpenAI-compatible endpoints in its model picker.
// It keeps them in openbot-custom-providers-v1.json in Electron's userData
// folder, app data joined with its product name: ~/Library/Application
// Support/OpenBot on a Mac, %APPDATA%\OpenBot on Windows, ~/.config/OpenBot
// on Linux (src/main/index.ts, application-services.ts; v0.58.0). The file
// (custom-provider-store.ts, its zod fileSchema):
//
//	{"version":1,"providers":[{"id":"magpie","name":"magpie",
//	  "baseUrl":"http://127.0.0.1:3425/v1",
//	  "models":[{"id":"<model>","name":"<name>","contextWindow":<n>},…],
//	  "secret":null}]}
//
// version must be 1: a file OpenBot can't parse leaves it with no endpoints
// and refuses every save, so magpie leaves alone a file it can't read the
// same way. An entry's key and headers are one ciphertext, secret, sealed
// with Electron's safeStorage (the Keychain, DPAPI), which magpie doesn't
// open, so magpie's entry has none: OpenBot hands it to OpenCode keyless
// (opencode-config.ts, providerEntry), and the gateway takes a request with
// no key from this computer. OpenCode sends it, so its requests count as
// OpenCode's. A secret the user gave magpie's entry in OpenBot stays.
//
// OpenBot's list decoders fail closed: one entry or model it wouldn't accept
// empties the whole list. So the models are the ones picked for it, at most
// 64 (CUSTOM_PROVIDER_LIMITS.models), each with an id OpenBot takes as
// magpie/<id> (isAgentModel: ASCII letters, digits and ._:/[],=-, at most
// 160), and a context window only inside its 1,024 to 100,000,000.
//
// OpenBot reads the file once, when it starts, and writes it whole from
// what it holds only when an endpoint is saved, edited or removed there.
// So a change shows after a restart, and an endpoint saved in OpenBot
// before that writes magpie's entry out again: the notice says both. The
// model is picked in OpenBot's own picker, so what magpie sets is whether
// its models are in it. The user's other entries stay as written, and so
// do the name and harness they give magpie's.
//
// MCP servers live in OpenBot's SQLite database (McpServerStore), which
// magpie doesn't write; magpie's own (/mcp/magpie/…) are added there by
// hand as HTTP servers.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

// openBotFile is the file of OpenBot's custom endpoints in its userData.
const openBotFile = "openbot-custom-providers-v1.json"

// openBotApp is OpenBot's macOS bundle (electron-builder.yml: appId
// app.openbot.desktop, productName OpenBot).
var openBotApp = desktopApp{id: "app.openbot.desktop", names: []string{"OpenBot.app"}}

func openBot(home, cfg string) *Agent {
	dir := openBotDir(home, cfg)
	path := filepath.Join(dir, openBotFile)

	return &Agent{
		ID: "openbot", Name: "OpenBot", Icon: "generic", Aliases: []string{"open-bot"},
		Dir: dir, Path: path,
		// its userData folder once it has run, else its app
		detect: func() bool {
			return isDir(dir) || macBundle(openBotApp) || openBotWindowsApp()
		},
	}
}

// openBotDir is OpenBot's userData folder, as Electron finds it.
func openBotDir(home, cfg string) string {
	return openBotDirOn(runtime.GOOS, home, cfg, appdir.Getenv("APPDATA"))
}

// openBotDirOn is openBotDir on the system goos, with the XDG_CONFIG_HOME
// folder cfg and %APPDATA% appData.
func openBotDirOn(goos, home, cfg, appData string) string {
	switch goos {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "OpenBot")
	case "windows":
		d := appData
		if d == "" {
			d = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(d, "OpenBot")
	}
	return filepath.Join(cfg, "OpenBot")
}

// openBotWindowsApp says OpenBot's installer has put it where it installs
// it for one user, %LOCALAPPDATA%\Programs\<its name>, or for all of them.
func openBotWindowsApp() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	var dirs []string
	if d := appdir.Getenv("LOCALAPPDATA"); d != "" {
		dirs = append(dirs, filepath.Join(d, "Programs"))
	}
	for _, v := range []string{"ProgramW6432", "ProgramFiles"} {
		if d := appdir.Getenv(v); d != "" {
			dirs = append(dirs, d)
		}
	}
	for _, d := range dirs {
		for _, name := range []string{"OpenBot", "openbot"} {
			if isFile(filepath.Join(d, name, "OpenBot.exe")) {
				return true
			}
		}
	}
	return false
}
