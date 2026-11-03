package agent

// Grok Build, xAI's grok CLI, keeps its settings in ~/.grok/config.toml (or
// $GROK_HOME's): the model new sessions start with under [models] default,
// and models of the user's own as [model."<id>"] tables. magpie adds one
// such table per catalog model, named "magpie/<provider>/<model>", pointing
// at the gateway with magpie's own key — a model with no key of its own
// would be sent the user's xAI sign-in — so the catalog joins Grok's own
// models in its /model picker.
//
// A signed-in Grok also takes remote "campaign" patches from xAI, applied
// above config.toml, and a launch campaign sets models.default (September
// 2026's grok-4.7-launch did): a default picked in magpie was ignored and
// every new session started on the campaign's model. So a default picked
// here turns campaigns off ([features] campaigns = false), and clearing it
// gives them back.

import (
	"path/filepath"
)

func grok(home string) *Agent { return grokIn(here(home)) }

// grokIn is Grok Build at a place: this machine's home, or a WSL distro's
// (see wsl.go), where GROK_HOME isn't read and its models name the gateway
// as the distro reaches it.
func grokIn(at place) *Agent {
	dir := at.getenv("GROK_HOME")
	if dir == "" {
		dir = at.join(".grok")
	}
	path := filepath.Join(dir, "config.toml")

	return &Agent{
		ID: "grok", Name: "Grok Build", Icon: "xai", Aliases: []string{"grok-build", "grok-cli"},
		UA:  []string{"grok-shell", "grok-pager", "xai-grok-build"}, // grok-pager: its terminal front end
		Dir: dir, Path: path,
	}
}
