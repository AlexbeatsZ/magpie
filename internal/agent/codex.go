package agent

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/yetone/magpie/internal/codexcat"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/filememo"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

func codex(home string) *Agent { return codexIn(here(home)) }

// The two notices magpie's multi-agent V1 setting (#141) can carry: Codex's
// own features.multi_agent_v2 is on in the config magpie routes and takes the
// V1 over, or magpie couldn't read that config to tell. magpie writes nothing
// here — the setting is the user's own to turn off.
var (
	codexV1ConflictNotice  = newNotice("Codex's own multi_agent_v2 is on in the config magpie routes, which takes over the multi-agent V1 magpie's models ask for; turn it off in Codex's config to run subagents on magpie's models.")
	codexV1ReadErrorNotice = newNotice("Codex's multi-agent V1 is on, but magpie couldn't read Codex's config to check its own multi_agent_v2.")
)

// multiAgentV2OnIn reads one [features] table for Codex's own multi_agent_v2,
// read-only: the bare multi_agent_v2 boolean, or enabled in the
// [features.multi_agent_v2] table, spelled as a table, an inline object or a
// dotted key. set reports whether magpie recognizes the key as on or off. A
// table whose enabled is left out, or any other value (a string "true"), is
// not recognized, so magpie reads it as nothing and warns on nothing — Codex
// itself rejects such a config at load (multi_agent_v2 is a bool or a table
// there), so this is magpie's read of an invalid file, not Codex's default.
func multiAgentV2OnIn(features map[string]any) (on, set bool) {
	switch v := features["multi_agent_v2"].(type) {
	case bool:
		return v, true
	case map[string]any:
		enabled, ok := v["enabled"].(bool)
		return enabled, ok
	}
	return false, false
}

// codexMultiAgentV2On reports whether a Codex config explicitly turns its own
// features.multi_agent_v2 on: the bare multi_agent_v2 = true key, or enabled =
// true in the [features.multi_agent_v2] table (a table, an inline object or a
// dotted key), read at the top level and, on the legacy in-file profile shape,
// in the profile the config selects. Codex 0.133.0 applied that overlay itself
// (core/config/mod.rs: Features::from_sources with base = [features] and
// profile = [profiles.<name>.features], the profile applied on top), and
// magpie's own Check already reads profiles.<profile> the same way; that legacy
// shape is what this mirrors. From 0.134 Codex dropped in-file profiles: an
// in-file `profile` key is an error and `--profile` selects a separate
// {profile}.config.toml layer, so this is not a claim about which profile a
// modern Codex is running under — only about the same in-file shape magpie's
// Check reads. The whole file is decoded (codexTopOf, as codexEnableEfforts
// reads it). A config magpie can't read is an error, never taken for off; no
// config file has no key, nothing Codex turns V2 on with.
func codexMultiAgentV2On(path string) (bool, error) {
	doc, err := codexTopOf(path)
	if err != nil {
		return false, err
	}
	top, _ := doc.Features.(map[string]any)
	on, _ := multiAgentV2OnIn(top)
	// the legacy in-file profile's features override the top level's, as Codex
	// 0.133.0 applied them (base then profile) and as magpie's Check reads
	// profiles.<profile>: take the profile's value when it sets one, else keep
	// the top level's. Only the profile the config names counts.
	if profile, _ := doc.Profile.(string); profile != "" {
		profiles, _ := doc.Profiles.(map[string]any)
		p, _ := profiles[profile].(map[string]any)
		pf, _ := p["features"].(map[string]any)
		if pon, set := multiAgentV2OnIn(pf); set {
			on = pon
		}
	}
	return on, nil
}

// codexTop is the tables of Codex's config magpie's looks decode the file
// for. A config trusting every project Codex was opened in holds thousands
// of [projects."…"] tables (the owner's 10,932, in 1.5 MB), and decoding
// those again for each look was 3 MB a look, over 20 MB a catalog sync.
type codexTop struct {
	Features       any `toml:"features"`
	ModelProviders any `toml:"model_providers"`
	Desktop        any `toml:"desktop"`
	Profile        any `toml:"profile"`
	Profiles       any `toml:"profiles"`
}

