package agent

// Pencil (pen.dev, the design canvas whose agent runs models) takes a
// custom provider of one's own in ~/.pencil/models.json, in Pi's format
// (徐小小 on Discord):
//
//	{"providers":{"<id>":{"name":…,"baseUrl":…,"api":"openai-responses",
//	  "models":[{"id":…,"name":…,"reasoning":true,"input":["text","image"],
//	    "cost":{"input":0,"output":0,"cacheRead":0,"cacheWrite":0},
//	    "contextWindow":…,"maxTokens":…,"compat":{}}]}}}
//
// Its settings offer a custom provider on OpenAI Completions, OpenAI
// Responses or Anthropic Messages (docs.pen.dev, Authentication) — Pi's
// openai-completions, openai-responses and anthropic-messages — set for the
// provider as a whole. So magpie's models are Pi's entries (piModelJSON)
// with every model on the provider's openai-completions at the gateway's
// /v1, which the gateway takes for every model: a model's own api and
// baseUrl, which Pi reads but Pencil's own editor never writes, are left
// out. Each model carries every field Pencil writes for one. The model is
// picked in Pencil's agent composer, not in a file magpie knows of, so what
// magpie sets is whether its models are in that picker; the user's other
// providers stay as they are. Its requests carry magpie-pencil, as its
// User-Agent isn't known.

import (
	"os"
	"path/filepath"
	"runtime"
)

func pencil(home string) *Agent {
	dir := filepath.Join(home, ".pencil")
	path := filepath.Join(dir, "models.json")

	return &Agent{
		ID: "pencil", Name: "Pencil", Icon: "generic", Aliases: []string{"pen.dev", "pen-dev", "pendev"},
		Dir: dir, Path: path,
		// ~/.pencil is the app's (and pen.dev's CLI's) once it has run; the
		// app installed and never opened is found by its bundle on a Mac.
		// No command: `pencil` and `pen` are common names.
		detect: func() bool { return isDir(dir) || pencilApp(home) },
	}
}

// pencilApp reports whether Pencil's app is installed where a Mac keeps
// apps, under its name before and after the pen.dev rename.
func pencilApp(home string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	for _, d := range macAppFolders(home) {
		for _, n := range []string{"Pencil.app", "pen.dev.app"} {
			if _, err := os.Stat(filepath.Join(d, n)); err == nil {
				return true
			}
		}
	}
	return false
}
