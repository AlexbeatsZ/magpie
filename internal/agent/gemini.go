package agent

import (
	"path/filepath"

	"github.com/yetone/magpie/internal/proc"
)

// Gemini CLI only speaks Google's own API, so "provider" here is how it
// authenticates: Google sign-in, a Gemini API key, or Vertex AI. The choice
// lives in settings.json (security.auth.selectedType). A Google provider
// added in magpie lends its key through ~/.gemini/.env, which the CLI loads.
//
// Any catalog model works too: the gateway serves the Gemini API, so magpie
// points GOOGLE_GEMINI_BASE_URL at it, uses the gateway token as the API
// key and names the catalog model in model.name.

func gemini(home string) *Agent { return geminiIn(here(home)) }

// geminiIn is Gemini CLI at a place: this machine's home, or a WSL
// distro's (see wsl.go), whose .env names the gateway as the distro
// reaches it, with the key it takes from there.
func geminiIn(at place) *Agent {
	dir := at.join(".gemini")
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "gemini", Name: "Gemini CLI", Icon: "geminicli-color", Aliases: []string{"gemini-cli"},
		UA:  []string{"geminicli", "gemini-cli"},
		Bin: "gemini", Dir: dir, Path: path,

		// Gemini CLI is its binary. What it leaves in ~/.gemini stays when it
		// is uninstalled (#230), and Antigravity keeps its folders there too
		// (antigravity, antigravity-cli, config) and reads GEMINI.md (#330),
		// so nothing there says Gemini CLI is here: a row for it would set
		// up a CLI that can't run. The desktop app's PATH has the user's
		// shell's (proc.UserPath).
		detect: func() bool {
			if Taken(dir) {
				return false
			}
			_, err := proc.Look("gemini")
			return err == nil
		},
	}
}
