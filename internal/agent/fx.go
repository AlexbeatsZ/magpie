package agent

// fx (fx.sh, Vercel Labs) keeps its settings in ~/.fx/settings.json: the
// provider it talks to under "provider" (gateway, codex, grok or one of
// the user's own), each provider's model under "models", and providers of
// the user's own under "providers". magpie adds itself there as the
// provider "magpie", keyless (auth none), speaking chat completions, with
// the catalog as its model_metadata; a model through magpie is
// "magpie/<provider>/<model>" here, provider "magpie" and models.magpie
// "<provider>/<model>" in fx's settings. fx sends no effort to a provider
// of the user's own, so there is no effort to pick.

import (
	"path/filepath"
)

func fx(home string) *Agent { return fxIn(here(home)) }

// fxIn is fx at a place: this machine's home, or a WSL distro's (see
// wsl.go), its provider naming the gateway as the distro reaches it.
func fxIn(at place) *Agent {
	dir := at.join(".fx")
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "fx", Name: "fx", Icon: "fx",
		Dir: dir, Path: path,
		// ~/.fx is where its settings are. No command: `fx` is a common name
		// (the JSON viewer has it too), as in WSL (wsl.go).
		detect: func() bool { return agentDir(dir) },
	}
}
