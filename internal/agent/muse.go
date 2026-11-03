package agent

// Muse Code, Meta's muse CLI, keeps its settings in
// ${XDG_CONFIG_HOME:-~/.config}/muse/settings.json, a file it refuses
// without "schema_version": 1. Where its requests go is endpoint_transport:
// base_url, the /responses it asks (OpenAI's Responses API, streamed), and
// auth, "bearer" for the Meta sign-in's token (auth.json, or META_API_KEY)
// or "none" for no Authorization at all and no sign-in asked for. The model
// sessions start on is the top-level model, one of those Muse lists from
// <base_url's host>/muse-code/models, which the gateway serves it.
//
// magpie sets endpoint_transport to the gateway with auth "none" — so the
// Meta sign-in's token, which magpie never reads, isn't sent to it either —
// and model to the catalog id; the endpoint_transport and model the user had
// are stashed and put back when magpie steps out, and the file's other keys
// are left as they are.

import (
	"path/filepath"
)

func muse(cfg string) *Agent { return museAt(here(""), cfg) }

// museIn is Muse Code in a WSL distro (see wsl.go): ~/.config/muse, as
// XDG_CONFIG_HOME there isn't read.
func museIn(at place) *Agent { return museAt(at, at.join(".config")) }

// museAt is Muse Code with its config under cfg, reaching the gateway as
// at does.
func museAt(at place, cfg string) *Agent {

	dir := at.open(filepath.Join(cfg, "muse"))
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "muse", Name: "Muse Code", Icon: "meta-color", Aliases: []string{"muse-code", "musecode"},
		// muse-build/1.4.2 (non-interactive; macos-aarch64; …)
		UA:  []string{"muse-build", "muse-code"},
		Bin: "muse", Dir: dir, Path: path,
	}
}
