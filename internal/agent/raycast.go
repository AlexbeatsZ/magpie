package agent

// Raycast AI takes models of one's own as Custom Providers (Raycast Pro, on
// Mac and Windows; lx_78654 on Discord), from ~/.config/raycast/ai/
// providers.yaml in the user's folder on both (manual.raycast.com/ai/
// custom-providers):
//
//	providers:
//	  - id: magpie
//	    name: magpie
//	    base_url: http://127.0.0.1:3425/v1   # OpenAI compatible, without /chat/completions
//	    api_keys:
//	      magpie: magpie-raycast
//	    models:
//	      - id: <provider>/<model>
//	        name: …
//	        context: …
//	        abilities:
//	          temperature: {supported: true}
//	          vision: {supported: …}
//	          system_message: {supported: true}
//	          tools: {supported: true}
//	          reasoning_effort: {supported: …}
//
// Raycast watches the file and picks a change up without a restart. The
// model is picked in Raycast's own model picker (Quick AI, AI Chat, AI
// Commands), not in a file, so what magpie sets is whether its models are in
// that picker: one entry of the providers list, id "magpie". The user's other
// entries, the top-level tools block and comments stay, and so does a key
// magpie doesn't write in its own entry (include_user_email,
// additional_parameters, a model's description). Raycast's User-Agent isn't
// known, so its requests are told apart by their key.

import (
	"path/filepath"
)

// raycastApp is Raycast's macOS bundle.
var raycastApp = desktopApp{id: "com.raycast.macos", names: []string{"Raycast.app"}}

func raycast(home string) *Agent {
	// ~/.config in the user's folder on both systems, whatever
	// XDG_CONFIG_HOME says: the manual names that path
	dir := filepath.Join(home, ".config", "raycast")
	path := filepath.Join(dir, "ai", "providers.yaml")

	return &Agent{
		ID: "raycast", Name: "Raycast", Icon: "raycast", Aliases: []string{"raycast-ai"},
		Dir: dir, Path: path,
		// the app on a Mac; its config folder elsewhere, which Raycast for
		// Windows makes at Reveal Providers Config. No command: Raycast
		// has none
		detect: func() bool { return macBundle(raycastApp) || isDir(dir) },
	}
}