// codexTopOf is the config at path as codexTop, decoded again only once the
// file changes; it is looked at each time (filememo.ReadNow), as what is
// written is made from it. What it returns is shared: nothing in it is
// changed. A missing file is an empty config.
func codexTopOf(path string) (*codexTop, error) {
	doc, err := filememo.ReadNow("codex top", path, func(b []byte) (*codexTop, error) {
		var doc codexTop
		if err := toml.Unmarshal(edit.TrimBOM(b), &doc); err != nil {
			return nil, err
		}
		return &doc, nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return &codexTop{}, nil
	}
	return doc, err
}

// codexSubAgentsV1Warning is the notice for the config Codex is routed
// through (path, routed): settings.CodexAgentsV1 is on and Codex's own
// multi_agent_v2 is explicitly on there, so Codex's V2 wins over the V1
// magpie's own models ask for and a subagent on one can't read its task.
// "" when the setting is off, Codex isn't routed through magpie here, or its
// own multi_agent_v2 is off; a config magpie can't read is told, not taken
// for off.
func codexSubAgentsV1Warning(path string, routed bool) string {
	if !codexcat.V1() || !routed {
		return ""
	}
	on, err := codexMultiAgentV2On(path)
	if err != nil {
		return codexV1ReadErrorNotice.String()
	}
	if !on {
		return ""
	}
	return codexV1ConflictNotice.String()
}

// CodexSubAgentsV1Warning is codexSubAgentsV1Warning for the Codex config
// here, as the Settings row asks it (the page can't reach the config itself).
// "" when there is no Codex, or no conflict.
func CodexSubAgentsV1Warning() string {
	a, err := Find("codex")
	if err != nil || a.Routed == nil {
		return ""
	}
	return codexSubAgentsV1Warning(a.Path, a.Routed())
}

func codexIn(at place) *Agent {
	dir := at.codexHome()
	path := filepath.Join(dir, "config.toml")

	a := &Agent{
		ID: "codex", Name: "Codex", Icon: "codex-color", Bin: "codex", Dir: dir, Path: path,
		UA: []string{"codex", "acp-extension-codex"}, // The ACP adapter's default client name becomes Codex's originator.

	}

	return a
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// codexChatGPT reports whether Codex is signed in to a ChatGPT account:
// its built-in provider then runs and asks for the model list at
// openai_base_url. Signed in with an API key alone, it asks for none.
func codexChatGPT(dir string) bool {
	signedIn, _ := codexChatGPTKnown(dir)
	return signedIn
}

// codexChatGPTKnown is codexChatGPT with whether the read said anything:
// known is false for an auth.json caught while Codex rewrites it in place
// (empty, cut off) or a keychain read that failed, which is no answer
// (provider.CodexSignInKnown). Wiring goes by known alone: taking such a
// read for "no ChatGPT sign-in" made magpie Codex's provider, and the
// Codex app showed "magpie" for its account and magpie's catalog for its
// models, on and off (lucy on Discord, v0.1.1394).
func codexChatGPTKnown(dir string) (signedIn, known bool) {
	if settings.AccountDiscoveryOff() {
		return false, true
	}
	// forced_login_method = "api" in Codex's own config (another tool or
	// guide writes it; magpie never does): Codex won't use a ChatGPT
	// sign-in then. The TUI logs it out on start and the app server reads
	// no account (openai/codex login/src/auth/manager.rs
	// enforce_login_restrictions, load_auth), so wired beside it the
	// Codex app asked for a sign-in and offered an API key alone (Tiger on
	// Discord: "重启 codex 之后登不回去了"). Counted as no sign-in, magpie
	// is the provider and Codex needs none. The line is the user's own and
	// stays: taking it out is how they sign in to ChatGPT again.
	if v, ok := edit.GetTOMLTop(filepath.Join(dir, "config.toml"), "forced_login_method"); ok && strings.EqualFold(strings.TrimSpace(v), "api") {
		return false, true
	}
	var a struct {
		Tokens *struct {
			Access string `json:"access_token"`
		} `json:"tokens"`
	}
	// auth.json, or the macOS keychain item when Codex's config.toml
	// keeps the sign-in there (#2367)
	b, ok, known := provider.CodexSignInKnown(dir)
	if !ok {
		return false, known
	}
	if json.Unmarshal(b, &a) != nil {
		// valid JSON that isn't an object: not a sign-in Codex wrote
		return false, false
	}
	if a.Tokens == nil || a.Tokens.Access == "" {
		return false, true
	}
	// a sign-in OpenAI refused to refresh for good (revoked, or failed for
	// good, #4308): Codex keeps its auth.json, but beside it every turn
	// fails at Codex's workspace routing discovery, magpie's models too.
	// None to magpie, as forced api above, until it is signed in again
	// (another refresh token). A record that can't be read says nothing.
	if gone, known := provider.CodexSignInGone(dir, b); !known || gone {
		return false, known
	}
	return true, true
}

// reachedFor is how long a read of Codex's log stands while its files say
// nothing changed: a file system whose times are coarse can miss a write
// that kept the size.
const reachedFor = 15 * time.Second

type fileStamp struct{ size, mod int64 }

func stampOf(path string) fileStamp {
	fi, err := os.Stat(path)
	if err != nil {
		return fileStamp{size: -1}
	}
	return fileStamp{size: fi.Size(), mod: fi.ModTime().UnixNano()}
}
