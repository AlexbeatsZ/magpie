package agent

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/proc"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/sessions"
	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/usage"
)

// The gateway knows each agent's requests by what this package says of it.
func init() {
	usage.Agents = func() []usage.Known {
		var out []usage.Known
		for _, a := range Clients() {
			out = append(out, usage.Known{ID: a.ID, Names: append([]string{a.ID}, a.Aliases...), UA: a.UA})
		}
		return out
	}
	// the Sessions page reads the sessions of the agents in WSL distros
	sessions.WSLHomes = wslHomes
	sessions.WSLRunning = WSLRunning
	sessions.WSLOff = func() bool { return !wslLooks() }
	// a Copilot CLI signed in inside a running distro (#723)
	provider.CopilotWSLDirs = copilotWSLDirs

}

// others are clients that reach the gateway without being agents magpie
// sets up: known only by their requests, to be drawn with a logo.
var others = []*Agent{
	{ID: "magpie", Name: "magpie", Icon: "magpie", UA: []string{"magpie"}},
	{ID: "curl", Name: "curl", Icon: "curl", UA: []string{"curl"}},
	// PI-Desktop (vastsa/PI-Desktop) runs Pi's own library, which says
	// "pi (…)"; given a User-Agent of its own in its provider's settings
	// ("pi-desktop/0.1.0"), it is told apart from Pi by that name, which
	// Pi's "pi-" would take otherwise (#1966, dgmico)
	{ID: "pi-desktop", Name: "PI-Desktop", Icon: "pi", UA: []string{"pi-desktop"}},
	// Kimi Code Desktop runs Kimi Code's engine on ~/.kimi-code, so it
	// reaches magpie through Kimi Code's config, under a client name of
	// its own ("kimi-code-desktop/…", jiabubu on Discord). Its icon is the
	// app's own, from KimiCode-1.2.1-win-x64.exe's build/icon.ico.
	{ID: "kimi-code-desktop", Name: "Kimi Code Desktop", Icon: "kimi-code-desktop", UA: []string{"kimi-code-desktop"}},
	// Raft (slock's raft-computer) runs Pi in each of its agents' folders;
	// Sessions lists those as Raft's (internal/sessions/raft.go), while
	// their requests say "pi (…)" and are Pi's
	{ID: "raft", Name: "Raft", Icon: "generic"},
}

// Clients is everyone whose requests the gateway knows by name: every
// agent, detected or not, and the others. It is what the Usage and
// Routing views draw a request's client with.
func Clients() []*Agent {
	var out []*Agent
	home, cfg := homes()
	for _, a := range onMachine(home, cfg) {
		// Keep request attribution without exposing config writers, launch
		// hooks or native integrations to callers of Clients.
		out = append(out, &Agent{ID: a.ID, Name: a.Name, Icon: a.Icon, Aliases: a.Aliases, UA: a.UA})
	}
	return append(out, others...)
}

// All is empty: this fork leaves agents' model configuration manual.
// Clients retains read-only request attribution.
// Resource management uses Inventory; profiles and model synchronization
// continue to see no writable agents here.
func All() []*Agent {
	return nil
}

// Inventory describes the agents the Library can give resources to. It uses
// the existing registry and detection rules, exposing no model configuration
// fields, launch commands or integration hooks.
func Inventory() []*Agent {
	home, cfg := homes()
	own := builtins(home, cfg)
	return append(own, pluggedAgents(home, own, nil)...)
}

// InventoryDetected lists installed resource targets without enabling model
// management through Detected.
func InventoryDetected() []*Agent {
	var out []*Agent
	for _, a := range Inventory() {
		if a.Detected() {
			out = append(out, a)
		}
	}
	return out
}

func homes() (home, cfg string) {
	home, _ = os.UserHomeDir()
	cfg = appdir.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		cfg = filepath.Join(home, ".config")
	}
	return home, cfg
}

// builtins are resource targets described by Magpie, including WSL installations.
func builtins(home, cfg string) []*Agent {
	return append(onMachine(home, cfg), wslAgents()...)
}

// local is All() without the agents in WSL distros, which wsl.exe is asked
// for: what start-up moves look in, on every command, `magpie version`
// too, and the moves came before WSL agents. A wsl.exe that hangs held each
// command 20s, its two listings' timeouts.
func local() []*Agent {
	return nil
}

