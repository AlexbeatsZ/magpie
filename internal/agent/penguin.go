package agent

// PenguinHarness (Prism-Shadow/penguin-harness, the penguin command; npm
// @prismshadow/penguin-cli) keeps each Project's models in one TOML file
// under its data root: $PENGUIN_HOME, else ~/.penguin/data on every system
// (core's resolveRoot, os.homedir()), the CLI's and the desktop app's
// alike. The CLI works in the Project $PENGUIN_PROJECT_ID, else
// default_project (client.ts), the one its server makes first, so the file
// is <root>/<project>/.project_config.toml (projectConfigPath). Its
// shape (core state/project-config.ts, penguin 0.2.17):
//
//	default_model = { provider = "magpie", model_id = "<model>" }
//	[providers.magpie]
//	base_url = "http://127.0.0.1:3425/v1"
//	client_type = "openai-responses"
//	api_key = "magpie-penguin"
//	[[models]]
//	provider = "magpie"
//	model_id = "<model>"
//	display_name = "…"
//	context_window = …
//	vision = true
//	max_tokens = …
//	[models.pricing]
//	unit = "usd_per_mtok"
//	cache_read = …
//	cache_write = …
//	output = …
//
// A model is the pair (provider, model_id); one without a base_url, client
// type or key of its own takes its group's, [providers.<provider>]. Its
// requests go through MMSP's client of that type: openai-responses posts
// base_url + /responses, with its thinking level as reasoning.effort,
// which openai-chat leaves out. The server reads the file again whenever
// its mtime moves (ProjectConfigService), so a change needs no restart;
// its own Models page and `penguin config model …` write the whole file
// back (smol-toml), every key magpie writes being one it knows. A
// default_model naming no entry is refused, so the entries go in before
// the default and out after it.

import (
	"path/filepath"
	"strings"
)

type penguinRef struct {
	Provider string `toml:"provider"`
	ModelID  string `toml:"model_id"`
}

// value is the pair as the model field spells it: provider/model_id.
func (r *penguinRef) value() string {
	if r == nil || r.ModelID == "" {
		return ""
	}
	return r.Provider + "/" + r.ModelID
}

// penguinRoot is the data root at a place; a WSL distro's variables
// aren't read, so it is ~/.penguin/data there.
func penguinRoot(at place) string {
	if d := at.getenv("PENGUIN_HOME"); d != "" {
		return d
	}
	return at.join(".penguin", "data")
}

// penguinProject is the Project the CLI works in: $PENGUIN_PROJECT_ID,
// where it names a folder of the root, else default_project.
func penguinProject(at place) string {
	if p := strings.TrimSpace(at.getenv("PENGUIN_PROJECT_ID")); p != "" && p != "." && p != ".." && !strings.ContainsAny(p, `/\`) {
		return p
	}
	return "default_project"
}

func penguin(home string) *Agent { return penguinIn(here(home)) }

// penguinIn is PenguinHarness at a place: this machine's home, or a WSL
// distro's (see wsl.go), its provider naming the gateway as it reaches it
// from there.
func penguinIn(at place) *Agent {
	root := penguinRoot(at)
	path := filepath.Join(root, penguinProject(at), ".project_config.toml")

	return &Agent{
		ID: "penguin", Name: "PenguinHarness", Icon: "penguin-color", Bin: "penguin",
		Aliases: []string{"penguin-harness", "penguinharness"},
		Dir:     filepath.Dir(root), Path: path,
		// the desktop app, on this machine (a distro has none): installed
		// and not opened yet, it has made no ~/.penguin
		sign: func() bool { return at.spell == nil && (&Agent{ID: "penguin"}).HasApp() },
	}
}
