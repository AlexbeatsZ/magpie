package agent

// Mister Morph (morph) keeps its settings in ~/.morph/config.yaml, or the
// file $MISTER_MORPH_CONFIG names: the main model under llm.model, asked of
// the vendor llm.inference_provider names at llm.endpoint with llm.api_key.
// It has no list of providers to add one to, so magpie takes llm over:
// inference_provider openai_response_compatible at the gateway's /v1 (morph
// adds no second /v1 to a base ending in one), spoken to on OpenAI's Responses
// API, morph's own default, which the gateway serves for every vendor; its
// token as the key, and the catalog id as the model; a model through magpie
// is "magpie/<provider>/<model>" here. llm.provider is left as it is: morph
// derives it from inference_provider. Morph knows a model's window only from a list of its own, by the id
// after its last slash, so magpie hands it the window it knows in
// llm.context_window_tokens. What the user had under those keys is stashed
// and put back when magpie steps out. llm.profiles and llm.routes are the
// user's: a route that sends the main loop to another profile is left so.

import (
	"path/filepath"
)

// morphUA is the User-Agent magpie has morph send, in llm.headers: its
// requests otherwise carry the OpenAI SDK's, which tells them from no one's.
const morphUA = "mistermorph"

func morph(home string) *Agent { return morphIn(here(home)) }

// morphIn is Mister Morph at a place: this machine's home, or a WSL distro's
// (see wsl.go), where MISTER_MORPH_CONFIG isn't read.
func morphIn(at place) *Agent {
	dir := at.join(".morph")
	path := at.getenv("MISTER_MORPH_CONFIG")
	if path == "" {
		path = filepath.Join(dir, "config.yaml")
	}

	return &Agent{
		ID: "morph", Name: "Mister Morph", Icon: "mistermorph-color", Aliases: []string{"mistermorph", "mister-morph"},
		UA: []string{morphUA},
		// no Bin: morph is other tools' name too; ~/.morph is made by its
		// first run, `morph install` or the desktop app's setup
		Dir: dir, Path: path,
	}
}