// onMachine are builtins but those in WSL.
func onMachine(home, cfg string) []*Agent {
	return append([]*Agent{
		claude(home),
		claudeDesktop(home),
		claudeScience(),
		codex(home),
		gemini(home),
		agy(home),
		antigravity(home),
		opencode(home, cfg),
		openChamber(home, cfg),
		mimocode(home, cfg),
		deveco(home, cfg),
		pi(home),
		stepcode(home),
		mothx(home),
		penguin(home),
		aside(home),
		omo(home),
		primeAgent(home),
		goose(home, cfg),
		cursor(home),
		cursorLocal(),
		zed(home, cfg),
		vscode(home, cfg),
		vscodeInsidersAgent(home, cfg),
		vscodium(home, cfg),
		zooCode(home, cfg),
		copilotJetBrains(home),
		air(home, cfg),
		copilot(home),
		crush(home, cfg),
		dsh(home),
		reasonix(home),
		commandCode(home),
		fx(home),
		omp(home),
		devin(home, cfg),
		hermes(home),
		morph(home),
		kimi(home),
		qwen(home),
		muse(cfg),
		empryo(home),
		ante(home),
		astrbot(home),
		qwenpaw(home),
		openClaw(home),
		jcode(home),
		miniMax(home),
		droid(home),
		cline(home),
		qoder(home),
		qoderCN(home),
		grok(home),
		grokApp(home),
		zcode(home),
		workbuddy(home),
		workbuddyAI(home),
		codebuddy(home),
		pencil(home),
		openPencil(home, cfg),
		openBot(home, cfg),
		newMax(home),
		raycast(home),
		kiroCrew(home),
		t3code(home),
		hanako(home),
		atomcode(home),
		snow(home),
		alma(),
		cindy(),
		aionui(),
		lody(home),
	}, append(ompProfiles(home), qwenpawAgents(home)...)...)
}

// prefixed is an agent's Spelled whose values of magpie's are all
// magpie/<ref>, as its provider/model names them.
func prefixed(v string) bool { return strings.HasPrefix(v, magpieID+"/") }

func hostOf(u string) string { return provider.HostOf(u) }

// magpieProviderJSONAt is magpieProviderJSONFor for an agent that reaches
// the gateway at gw: one in a WSL distro under NAT (see wsl.go).
func magpieProviderJSONAt(shape, catalog, gw string) any {
	models := magpieModels(catalog) // the catalog is the agent's
	switch shape {
	case "opencode":
		// models in the order magpieModels has them (#4575): OpenCode lists
		// the models in the file's key order, and a map's keys marshal
		// sorted — the variants under each kept theirs since #713
		ms := orderedJSON{}
		for _, m := range models {
			e := map[string]any{"name": m.Name}
			// modalities alone says a model takes images: OpenCode 1 takes
			// its image input from them, and its attachment flag, which
			// nothing of OpenCode 1.18 reads, is a legacy field OpenCode 2
			// drops with a WARN per model on every load (#2209)
			if m.Images {
				e["modalities"] = map[string]any{"input": []string{"text", "image"}, "output": []string{"text"}}
			}
			// without it OpenCode doesn't know when to compact, and a
			// group's context (magpie group set … context=) never reaches
			// it; an output of 0 is OpenCode's own default
			if m.Context > 0 {
				e["limit"] = map[string]any{"context": m.Context, "output": maxTokens(m)}
			}
			if openCodeReasons(m) {
				e["reasoning"] = true
			}
			e["variants"] = openCodeVariants(m.Efforts, m.CodexFast)
			// OpenCode counts a session's cost by it, in USD per million
			// tokens (models.dev's cost; OpenCode 2.0.25 reads input,
			// output, cache_read, cache_write), and counted 0 without it
			if p := m.Price; p != nil {
				e["cost"] = map[string]any{"input": p.Input, "output": p.Output, "cache_read": p.CacheRead, "cache_write": p.CacheWrite}
			}
			ms = append(ms, jsonPair{m.ID, e})
		}
		return map[string]any{"npm": "@ai-sdk/openai-compatible", "name": "magpie",
			"options": map[string]any{"baseURL": gw + "/v1", "apiKey": openCodeKeyAt(catalog, gw)}, "models": ms}
	case "crush":
		var ms []map[string]any
		for _, m := range models {
			window := m.Context
			if window == 0 {
				window = 200000
			}
			// without it Crush caps every reply at 16384 tokens, a model
			// that can write more of them never asked for it; an output
			// above the window is cut to it, as it is for every other
			// agent magpie hands a limit to
			tokens := maxTokens(m)
			if tokens == 0 {
				tokens = 16384
			}
			cm := map[string]any{"id": m.ID, "name": m.Name, "context_window": window, "default_max_tokens": tokens,
				"can_reason": len(m.Efforts) > 0}
			// Crush counts a session's cost by it (catwalk's cost_per_1m_*,
			// Crush 0.39.3): in_cached is a cache write, out_cached a read
			if p := m.Price; p != nil {
				cm["cost_per_1m_in"], cm["cost_per_1m_out"] = p.Input, p.Output
				cm["cost_per_1m_in_cached"], cm["cost_per_1m_out_cached"] = p.CacheWrite, p.CacheRead
			}
			ms = append(ms, cm)
		}
		if ms == nil {
			ms = []map[string]any{}
		}
		return map[string]any{"type": "openai", "name": "magpie", "base_url": gw + "/v1", "api_key": keyAt(gw), "models": ms}
	case "pi":
		var ms []map[string]any
		for _, m := range models {
			ms = append(ms, piModelJSON(piAsked(catalog, m), gw, true))
		}
		if ms == nil {
			ms = []map[string]any{}
		}
		// Pi names its session to a provider only when asked to:
		// sendSessionAffinityHeaders has it send it on Chat
		// (x-session-affinity, session_id) and Anthropic's Messages API
		// (x-session-affinity). On Responses it sends session_id anyway.
		// Without it a turn of Pi's on Chat or Messages reached the gateway
		// with no session: its usage had none and "Record gateway
		// conversations" kept none of it (akic404 on Discord).
		return map[string]any{"name": "magpie", "baseUrl": gw + "/v1", "api": "openai-completions", "apiKey": keyAt(gw),
			"compat": map[string]any{"sendSessionAffinityHeaders": true}, "models": ms}
	}
	return nil
}

