package agent

// MothX (oschina/mothx, the mothx command; formerly vibecoding) is a Go
// terminal coding agent that reads its providers and the model it starts
// on from one settings.json in its config folder: $MOTHX_DIR, else
// %APPDATA%\mothx on Windows and ~/.mothx elsewhere (platform.ConfigDir,
// configDirForOS). Its fields (config.Settings, ProviderConfig,
// ModelConfig, mothx 1.3.104):
//
//	{"defaultProvider":"magpie","defaultModel":"<model>",
//	 "defaultThinkingLevel":"high",
//	 "providers":{"magpie":{"baseUrl":"http://127.0.0.1:3425/v1",
//	   "api":"openai-chat","apiKey":"magpie-mothx","thinkingFormat":"openai",
//	   "models":[{"id":…,"name":…,"reasoning":true,"contextWindow":…,
//	     "maxTokens":…,"cost":{…},"input":["text","image"]}]}}}
//
// A session asks provider defaultProvider for model defaultModel on the
// provider's api (openai-chat: baseUrl + /chat/completions) under Bearer
// apiKey. Its requests carry "Vibecoding Client/<version>" as their
// User-Agent, its old name an alias here; the key magpie gives it names it
// too (gateway.TokenFor). It
// reads settings.json at start-up, and its own /model and /auth dialogs
// write the whole file back (SaveGlobalSettings), every default spelled out.
//
// Thanks to zhenruyan, whose fork's internal/agent/mothx.go (#2353) this
// follows.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

func mothx(home string) *Agent { return mothxIn(here(home)) }

// mothxIn is MothX at a place: this machine's home, or a WSL distro's (see
// wsl.go), its settings.json naming the gateway as it reaches it from there.
func mothxIn(at place) *Agent {
	dir := mothxDir(at)
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "mothx", Name: "MothX", Icon: "mothx", Bin: "mothx", Dir: dir, Path: path,
		// the command and the package were vibecoding before (its npm
		// vibecoding-installer still installs it under that name), and
		// its requests still say "Vibecoding Client/<version>": usage
		// names them by the word before the space (usage.AgentOf)
		Aliases: []string{"vibecoding"},
	}
}

// mothxDir is MothX's config folder at a place: $MOTHX_DIR (a relative one
// is its working directory's, which magpie can't know), else
// %APPDATA%\mothx on Windows, else ~/.mothx. A distro's variables aren't
// read, so it is ~/.mothx there.
func mothxDir(at place) string {
	if d := at.getenv("MOTHX_DIR"); d != "" {
		return d
	}
	if runtime.GOOS == "windows" && at.spell == nil {
		if app := appdir.Getenv("APPDATA"); app != "" {
			return filepath.Join(app, "mothx")
		}
		return at.join("AppData", "Roaming", "mothx")
	}
	return at.join(".mothx")
}
