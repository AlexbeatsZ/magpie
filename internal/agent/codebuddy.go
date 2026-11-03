package agent

// CodeBuddy Code (Tencent's CLI, npm @tencent-ai/codebuddy-code: codebuddy,
// cbc) is one package for the international edition (codebuddy.ai) and the
// China one (codebuddy.cn, CODEBUDDY_INTERNET_ENVIRONMENT=internal). Both
// take models of one's own from ~/.codebuddy/models.json ($CODEBUDDY_CONFIG_DIR
// if set), in WorkBuddy's format (workbuddy.go):
//
//	{"models":[{"id":"deepseek/pro","name":…,"vendor":"magpie",
//	  "apiKey":"magpie-codebuddy","url":"http://127.0.0.1:3425/v1/chat/completions",
//	  "maxInputTokens":…,"maxOutputTokens":…,"supportsToolCall":true,…}],
//	 "availableModels":[…]}
//
// and pick up a change to it without a restart. Such a model needs no
// CodeBuddy sign-in: the CLI asks url for it under Bearer apiKey. The model
// a session starts on is settings.json's "model" (an id from that list, or
// one of CodeBuddy's own), read when a session starts. Its requests say
// "CLI/<ver> CodeBuddy/<ver>", so they are told apart by their key.
//
// The CodeBuddy IDE (CodeBuddy.app and CodeBuddy CN.app, 4.12.2; see
// desktopApps) reads the same file: its genie extension's
// CustomModelsProductProvider loads ~/.codebuddy/models.json's "models"
// (an entry with no string id or url is skipped) and "availableModels",
// watches the file and syncs its model list on a change, and lists each one
// in its own model picker. It needs no CLI, so one row, CodeBuddy, stands
// for both (bjzhangwn on X: only the CLI was supported, not the IDE). The
// IDE reads <home>/.codebuddy whatever CODEBUDDY_CONFIG_DIR says, and not
// settings.json's "model": its model is picked in the IDE.

import (
	"path/filepath"
)

func codebuddy(home string) *Agent { return codebuddyIn(here(home)) }

// codebuddyIn is CodeBuddy Code at a place: this machine's home, or a WSL
// distro's (see wsl.go), where CODEBUDDY_CONFIG_DIR isn't read and its
// models name the gateway as the distro reaches it, with the key it takes
// from there.
func codebuddyIn(at place) *Agent {
	dir := at.getenv("CODEBUDDY_CONFIG_DIR")
	if dir == "" {
		dir = at.join(".codebuddy")
	}
	path := filepath.Join(dir, "models.json")

	// the IDE alone is CodeBuddy here, its CLI not installed
	ide := func() bool { return at.spell == nil && (&Agent{ID: "codebuddy"}).HasApp() }
	return &Agent{
		ID: "codebuddy", Name: "CodeBuddy", Icon: "codebuddy-color", Aliases: []string{"codebuddy-code", "cbc"},
		UA:  []string{"codebuddy"},
		Bin: "codebuddy", Dir: dir, Path: path, sign: ide,
	}
}
