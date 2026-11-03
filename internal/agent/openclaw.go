package agent

// OpenClaw (openclaw/openclaw, a personal assistant that answers on
// WhatsApp, Telegram, Discord and the rest) keeps its settings in one file,
// openclaw.json. Where it is follows OpenClaw's own resolveConfigPath
// (src/config/paths.ts) and resolveStateDir (src/config/state-dir.ts):
// $OPENCLAW_CONFIG_PATH, else openclaw.json in $OPENCLAW_STATE_DIR, else in
// ~/.openclaw-<profile> for a named $OPENCLAW_PROFILE, else in ~/.openclaw,
// where ~ is $OPENCLAW_HOME when it is set.
//
// A provider of the user's own is an entry of models.providers
// (src/config/zod-schema.core.ts, ModelProviderSchema): a baseUrl, an
// apiKey, an api (openai-completions is OpenAI's chat completions) and its
// models, each an id and a name, with input, reasoning and contextWindow
// (ModelDefinitionSchema). The schema is strict: a key OpenClaw doesn't
// know makes the file invalid, so magpie writes only those. A model is
// named provider/model, split at the first slash
// (src/agents/model-selection-normalize.ts), so magpie's
// magpie/<provider>/<model> is OpenClaw's ref as it is. The model OpenClaw
// runs on is agents.defaults.model, a ref or {primary, fallbacks}
// (docs/concepts/models.md); `openclaw models set` sets primary and keeps
// the fallbacks, and adds the ref to agents.defaults.models
// (src/commands/models/shared.ts), which a config from before 2026.8.1
// still reads as the models allowed. A non-empty
// agents.defaults.modelPolicy.allow is the models allowed: magpie adds
// "magpie/*" to it when nothing of magpie's is in it.
//
// magpie adds itself as the provider "magpie", the gateway's /v1 with a key
// naming OpenClaw and one model entry per catalog model, and makes the model
// picked the primary. What was the primary is stashed and put back on
// Disconnect, while magpie's is still the primary. The user's own keys stay
// as written, in magpie's provider and its model entries too (headers,
// timeoutSeconds, a model's maxTokens). OpenClaw writes the file as JSON
// (src/config/io.write.ts) but reads JSON5; a file that isn't JSON magpie
// can read is left alone. OpenClaw's gateway takes a change to models and
// the default model without a restart (docs/gateway/configuration/
// hot-reload.md), but a chat already pinned to a model keeps it until
// /model default.
//
// Not wired: a per-agent model (agents.entries.*.model) stays the user's,
// and so does a config split with $include.

import (
	"path/filepath"
	"regexp"
	"strings"
)

// openClawProfile is a profile name OpenClaw takes (profile-utils.ts).
var openClawProfile = regexp.MustCompile(`(?i)^[a-z0-9][a-z0-9_-]{0,63}$`)

// openClawHome is OpenClaw's ~: $OPENCLAW_HOME, a leading ~ in it the
// system's home (resolveEffectiveHomeDir).
func openClawHome(at place) string {
	h := strings.TrimSpace(at.getenv("OPENCLAW_HOME"))
	if h == "" || h == "undefined" || h == "null" {
		return at.home
	}
	return openClawTilde(h, at.home)
}

// openClawTilde is p with a leading ~ as home.
func openClawTilde(p, home string) string {
	if p == "~" {
		return home
	}
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(home, rest)
	}
	if rest, ok := strings.CutPrefix(p, `~\`); ok {
		return filepath.Join(home, rest)
	}
	return p
}

// openClawPaths are OpenClaw's state folder and its openclaw.json.
func openClawPaths(at place) (dir, path string) {
	home := openClawHome(at)
	dir = filepath.Join(home, ".openclaw")
	if p := strings.TrimSpace(at.getenv("OPENCLAW_PROFILE")); p != "" && !strings.EqualFold(p, "default") && openClawProfile.MatchString(p) {
		dir = filepath.Join(home, ".openclaw-"+p)
	}
	if d := strings.TrimSpace(at.getenv("OPENCLAW_STATE_DIR")); d != "" {
		dir = openClawTilde(d, home)
	}
	path = filepath.Join(dir, "openclaw.json")
	if p := strings.TrimSpace(at.getenv("OPENCLAW_CONFIG_PATH")); p != "" {
		path = openClawTilde(p, home)
	}
	return dir, path
}

func openClaw(home string) *Agent { return openClawAt(here(home)) }

func openClawAt(at place) *Agent {
	dir, path := openClawPaths(at)

	return &Agent{
		ID: "openclaw", Name: "OpenClaw", Icon: "openclaw",
		Bin: "openclaw", Dir: dir, Path: path,
	}
}
