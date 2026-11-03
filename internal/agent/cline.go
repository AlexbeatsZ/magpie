package agent

// Cline's CLI (3.x) keeps its providers in $CLINE_DIR/data/settings/
// providers.json, ~/.cline by default:
//
//	{"version":1,"lastUsedProvider":"<id>","modes":{},"providers":{"<id>":{
//	  "settings":{"provider":"<id>","apiKey":…,"model":…,"baseUrl":…,
//	    "headers":{…},"reasoning":{"effort":…}},
//	  "updatedAt":…,"tokenSource":"manual"}}}
//
// A session runs the model of lastUsedProvider, asking for its
// reasoning.effort. A provider of one's own, added to models.json, is
// refused when a session starts ("Unknown or disabled provider",
// cline/cline#14180), so magpie takes Cline's built-in openai-compatible
// provider, which reads the models from the gateway's /v1/models, and makes
// it the one in use. That provider lists only gpt-4o, so magpie's models
// are put in models.json beside it, as the entry Cline's own migration writes
// for it ({"provider":{"name","baseUrl","defaultModelId"},"models":{…}}),
// which takes the place of the built-in list. What was there, and the
// provider in use, are stashed and put back when magpie steps out. Cline's
// requests name only the AI SDK, so the header says they are Cline's.
//
// reasoning is {"enabled":false} for no thinking, as Cline's own pickers
// write it; an effort beside a false enabled is dropped, so magpie writes
// {"enabled":true,"effort":…} for an effort.
//
// Cline's VS Code extension (4.x) runs on the same files, but takes the
// provider, model, base URL and key from its own state beside them first —
// $CLINE_DATA_DIR (else $CLINE_DIR/data) globalState.json's
// act/planModeApiProvider ("openai" for openai-compatible),
// act/planModeOpenAiModelId and openAiBaseUrl, secrets.json's openAiApiKey
// — and providers.json's lastUsedProvider only when no mode has one. Where
// that state is, magpie sets those too, stashing what they were. The
// extension reads the files once, at start-up.
//
// Cline's desktop app (0.0.x) keeps the provider, model and effort picked
// in its composer in its web view's localStorage, per provider, and never
// reads them from these files: magpie's models are in its picker, under
// magpie, but which one it runs is picked there.

import (
	"path/filepath"
)

func cline(home string) *Agent { return clineIn(here(home)) }

// clineIn is Cline at a place: this machine's home, or a WSL distro's (see
// wsl.go), where CLINE_DIR and CLINE_DATA_DIR aren't read and its provider
// names the gateway as the distro reaches it, with the key it takes from
// there.
func clineIn(at place) *Agent {
	dir := at.getenv("CLINE_DIR")
	if dir == "" {
		dir = at.join(".cline")
	}
	data := at.getenv("CLINE_DATA_DIR")
	if data == "" {
		data = filepath.Join(dir, "data")
	}
	path := filepath.Join(data, "settings", "providers.json")

	return &Agent{
		ID: "cline", Name: "Cline", Icon: "cline", Aliases: []string{"cline-cli"},
		UA:  []string{"cline"},
		Bin: "cline", Dir: dir, Path: path,
	}
}