// piBlockKept is theirsKept for magpie's provider in a Pi models.json at
// path, where a sendSessionAffinityHeaders the file already has, the
// user's own or pi-cache-optimizer's (#1103), stays as it is.
func piBlockKept(path string, value func() any) func() any {
	return theirsKept(path, "providers."+magpieID, func() any {
		v := value()
		if _, set := edit.GetJSON(path, "providers."+magpieID+".compat.sendSessionAffinityHeaders"); !set {
			return v
		}
		if m, ok := v.(map[string]any); ok {
			if c, ok := m["compat"].(map[string]any); ok {
				delete(c, "sendSessionAffinityHeaders")
				if len(c) == 0 {
					delete(m, "compat")
				}
			}
		}
		return v
	}, "models")
}

// piAsked is m as agentID's models.json can ask it: without the Gemini API
// where nothing that reads the file can be trusted to stream
// google-generative-ai, so such a model stays on the provider's Chat,
// which the gateway translates.
//
//   - StepCode registers no google-generative-ai stream, only Pi's
//     anthropic-messages, openai-completions and openai-responses
//     (packages/providers/src/compat.ts BUILTIN_APIS, Step-Code at
//     90a57820), though its docs list it (#3282).
//   - Pi's own agent folder is read by every app that runs Pi's SDK on it,
//     not by Pi alone. Raft (slock's raft-daemon, the Pi runtime of its
//     agents) opens ~/.pi/agent/models.json (or PI_CODING_AGENT_DIR's) and
//     hands every request its own fetch (installProviderHttpClient), which
//     pi-ai's Google adapter refuses before sending: "Custom fetch is not
//     supported by the Google Generative AI adapter" (pi-ai 0.83.0 and
//     1.4.0, api/google-generative-ai.js; yunfei9958 on X). Per-request
//     fetch is a feature of Pi's SDK, so any other host of it can do the
//     same. OmO's and Prime Agent's folders are their own, and keep it.
func piAsked(agentID string, m catalog.Model) catalog.Model {
	if agentID == "stepcode" || agentID == "pi" {
		m.APIs = slices.DeleteFunc(slices.Clone(m.APIs), func(a string) bool { return a == string(provider.Gemini) })
	}
	return m
}

// piModelJSON is one of magpie's models as an entry of Pi's models.json.
// native selects the model's preferred gateway API (its own api and
// baseUrl); without it the model is left on the provider's
// openai-completions (Pencil, pencil.go).
func piModelJSON(m catalog.Model, gw string, native bool) map[string]any {
	// reasoning lets Pi offer its thinking levels for the model: as
	// /v1/models says it (provider.Entry.Thinks), so one that thinks with
	// no levels magpie knows (Trae CN's kimi-k3-max, #2380) is offered
	// Pi's own off to high, which the gateway fits to the model, where
	// it had off alone
	e := map[string]any{"id": m.ID, "name": m.Name, "reasoning": m.Reasoning || len(m.Efforts) > 0}
	// each model is asked on the API its provider speaks natively,
	// so the gateway relays what Pi sent as it is instead of
	// translating Chat. One served on OpenAI's Responses API alone,
	// or best there (a ChatGPT sign-in, GPT on OpenAI's API or
	// Copilot's), goes to baseUrl/responses; one on Anthropic's
	// Messages API alone to the gateway's /v1/messages (Anthropic's
	// SDK adds the /v1). A Claude that thinks only adaptively is
	// told so: Pi would otherwise ask it for a thinking budget,
	// which it turns away.
	switch {
	case !native:
	case slices.Contains(m.APIs, string(provider.Responses)):
		// strict left out of a tool of a turn asked there is taken as a
		// strict-mode schema, every property required, optional ones
		// filled (#3055): pi-ai says strict: false on an ordinary tool
		// only when the model's compat says so
		e["api"] = "openai-responses"
		e["compat"] = map[string]any{"supportsStrictMode": true}
	case slices.Contains(m.APIs, string(provider.Anthropic)):
		e["api"], e["baseUrl"] = "anthropic-messages", gw
		if gateway.AdaptiveThinking(m.ID) {
			e["compat"] = map[string]any{"forceAdaptiveThinking": true}
		}
	case slices.Contains(m.APIs, string(provider.Gemini)):
		e["api"], e["baseUrl"] = "google-generative-ai", gw+"/v1beta"
	}
	if m.Images {
		e["input"] = []string{"text", "image"}
	}
	if levels := piThinkingLevels(m.Efforts, e["api"] == "anthropic-messages" && gateway.ThinksOnlyWhenAsked(m.ID)); levels != nil {
		// on google-generative-ai Pi resolves a level through the map to
		// one of Gemini's four (pi-ai's resolveGoogleThinkingLevel) and
		// throws before sending on any other name, so a level the model
		// calls something else (xhigh, max) is hidden there (#3282)
		if e["api"] == "google-generative-ai" {
			for l, v := range levels {
				if s, ok := v.(string); ok && l != "off" && !slices.Contains([]string{"minimal", "low", "medium", "high"}, strings.ToLower(s)) {
					levels[l] = nil
				}
			}
		}
		e["thinkingLevelMap"] = levels
	}
	// without it Pi takes every model for a 128K one, and compacts
	// a 272K or 922K one long before it has to
	if m.Context > 0 {
		e["contextWindow"] = m.Context
	}
	// without it Pi caps every reply at 16384 tokens, a model
	// that can write 128K included
	if m.Output > 0 {
		e["maxTokens"] = maxTokens(m)
	}
	// Pi shows a call's cost by it, and weighs warming a model's cache by
	// it and promptCache (#781): without either it warms nothing
	if m.Price != nil {
		e["cost"] = map[string]any{"input": m.Price.Input, "output": m.Price.Output, "cacheRead": m.Price.CacheRead, "cacheWrite": m.Price.CacheWrite}
	}
	if c := piPromptCache(e["api"], m.ID); c != nil {
		e["promptCache"] = c
	}
	return e
}

