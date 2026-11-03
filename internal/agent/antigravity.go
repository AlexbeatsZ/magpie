package agent

// Antigravity, Google's desktop app (2.15 and later), sends its requests to
// a gateway of the user's instead of Google when ~/.antigravity/settings.json
// turns one on: Google's Antigravity Enterprise Gateway, documented at
// docs.cloud.google.com/gemini/enterprise/docs/ai-developer-tools/connect-enterprise-gateway
// (ryougi_hachi on X):
//
//	{"gateway":{"enabled":true,"url":…,"apiKey":…,"wireProtocol":"openai"},
//	 "models":{"defaultModelId":…,"list":[{"modelId":…}]}}
//
// On the openai wire protocol the app asks <url>/v1/models for its model
// picker (models.list, when there is one, narrows it) and sends a chat to
// <url>/v1/chat/completions, streamed, with apiKey as a bearer token and a
// User-Agent of antigravity/hub/<version> (auth_method=gateway). magpie's
// gateway answers both, so what magpie writes is that block at its own
// address, with a key that names the app, and the model the app starts on.
// The user's own gateway, default model and model list are stashed and put
// back when magpie steps out. On a gateway the app asks for no Google
// sign-in and uses none: Google's docs say the two can't be used at once.
//
// The same file is read by agy (Antigravity's CLI) too, below its own
// ~/.gemini/antigravity-cli/settings.json: an agy connected on its own row
// (modelProvider gemini) keeps that route, one that isn't follows the app's
// gateway. admin_settings.json, which an organisation installs, and the
// AGY_LLM_GATEWAY_* variables win over this file, field by field.

import (
	"path/filepath"
)

func antigravity(home string) *Agent {
	dir := filepath.Join(home, ".antigravity")
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "antigravity", Name: "Antigravity", Icon: "antigravity-color", Aliases: []string{"antigravity-app", "agy-app"},

		// ~/.antigravity is also where Antigravity IDE (a VS Code fork) and
		// agy keep theirs, so it says nothing of the app: the app is its
		// bundle (2.15 or later, the first to take a gateway) or the file
		// only its own backend writes
		Dir: dir, Path: path, dirShared: true,
		sign: func() bool {
			if v := (&Agent{ID: "antigravity"}).AppVersion(); v != "" && !Newer("2.15.0", v) {
				return true
			}
			return isFile(filepath.Join(home, ".gemini", "antigravity", "agyhub_summaries_proto.pb"))
		},
	}
}
