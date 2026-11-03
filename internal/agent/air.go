package agent

// JetBrains Air, JetBrains' agentic development environment, runs its own
// agents (Claude Agent, Codex, Gemini CLI, Junie) on the accounts signed in
// under Settings | Account | AI Providers: a sign-in or an API key, with no
// base URL of its own to point at magpie (AIR-4541 asks for one). What it
// does take is any agent that speaks the Agent Client Protocol, listed in
// acp.json in its config folder — ~/Library/Application Support/JetBrains/Air
// on macOS, $XDG_CONFIG_HOME/JetBrains/Air on Linux, %APPDATA%\JetBrains\Air
// on Windows — as
//
//	{"agent_servers": {"<name in Air's agent menu>": {"command": …, "args": […], "env": {…}}}}
//
// and the models such an agent reports at session/new are its model menu
// in Air. So magpie adds an agent of its own there, "Magpie": OpenCode's ACP
// server (opencode acp), handed with OPENCODE_CONFIG a config file magpie
// keeps beside acp.json that has magpie's provider, its models, the one
// picked as the model a task starts on, and enabled_providers narrowed to
// magpie, so Air's menu lists magpie's models and routing groups and nothing
// of OpenCode's own. OpenCode merges that file over the user's own config
// and leaves the rest of it (their plugins, agents, MCP servers) as it is.
// Air's own agents, and the accounts they run on, aren't touched.
//
// Checked with OpenCode 2.0.20's opencode acp under a sandbox HOME, as Air
// starts it: session/new's model option lists magpie's models alone, the
// one set here current. Not with Air itself. One gap is OpenCode's: when a
// provider's key is in the environment it is started with (OPENAI_API_KEY,
// DEEPSEEK_API_KEY, … set, not empty), 2.0.20's ACP server lists that
// provider and none of its config's — this file's, the global one or a
// project's alike — so magpie's models don't show; an emptied key, or one
// kept in OpenCode's auth.json, doesn't do it.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

// airDir is where Air keeps its settings and acp.json.
func airDir(home, cfg string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "JetBrains", "Air")
	case "windows":
		d := appdir.Getenv("APPDATA")
		if d == "" {
			d = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(d, "JetBrains", "Air")
	}
	return filepath.Join(cfg, "JetBrains", "Air")
}

func air(home, cfg string) *Agent {
	dir := airDir(home, cfg)
	path := filepath.Join(dir, "acp.json")

	return &Agent{
		ID: "air", Name: "JetBrains Air", Icon: "air", Aliases: []string{"jetbrains-air"},
		// no Bin: air is other tools' name too (Go's live reloader); its
		// requests are its OpenCode's, and counted as OpenCode's
		Dir: dir, Path: path,
	}
}