// piPromptCache is how long, in seconds, the vendor keeps a model's prompt
// cache for each of Pi's retention tiers, at the short end of what it
// publishes; nil where magpie doesn't relay Pi's request as it is, or the
// vendor's lifetime isn't known. Claude on Anthropic's Messages API keeps
// one 5 minutes, or an hour asked with a 1h cache_control (Pi's long); GPT
// on OpenAI's Responses API keeps one in memory 5 to 10 minutes, and only
// some models offer the 24h retention, so long is left out. A Claude
// subscription runs through Claude Code, which caches on its own: no api is
// set for it, and none is declared.
func piPromptCache(api any, id string) map[string]any {
	name := strings.ToLower(id[strings.LastIndex(id, "/")+1:])
	switch {
	case api == "anthropic-messages" && strings.HasPrefix(name, "claude"):
		return map[string]any{"short": 300, "long": 3600}
	case api == "openai-responses" && strings.HasPrefix(name, "gpt"):
		return map[string]any{"short": 300}
	}
	return nil
}

// openCodeVariants are the reasoning levels OpenCode offers for a model of
// magpie's, each asking the gateway for that effort as reasoning_effort
// (@ai-sdk/openai-compatible's reasoningEffort). OpenCode 1.x offers none
// for a model the config doesn't mark as reasoning, and adds these. OpenCode
// 2 makes low, medium and high for every model of an openai-compatible
// provider whose config names no variants (packages/core/src/variant.ts,
// config/plugin/provider.ts), so a model whose levels are none/high/max, or
// go past high to xhigh and max, or that has none, was offered levels it
// doesn't have and not the ones it has. A model without levels gets an
// empty set, which OpenCode 2 takes as none rather than guessing. They are
// written weakest first: OpenCode and OpenChamber list a model's variants in
// the object's key order, and a map had them alphabetical — default, high,
// low, max, medium, ultra, xhigh (#713).
//
// A model Codex's /fast is asked of (catalog.Model.CodexFast) has each level
// again fast after them, low-fast to xhigh-fast, or fast alone when it has
// none (#2379): OpenCode has no /fast, and a variant is one pick, level and
// tier together.
func openCodeVariants(efforts []string, fast bool) orderedJSON {
	out := orderedJSON{}
	for _, e := range gateway.ByStrength(efforts) {
		if !slices.ContainsFunc(out, func(p jsonPair) bool { return p.k == e }) {
			out = append(out, jsonPair{e, openCodeVariant(e)})
		}
	}
	if !fast {
		return out
	}
	if len(out) == 0 {
		return orderedJSON{{openCodeFast, openCodeVariant(openCodeFast)}}
	}
	for _, p := range slices.Clone(out) {
		k := p.k + "-" + openCodeFast
		out = append(out, jsonPair{k, openCodeVariant(k)})
	}
	return out
}

// openCodeFast is the name of a fast variant, alone or after a level's.
const openCodeFast = "fast"

// openCodeVariant is the variant magpie writes under name: a level asked as
// reasoning_effort (@ai-sdk/openai-compatible's reasoningEffort, which
// OpenCode 2 reads too), and a fast one asks for priority processing too,
// in the words Codex's /fast sends it: service_tier "priority". Each
// OpenCode puts that on the request its own way: OpenCode 1 hands a variant's
// options to @ai-sdk/openai-compatible under the provider's name, which
// sends the ones it doesn't know, service_tier among them, as they are;
// OpenCode 2 sends what extraBody holds (packages/core/src/v1/config/
// provider-options.ts) and drops a key it doesn't know. Checked with
// OpenCode 2.0.25 against a fake server: service_tier alone, or
// serviceTier, never reached it; with extraBody it went as
// "service_tier":"priority" beside reasoning_effort. OpenCode 1 sends
// extraBody on as a field of its own, which the gateway drops when it asks
// the ChatGPT backend on Responses.
func openCodeVariant(name string) map[string]any {
	level, fast := strings.CutSuffix(name, "-"+openCodeFast)
	if name == openCodeFast {
		level, fast = "", true
	}
	v := map[string]any{}
	if level != "" {
		v["reasoningEffort"] = level
	}
	if fast {
		v["service_tier"] = "priority"
		v["extraBody"] = map[string]any{"service_tier": "priority"}
	}
	return v
}

