package agent

// Antigravity's CLI, agy, keeps its settings in
// ~/.gemini/antigravity-cli/settings.json. With "modelProvider": "gemini" it
// asks a Gemini API endpoint rather than Google's own backend, and lists the
// models of customModelsConfig.customModels, each {apiProvider, modelName},
// the one picked being "model". magpie's gateway speaks the Gemini API, so
// magpie adds one custom model per catalog model, keyed
// "magpie/<provider>/<model>", and sets modelProvider and model; the user's
// own are stashed and put back when magpie steps out. The endpoint and key
// agy takes only from its environment (GOOGLE_GEMINI_BASE_URL and
// GEMINI_API_KEY; no .env file), so magpie can't set them: the agent's row
// gives the command that starts agy with them (see Agent.Launch). That
// names the model with --model too: agy 1.2 asks for a custom model picked
// in settings.json or its /model picker with no model name ("model is
// empty"); only one given on the command line resolves.
//
// So that typing agy just works (bestony on Discord: the command was
// copied at each start), magpie puts a function named agy in the profile
// of each shell the user has, between "# >>> magpie agy >>>" and
// "# <<< magpie agy <<<" (update.ShellFunction), while agy is on one of
// magpie's models. It runs `magpie agy run`, which starts the real agy
// with the two variables and --model when settings.json's model is still
// magpie's and as it is otherwise (cli/agy_cli.go); PowerShell's asks
// `magpie agy env` for them, since the app it would run is a GUI program
// PowerShell doesn't wait for. A terminal open before takes the command
// from the row.

import (
	"path/filepath"
)

func agy(home string) *Agent { return agyIn(here(home)) }

// agyIn is agy at a place: this machine's home, or a WSL distro's, whose
// shells get a function of their own (agyDistroFunction).
func agyIn(at place) *Agent {
	home := at.home
	dir := at.open(filepath.Join(home, ".gemini", "antigravity-cli"))
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "agy", Name: "Antigravity CLI", Icon: "antigravity-color", Aliases: []string{"antigravity-cli"},
		Bin: "agy", Dir: dir, Path: path,
	}
}
