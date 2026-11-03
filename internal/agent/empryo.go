package agent

// Empryo keeps its settings in ~/.empryo/config.json (a project's
// .empryo/config.json goes over it). Its own providers are a "providers"
// array of {id, name, baseURL, envVar, models, modelsAPI}, OpenAI-compatible
// endpoints it lists models from (modelsAPI, else baseURL's /models) and
// asks /chat/completions of; with no envVar it sends no key of the user's.
// The model sessions start on is defaultModel, "<provider id>/<model id>",
// split at the first "/".
//
// magpie adds a provider of its own, id "magpie", at the gateway, and sets
// defaultModel to magpie/<catalog id>; the defaultModel the user had is
// stashed and put back when magpie steps out, and the "magpie" provider
// goes again then. The user's providers and other keys are left as they
// are. With no envVar Empryo only lists a provider whose baseURL answers
// (2xx, 401 or 403), which the gateway's GET /v1 does.

import (
	"path/filepath"
)

func empryo(home string) *Agent { return empryoAt(here(home)) }

// empryoIn is Empryo in a WSL distro (see wsl.go).
func empryoIn(at place) *Agent { return empryoAt(at) }

// empryoAt is Empryo with its home at at.home, reaching the gateway as at
// does.
func empryoAt(at place) *Agent {

	dir := at.join(".empryo")
	path := filepath.Join(dir, "config.json")

	return &Agent{
		ID: "empryo", Name: "Empryo", Icon: "empryo-color", Aliases: []string{"soulforge"},
		Bin: "empryo", Dir: dir, Path: path,
	}
}
