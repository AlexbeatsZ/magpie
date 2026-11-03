package agent

// Factory's Droid CLI keeps its settings in ~/.factory/settings.json
// ($FACTORY_HOME_OVERRIDE/.factory when set). Models of the user's own
// (BYOK) are the customModels array, each {model, displayName, baseUrl,
// apiKey, provider, …}, provider being "anthropic" (Messages), "openai"
// (Responses) or "generic-chat-completion-api" (Chat Completions); the
// older ~/.factory/config.json custom_models is still read, merged under
// settings.json, and magpie leaves it alone. A custom model is picked as
// "custom:<id>": its "id" when it has one, else its displayName with
// whitespace as dashes and a count of the same names before it
// ("custom:Kimi-K2-[Groq]-0"). The model sessions start on is
// sessionDefaultSettings.model; a top-level "model", as the docs show it,
// is moved there by droid when that is unset.
//
// magpie appends one entry per catalog model, id "custom:magpie/<ref>", so
// the catalog joins droid's /model picker under Custom models, each on the
// API its provider speaks natively so the gateway relays it as it is, and
// sets sessionDefaultSettings.model; the user's entries stay as they were
// and in front, so their own ids don't move, and the default they had is
// stashed and put back when magpie steps out.
//
// droid judges a custom model by its "model" too: one whose name reads as
// OpenAI's on another provider, or as a Claude off "anthropic", gets a
// warning at every session start (droidWants). Such a model is written on
// the provider droid asks for, and the gateway translates.
//
// A field of one of magpie's entries the user changed in settings.json
// stays theirs (droidKept): what magpie last wrote is stashed, and a value
// other than it is the user's. baseUrl and apiKey are the wiring, put back
// at every sync.

import (
	"path/filepath"
)

func droid(home string) *Agent { return droidIn(here(home)) }

// droidIn is Droid at a place: this machine's home, or a WSL distro's (see
// wsl.go), where FACTORY_HOME_OVERRIDE isn't read and its custom models
// name the gateway as the distro reaches it.
func droidIn(at place) *Agent {
	home := at.home
	if h := at.getenv("FACTORY_HOME_OVERRIDE"); h != "" {
		home = h
	}

	dir := at.open(filepath.Join(home, ".factory"))
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "droid", Name: "Droid", Icon: "factory", Aliases: []string{"factory", "factory-droid"},
		// droid's requests to a custom model carry factory-cli/<version>
		UA:  []string{"factory-cli"},
		Bin: "droid", Dir: dir, Path: path,
	}
}
