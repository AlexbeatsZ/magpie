package agent

// Snow CLI (`snow`, npm snow-ai, MayDay-wpf/snow-cli) keeps its settings in
// ~/.snow ($SNOW_CONFIG_DIR when set). Each API setup is a profile,
// profiles/<name>.json, the whole of its config ({"snowcfg": {baseUrl,
// baseUrlMode, apiKey, requestMethod, advancedModel, basicModel,
// maxContextTokens, maxTokens, …}, …}); active-profile.json
// ({"activeProfile": "<name>"}) names the one in use, "default" when it is
// missing, and config.json is a copy of that one, which Snow writes again
// from the profile at every start (utils/config/configManager.ts
// initializeProfiles, switchProfile).
//
// magpie adds a profile of its own, profiles/magpie.json: a copy of the
// profile the user was on — their other settings kept, a subscription
// sign-in of theirs (snowcfg.oauth) left out — with the gateway's /v1 as
// its base URL, a key naming Snow CLI, Chat Completions as its protocol and
// the model picked as its main (advancedModel) and light (basicModel) model,
// the gateway resolving "<provider>/<model>" itself, and the model's
// window, output and image support from the catalog (snowSynced). It makes
// that profile the active one and copies it to config.json, as Snow's own
// profile switch does; the profile the user was on is stashed and made
// active again when magpie steps out, and magpie.json goes. Snow App
// (MayDay-wpf/snow-app) keeps API setups of its own: its "Sync Snow CLI API
// config" button copies these profiles, magpie's among them, into it.
//
// Snow's requests carry no User-Agent of its own (an x-snow: <version>
// header), so the key is what tells them apart: gateway.TokenFor("snow").

import (
	"path/filepath"
	"strings"
)

// snowProfile is the name of the profile magpie writes.
const snowProfile = magpieID

func snow(home string) *Agent { return snowIn(here(home)) }

// snowIn is Snow CLI at a place: this machine's home, or a WSL distro's (see
// wsl.go), whose profile names the gateway as the distro reaches it, with
// the key it takes from there.
func snowIn(at place) *Agent {
	dir := at.join(".snow")
	if d := strings.TrimSpace(at.getenv("SNOW_CONFIG_DIR")); filepath.IsAbs(d) {
		dir = filepath.Clean(d)
	}
	profiles := filepath.Join(dir, "profiles")
	ours := filepath.Join(profiles, snowProfile+".json")

	return &Agent{
		ID: "snow", Name: "Snow CLI", Icon: "snow", Aliases: []string{"snow-cli", "snow-ai", "snow-app"},
		Bin: "snow", Dir: dir, Path: ours,
	}
}
