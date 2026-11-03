package agent

// jcode (1jehuang/jcode, a coding agent in Rust) keeps its settings in
// config.toml in $JCODE_HOME, else ~/.jcode. A provider of the user's own
// making is a named profile, [providers.<name>], which `jcode provider add`
// writes: type "openai-compatible" (jcode's chat completions client), a
// base_url, a key (api_key inline, or api_key_env) and its models as
// [[providers.<name>.models]] entries (id, context_window, input,
// reasoning). jcode starts on [provider] default_provider and default_model,
// a profile's name and one of its models.
//
// magpie adds itself as the profile "magpie": the gateway, with one model
// entry per catalog model, so the catalog is in jcode's /model picker, and
// jcode is started on the model picked. What [provider] named before is
// stashed and put back when magpie steps out. Every other key of the file is
// the user's and stays as written, the user's own keys in magpie's profile
// too (headers, extra_body, a model's reasoning_effort).

import (
	"path/filepath"
	"strings"
)

// jcodeDir is jcode's folder: $JCODE_HOME, else ~/.jcode (jcode-storage's
// jcode_dir).
func jcodeDir(at place) string {
	if d := strings.TrimSpace(at.getenv("JCODE_HOME")); d != "" {
		return d
	}
	return at.join(".jcode")
}

func jcode(home string) *Agent { return jcodeAt(here(home)) }

func jcodeAt(at place) *Agent {
	dir := jcodeDir(at)
	path := filepath.Join(dir, "config.toml")

	return &Agent{
		ID: "jcode", Name: "jcode", Icon: "jcode",
		Bin: "jcode", Dir: dir, Path: path,
	}
}