// openCodeReasons is whether a model of magpie's is marked reasoning in
// opencode.json, which OpenCode 1 shows as 支持推理 in its model tooltip
// and takes for a model that thinks (#725: OpenCode Zen's
// mimo-v2.6-flash-free, which thinks with no levels to pick, said 不支持推理
// through magpie where it says it reasons under OpenCode's own Zen). OpenCode
// 1 also adds low, medium and high of its own to the variants of a reasoning
// model (ProviderTransform.variants; checked with 1.18.34), so one with levels
// is marked only when it has those three, else it would be offered levels it
// doesn't have; one with none gets them as it does under OpenCode's own
// provider, and the gateway asks a model with a thinking switch alone no more
// than high (fitFor). OpenCode 2 doesn't read the flag.
func openCodeReasons(m catalog.Model) bool {
	if len(m.Efforts) == 0 {
		return m.Reasoning
	}
	for _, e := range []string{"low", "medium", "high"} {
		if !slices.Contains(m.Efforts, e) {
			return false
		}
	}
	return true
}

// orderedJSON is a JSON object that keeps its keys in the order given,
// where a map's would come out sorted.
type orderedJSON []jsonPair

type jsonPair struct {
	k string
	v any
}

// modelEffort is the model's own name for level, one of pi-ai's (none for
// its off): level itself when efforts name it, else the one of efforts
// that names it in a vendor's words (catalog.EffortAliases: Trae's light
// and extra_high, as Our Free Model's plugin lists them), "" when it takes
// none such. dsh and Pi offered only the levels named in pi-ai's words, so
// such a model's light, high and extra_high showed in dsh as high alone
// (nianshou5 on Discord). The name returned is what the agent sends, so
// the vendor gets the level in the words it listed.
func modelEffort(efforts []string, level string) string {
	if slices.Contains(efforts, level) {
		return level
	}
	for _, e := range efforts {
		if catalog.EffortLevel(e) == level {
			return e
		}
	}
	return ""
}

// piThinkingLevels is the thinkingLevelMap for a model's efforts: every one
// of Pi's levels (piLevels, pi-ai's EXTENDED_THINKING_LEVELS), the model's
// own mapped to its name for it (modelEffort) and the others null. Pi
// builds /thinking from the map (getSupportedThinkingLevels): a level set
// to null is hidden and skipped, while one left out is offered (all but
// xhigh and max), so a map
// naming only max had Pi offer minimal and medium too, which the vendor
// turned away (#243). Pi's off is the model's none when it takes one. A
// model without none has off hidden — Pi's off asks a Responses model for
// effort none and leaves a Chat model on the vendor's default thinking —
// except a Claude on Anthropic's Messages API (claudeOff), where Pi's off
// sends thinking disabled, which Claude takes whatever its levels; there
// off is left out, for Pi to offer as before. Another vendor's model there
// keeps off hidden: without none it always thinks, and ZCode's
// GLM-5.3-Flash turned Pi's off away ("该模型始终支持思考，不可关闭", #699),
// as a level left out is one Pi offers. A model whose levels magpie doesn't
// know gets no map: Pi's own defaults, as before.
func piThinkingLevels(efforts []string, claudeOff bool) map[string]any {
	if len(efforts) == 0 {
		return nil
	}
	levels := map[string]any{}
	for _, l := range piLevels {
		w := l
		if l == "off" {
			w = "none"
		}
		if e := modelEffort(efforts, w); e != "" {
			levels[l] = e
		} else {
			levels[l] = nil
		}
	}
	if levels["off"] == nil && claudeOff {
		delete(levels, "off")
	}
	return levels
}

// openCodeLike is OpenCode and the forks that keep its config shape
// (mimocode): a provider block with npm/@ai-sdk/openai-compatible, model and
// small_model fields, and its own auth file beside its config.
// at is where it lives: this machine, or a WSL distro (see wsl.go), whose
// gateway address its provider names.
func openCodeLike(at place, id, name, icon, bin, dir, _ string, ua []string, aliases ...string) *Agent {
	// the first of its files there is, as the agent looks for them; with
	// none, a new <id>.json
	path := filepath.Join(dir, id+".json")
	for _, name := range []string{id + ".jsonc", id + ".json", "config.json"} {
		if at.exists(filepath.Join(dir, name)) {
			path = filepath.Join(dir, name)
			break
		}
	}

	return &Agent{
		ID: id, Name: name, Icon: icon, Aliases: aliases,
		UA:  ua,
		Bin: bin, Dir: dir, Path: path,
	}
}

// openCodeKeyAt is the key an OpenCode-shaped agent (id) sends the gateway
// at gw: the gateway's own, which OpenCode and MiMo Code are told apart by
// their User-Agents with, or for DevEco Code, parts of whose code still
// send OpenCode's User-Agent, one that names it.
func openCodeKeyAt(id, gw string) string {
	if id == "deveco" {
		return agentKeyAt(id, gw)
	}
	return keyAt(gw)
}

