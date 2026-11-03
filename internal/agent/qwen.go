package agent

// Qwen Code (`qwen`, QwenLM/qwen-code, a Gemini CLI fork) keeps its
// settings in ~/.qwen/settings.json ($QWEN_HOME when set). Models of the
// user's own are the modelProviders.<protocol> arrays, each
// {baseUrl, envKey, id, name}: id is the model asked for, name what its
// /model picker shows, and envKey names a variable in the same file's env
// object (or ~/.qwen/.env) with the key. The model sessions start on is
// settings.model {name, baseUrl}, baseUrl telling two entries of one id
// apart, and security.auth.selectedType "openai" skips the /auth prompt
// on start.
//
// magpie appends one entry per catalog model to modelProviders.openai,
// the wire id the gateway takes ("<provider>/<model>") and the name
// "magpie/<provider>/<model>", every one on the same env key of its own,
// which is how an entry of magpie's is told from the user's own, and
// points settings.model and security.auth.selectedType at a pick. The key
// itself lives in the settings' env object — a shell-exported or
// ~/.qwen/.env variable of the same name would win over it, which the
// qwen-specific name makes unlikely. The user's entries stay as they were
// and in front, so their own ids don't move, and the default, auth type
// and any key of theirs they had are stashed and put back when magpie
// steps out.

import (
	"path/filepath"
)

func qwen(home string) *Agent { return qwenIn(here(home)) }

// qwenIn is Qwen Code at a place: this machine's home, or a WSL distro's
// (see wsl.go), whose entries name the gateway as the distro reaches it,
// with the key it takes from there.
func qwenIn(at place) *Agent {
	dir := at.join(".qwen")
	if d := at.getenv("QWEN_HOME"); d != "" {
		dir = d
	}
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "qwen", Name: "Qwen Code", Icon: "qwen-color", Aliases: []string{"qwen-code", "qwen-cli"},
		// qwen's requests carry QwenCode/<version> (packages/core/src/core/
		// openaiContentGenerator/provider/default.ts buildHeaders)
		UA:  []string{"qwencode"},
		Bin: "qwen", Dir: dir, Path: path,
	}
}