// sameGateway reports whether a base URL is magpie's gateway's (v1), its
// host spelled 127.0.0.1 or localhost.
func sameGateway(base, v1 string) bool {
	norm := func(u string) string {
		u = strings.TrimRight(strings.TrimSpace(u), "/")
		return strings.Replace(u, "://localhost:", "://127.0.0.1:", 1)
	}
	return base != "" && norm(base) == norm(v1)
}

func opencode(home, cfg string) *Agent {
	return openCodeLike(here(home), "opencode", "OpenCode", "opencode", "opencode",
		openCodeDir(cfg), filepath.Join(home, ".local", "share", "opencode", "auth.json"),
		[]string{"opencode"}, "oc")
}

// opencodeIn is OpenCode in a WSL distro (see wsl.go): its config in
// ~/.config/opencode, as XDG_CONFIG_HOME and OPENCODE_CONFIG_DIR, the
// distro's variables, aren't read; OpenChamber, a Windows app, isn't there.
func opencodeIn(at place) *Agent {
	return openCodeLike(at, "opencode", "OpenCode", "opencode", "opencode",
		at.join(".config", "opencode"), at.join(".local", "share", "opencode", "auth.json"),
		[]string{"opencode"}, "oc")
}

// mimocodeIn is MiMo Code in a WSL distro: ~/.config/mimocode, as
// MIMOCODE_HOME isn't read there.
func mimocodeIn(at place) *Agent {
	return openCodeLike(at, "mimocode", "MiMo Code", "mimocode", "mimo",
		at.join(".config", "mimocode"), at.join(".local", "share", "mimocode", "auth.json"),
		[]string{"mimocode"}, "mimo")
}

// openCodeDir is the folder OpenCode reads its user config from:
// $OPENCODE_CONFIG_DIR when set, else $XDG_CONFIG_HOME/opencode (cfg).
// OpenCode 2, the one OpenChamber bundles (#321), reads that folder in place
// of the other; OpenCode 1 reads both, the variable's last, so what magpie
// writes there wins in either.
func openCodeDir(cfg string) string {
	if d := strings.TrimSpace(appdir.Getenv("OPENCODE_CONFIG_DIR")); d != "" {
		if abs, err := filepath.Abs(d); err == nil {
			return abs
		}
	}
	return filepath.Join(cfg, "opencode")
}

// mimocode is MiMo Code, the CLI, and the engine inside Xiaomi MiMo, the
// desktop app (#249): both read the same config, which MIMOCODE_HOME moves
// to its own config folder.
func mimocode(home, cfg string) *Agent {
	dir := filepath.Join(cfg, "mimocode")
	if h := appdir.Getenv("MIMOCODE_HOME"); filepath.IsAbs(h) {
		dir = filepath.Join(h, "config")
	}
	return openCodeLike(here(home), "mimocode", "MiMo Code", "mimocode", "mimo",
		dir, filepath.Join(home, ".local", "share", "mimocode", "auth.json"),
		[]string{"mimocode"}, "mimo")
}

// deveco is DevEco Code, Huawei's fork of OpenCode for HarmonyOS apps
// (openharmony-sig/deveco-code, #1885). It keeps OpenCode's config shape
// in its own folder, $XDG_CONFIG_HOME/deveco, merging config.json,
// deveco.json and deveco.jsonc there (packages/opencode/src/config,
// loadGlobal); DEVECO_CONFIG_DIR adds a folder read after it, as
// OPENCODE_CONFIG_DIR does OpenCode's, so what magpie writes there wins.
// Its auth is xdg-basedir's data folder's, ~/.local/share/deveco. 0.2.3's
// requests to a provider of the config's say "deveco/0.2.3 ai-sdk/…", but
// much of the OpenCode code it keeps still sends OpenCode's own User-Agent
// (opencode/<version>), so its key names it as well (openCodeKeyAt).
func deveco(home, cfg string) *Agent {
	return openCodeLike(here(home), "deveco", "DevEco Code", "deveco-color", "deveco",
		devecoDir(cfg), filepath.Join(home, ".local", "share", "deveco", "auth.json"), []string{"deveco"})
}

// devecoDir is the folder DevEco Code reads its user config from:
// $DEVECO_CONFIG_DIR when set to an absolute folder, else
// $XDG_CONFIG_HOME/deveco (cfg).
func devecoDir(cfg string) string {
	if d := strings.TrimSpace(appdir.Getenv("DEVECO_CONFIG_DIR")); filepath.IsAbs(d) {
		return d
	}
	return filepath.Join(cfg, "deveco")
}

func pi(home string) *Agent { return piIn(here(home)) }

// piIn is Pi as it lives at a place: this machine's home, or a WSL
// distro's (see wsl.go), its models.json naming the gateway as it reaches
// it from there.
func piIn(at place) *Agent {
	a := piLike(at, "pi", "Pi", piDir(at))
	a.UA = []string{"pi-"}
	return a
}

// piDir is Pi's agent folder at a place: PI_CODING_AGENT_DIR's when set,
// "~" in it standing for home, else ~/.pi/agent (config.js, getAgentDir).
// A relative one is Pi's working directory's, which magpie can't know, so
// it is not taken; nor this machine's variable for a WSL distro's Pi.
func piDir(at place) string {
	if at.spell == nil {
		if d := homeDir(at.home, appdir.Getenv("PI_CODING_AGENT_DIR")); d != "" {
			return d
		}
	}
	return at.join(".pi", "agent")
}

// homeDir is the folder an agent's variable names, "~" and "~/…" expanded
// to home as Pi does; "" when it is empty or relative.
func homeDir(home, d string) string {
	if d == "~" || strings.HasPrefix(d, "~/") || (runtime.GOOS == "windows" && strings.HasPrefix(d, `~\`)) {
		d = filepath.Join(home, d[1:])
	}
	if !filepath.IsAbs(d) {
		return ""
	}
	return filepath.Clean(d)
}

// beforeMagpie is a copy of an agent's own files taken before magpie first
// writes them (file.before-magpie beside each), which Disconnect puts back
// byte for byte when what it left says the same as the copy: the same
// JSON, in the user's own order, spacing and comments-free layout. Where
// they differ (the user changed something while connected), what
// Disconnect left stands. A file that wasn't there (named in the stash
// under absent) goes again when nothing but empty objects is left in it.
type beforeMagpie struct {
	absent string
	files  []string
}

// restore puts back what take kept, where Disconnect left the same.
func (b beforeMagpie) restore() error {
	gone := strings.Split(unstash(b.absent), "\n")
	for _, f := range b.files {
		cur, err := os.ReadFile(f)
		if err != nil {
			os.Remove(f + ".before-magpie")
			continue
		}
		var now any
		if json.Unmarshal(edit.TrimBOM(cur), &now) != nil { // a BOM is the user's own (#1534)
			continue
		}
		if was, err := os.ReadFile(f + ".before-magpie"); err == nil {
			var then any
			if json.Unmarshal(edit.TrimBOM(was), &then) == nil && reflect.DeepEqual(now, then) && !bytes.Equal(cur, was) {
				if err := edit.WriteAtomic(f, was); err != nil {
					return err
				}
			}
			os.Remove(f + ".before-magpie")
			continue
		}
		if slices.Contains(gone, filepath.Base(f)) && emptyJSON(now) {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}
	return nil
}

// emptyJSON says v is an object holding nothing but empty objects.
func emptyJSON(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	for _, x := range m {
		if !emptyJSON(x) {
			return false
		}
	}
	return true
}

// piLike is Pi, or a fork of it that keeps Pi's settings.json and
// models.json in an agent folder of its own (OmO, omo.go): id is its id,
// icon and command, dir its agent folder.
func piLike(at place, id, name, dir string) *Agent {
	return piLikeKeyed(at, id, name, dir, at.gwKey)
}

// piLikeKeyed is piLike whose provider "magpie" carries key's key: one
// that names the agent (agentKeyAt) for an agent whose requests carry
// nothing else of it (Prime Agent, primeagent.go).
func piLikeKeyed(_ place, id, name, dir string, _ func() string) *Agent {
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: id, Name: name, Icon: id, Bin: id, Dir: dir, Path: path,
	}
}

func goose(home, cfg string) *Agent { return gooseIn(here(home), cfg) }

// gooseIn is Goose at a place whose config folder is cfg: this machine's,
// or a WSL distro's ~/.config (see wsl.go), where Goose keeps Linux's
// layout and its custom provider names the gateway as the distro reaches
// it, with the key it takes from there.
func gooseIn(at place, cfg string) *Agent {
	path := filepath.Join(at.open(filepath.Join(cfg, "goose")), "config.yaml")
	if runtime.GOOS == "windows" && at.spell == nil {
		if app := appdir.Getenv("APPDATA"); app != "" {
			path = filepath.Join(app, "Block", "goose", "config", "config.yaml")
		}
	}

	return &Agent{
		ID: "goose", Name: "Goose", Icon: "goose", Bin: "goose", Dir: filepath.Dir(path), Path: path,
		UA: []string{"goose"},

		// a goose on PATH may be pressly's database migration tool, a Go
		// program; Block's goose is Rust, so a Go goose is not the agent
		detect: func() bool {
			if agentDir(filepath.Dir(path)) {
				return true
			}
			bin, err := proc.Look("goose")
			return err == nil && !goProgram(bin)
		},
	}
}

func cursor(home string) *Agent {
	path := filepath.Join(home, ".cursor", "cli-config.json")
	return &Agent{
		// Cursor CLI, not Cursor: the row is cursor-agent, and the Cursor
		// editor is not on it (@G_Z_Y on X looked for Cursor CLI on the
		// Agents page and took "Cursor" for the editor)
		ID: "cursor", Name: "Cursor CLI", Icon: "cursor", Aliases: []string{"cursor-agent"},
		UA:  []string{"cursor"},
		Bin: "cursor-agent", Dir: filepath.Dir(path), Path: path,
		// the editor reads ~/.cursor's mcp.json, skills and hooks, and its
		// chats are listed with the CLI's
		shared: "Cursor",
		// the Cursor editor makes ~/.cursor (argv.json, extensions, mcp.json,
		// projects) with no cursor-agent anywhere, so the folder says nothing
		// of the CLI: cli-config.json or the cursor-agent command does
		// (@G_Z_Y on X saw a Cursor row that wasn't the CLI)
		dirShared: true,
	}
}

func copilot(home string) *Agent {
	dir := filepath.Join(home, ".copilot")
	path := filepath.Join(dir, "settings.json")
	return &Agent{
		ID: "copilot", Name: "Copilot CLI", Icon: "githubcopilot", Aliases: []string{"gh-copilot"},
		UA:  []string{"copilot", "github-copilot"},
		Bin: "copilot", Dir: dir, Path: path,
		// GitHub Copilot's JetBrains and VS Code plugins write ~/.copilot
		// too (agents, skills, hooks, instructions, prompts, ide), so the
		// folder says nothing of the CLI: its settings.json or the copilot
		// command does (#1495)
		dirShared: true,
		// VS Code 1.143 ships a copilot launcher of its own in
		// bin\copilot-shim in its install folder, which is on the user's
		// PATH, and running it only asks for the CLI to be installed
		// (#3269): a command of the editor's is not the CLI
		binNot: editorLauncher,
	}
}

// editorLauncher reports whether bin is a launcher an editor ships beside
// itself, not the CLI it is named after (#3269): an editor's install is
// told by the product.json it keeps beside its program, as
// vscodeCodeBinary tells one. VS Code 1.143 puts its copilot launcher in
// <install>\bin\copilot-shim, with <install>\resources\app\product.json
// in a Windows or Linux install; a Mac's app keeps bin in
// Contents/Resources/app, beside product.json. So only bin's own folder
// and the two above it are looked in, never further up, where a folder of
// the user's (a home, a checkout of an editor's source) may hold a
// product.json of its own. A link on PATH is judged where it points.
// A copilot a user installed itself, in npm's global folder, Homebrew's,
// winget's Links or scoop's shims, is still the CLI.
//
// VS Code's versioned Windows layout (1.139's win32VersionedUpdate) moves
// that product.json one level under the install root: the app's payload
// lives in a folder named for the build's commit — the short hash, ten
// lowercase hex digits, 2c33b50b16 for 1.144.0 — with resources\app under
// it and bin beside it (microsoft/vscode#343045; Tony-ooo's 1.144.0, whose
// every product.json the six places above miss, #3269). A folder of the
// user's beside bin is not named as a commit's short hash is, so the cap
// keeps its meaning: nothing wider than the same three folders is read.
func editorLauncher(bin string) bool {
	if real, err := filepath.EvalSymlinks(bin); err == nil {
		bin = real
	}
	dir, err := filepath.Abs(filepath.Dir(bin))
	if err != nil {
		return false
	}
	for range 3 {
		if isFile(filepath.Join(dir, "product.json")) ||
			isFile(filepath.Join(dir, "resources", "app", "product.json")) {
			return true
		}
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if e.IsDir() && commitHashName(e.Name()) &&
					isFile(filepath.Join(dir, e.Name(), "resources", "app", "product.json")) {
					return true
				}
			}
		}
		up := filepath.Dir(dir)
		if up == dir {
			break
		}
		dir = up
	}
	return false
}

// commitHashName is whether a folder is named as a build's short commit
// hash is: ten lowercase hexadecimal digits (16507d91be, 2c33b50b16).
func commitHashName(name string) bool {
	if len(name) != 10 {
		return false
	}
	for _, c := range []byte(name) {
		if c < '0' || c > '9' && c < 'a' || c > 'f' {
			return false
		}
	}
	return true
}

func crush(home, cfg string) *Agent {
	path := filepath.Join(cfg, "crush", "crush.json")
	// Crush saves the models its own picker chooses to its data file, and
	// reads that after crush.json, so a pick there wins over one in path.
	// The picks are read as Crush merges the two files and written where
	// Crush writes them; magpie's provider stays in path, beside the
	// library's MCP servers.
	data := filepath.Join(home, ".local", "share", "crush", "crush.json")
	if dir := appdir.Getenv("XDG_DATA_HOME"); dir != "" {
		data = filepath.Join(dir, "crush", "crush.json")
	}
	if runtime.GOOS == "windows" {
		if app := appdir.Getenv("LOCALAPPDATA"); app != "" {
			path = filepath.Join(app, "crush", "crush.json")
		}
		// %LOCALAPPDATA%\crush is Crush's data folder there
		data = path
	}
	return crushAt(here(home), path, data)
}

// crushIn is Crush in a WSL distro: ~/.config/crush/crush.json, Linux's
// place for it.
func crushIn(at place) *Agent {
	// Crush's data file is in Linux's place there too
	return crushAt(at, at.join(".config", "crush", "crush.json"),
		at.join(".local", "share", "crush", "crush.json"))
}

// crushAt is Crush with its config at path and its data file at data,
// reaching the gateway as at does.
func crushAt(_ place, path, _ string) *Agent {

	return &Agent{
		ID: "crush", Name: "Crush", Icon: "crush", Bin: "crush", Dir: filepath.Dir(path), Path: path,
		UA: []string{"crush"},
	}
}

// readOwnAuth is os.ReadFile for an agent's own sign-ins (OpenCode's
// auth.json and its kin), left unread with account discovery off (#1535).
func readOwnAuth(path string) ([]byte, error) {
	if settings.AccountDiscoveryOff() {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(path)
}
