package agent

// omp (oh-my-pi, a fork of Pi) keeps its settings in ~/.omp/agent/config.yml
// (or where ompDir says its variables move it),
// the model of each role under modelRoles as "provider/model", and providers
// of the user's own in models.yml beside it. magpie adds itself there as the
// provider "magpie", keyless (auth: none) on loopback, with the catalog as its models; a
// model through magpie is "magpie/<provider>/<model>", which omp matches
// whole against provider/id. The other roles, the fallback chains and the
// like may name them as well (ompRefKeys); magpie stays while any does.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/dirent"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/provider"
)

// ompEfforts are the thinking levels omp knows.
var ompEfforts = []string{"minimal", "low", "medium", "high", "xhigh", "max"}

// ompProfileName is a profile name omp takes (pi-utils dirs.ts,
// normalizeProfileName); it refuses any other.
var ompProfileName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// ompReserved is a Windows device name omp refuses on every platform
// (pi-utils dirs.ts, normalizeProfileName): CON, PRN, AUX, NUL, COM0–COM9,
// LPT0–LPT9, and the same with a dot and anything after (con.work).
var ompReserved = regexp.MustCompile(`(?i)^(?:CON|PRN|AUX|NUL|COM[0-9]|LPT[0-9])(?:\..*)?$`)

// ompProfileOK says whether name is one omp takes as a profile (pi-utils
// dirs.ts): "default" is none, a trailing dot is refused, and so is a
// Windows reserved device name. What a profiles folder holds is checked
// against this, and so is what a distro's probe finds, and the profile
// OMP_PROFILE names for the default row.
func ompProfileOK(name string) bool {
	return name != "" && name != "default" && ompProfileName.MatchString(name) &&
		!strings.HasSuffix(name, ".") && !ompReserved.MatchString(name)
}

// ompRoot is where omp keeps its agent folder and its profiles: ~/.omp, or
// ~/$PI_CONFIG_DIR when that variable renames it.
func ompRoot(home string) string {
	if d := appdir.Getenv("PI_CONFIG_DIR"); d != "" {
		return filepath.Join(home, d)
	}
	return filepath.Join(home, ".omp")
}

// ompDir is omp's agent folder, found as omp's pi-utils (dirs.ts) finds it:
// under ~/.omp, or ~/$PI_CONFIG_DIR; a profile's (OMP_PROFILE, else
// PI_PROFILE; "default" is none) is profiles/<name>/agent there; with none,
// PI_CODING_AGENT_DIR — the variable Pi reads — moves it, else it is
// agent there. omp takes that variable as given, without expanding "~".
//
// shared says the folder is one omp has no claim to: PI_CODING_AGENT_DIR
// moved it, and Pi reads that variable too, so Pi's own folder may be what
// it names.
func ompDir(home string) (dir string, shared bool) {
	root := ompRoot(home)
	p, set := appdir.LookupEnv("OMP_PROFILE")
	if !set {
		p = appdir.Getenv("PI_PROFILE")
	}
	if p = strings.TrimSpace(p); ompProfileOK(p) {
		return filepath.Join(root, "profiles", p, "agent"), false
	}
	if d := appdir.Getenv("PI_CODING_AGENT_DIR"); filepath.IsAbs(d) {
		return filepath.Clean(d), true
	}
	return filepath.Join(root, "agent"), false
}

func omp(home string) *Agent {
	dir, shared := ompDir(home)
	a := ompAt(here(home), dir, func() ompProviderEntry { return ompProvider() })
	// Pi's folder is no sign of omp: its config.yml or its command is
	a.dirShared = shared
	return a
}

// ompProfiles are the named omp profiles under ~/.omp/profiles (or
// ~/$PI_CONFIG_DIR/profiles): each is an omp of its own, omp#<name>, with
// its own config.yml and models.yml in profiles/<name>/agent. The one the
// default row already picks (OMP_PROFILE) is left out, so it is not listed
// twice, and the name rules are omp's own. A missing profiles folder is no
// profiles, not an error; a profile whose agent folder has no config.yml
// yet is still a row, and the first Set makes the file.
func ompProfiles(home string) []*Agent {
	root := ompRoot(home)
	entries, err := os.ReadDir(filepath.Join(root, "profiles"))
	if err != nil {
		return nil
	}
	cur, _ := ompDir(home)
	var names []string
	for _, e := range entries {
		name := e.Name()
		if !dirent.IsFolder(filepath.Join(root, "profiles"), e) || !ompProfileOK(name) {
			continue
		}
		if filepath.Join(root, "profiles", name, "agent") == cur {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	var out []*Agent
	for _, name := range names {
		a := ompAt(here(home), filepath.Join(root, "profiles", name, "agent"), func() ompProviderEntry { return ompProvider() })
		a.ID, a.Name = "omp#"+name, "omp · "+name
		// the default row keeps omp's name, its command and its own lists
		a.Aliases, a.UA, a.Bin = nil, nil, ""
		profile := filepath.Join(root, "profiles", name)
		a.detect = func() bool { return isDir(profile) }
		out = append(out, a)
	}
	return out
}

// ompIn is omp in a WSL distro (see wsl.go): ~/.omp/agent, as the
// distro's variables that move it aren't read. The omp there isn't the one
// on Windows' PATH: its version is the one the distro's probe asked it
// (whqtian on Discord: a WSL omp's max became xhigh while its version wasn't
// known), and without one its models offer xhigh rather than a max an older
// omp would refuse.
func ompIn(at place) *Agent {
	return ompAt(at, at.join(".omp", "agent"), func() ompProviderEntry { return ompProviderAt(at.gw(), at.version) })
}

// ompAt is omp with its agent folder at dir, magpie's entry in its
// models.yml being entry's.
func ompAt(at place, dir string, _ func() ompProviderEntry) *Agent {

	// omp reads the .yml and falls back to the .yaml
	pick := func(name string) string {
		yml := filepath.Join(dir, name+".yml")
		if !at.exists(yml) && at.exists(filepath.Join(dir, name+".yaml")) {
			return filepath.Join(dir, name+".yaml")
		}
		return yml
	}
	path := pick("config")

	return &Agent{
		ID: "omp", Name: "omp", Icon: "omp", Aliases: []string{"oh-my-pi"},
		UA:  []string{"oh-my-pi"},
		Bin: "omp", Dir: dir, Path: path,
	}
}

type ompModel struct {
	ID        string       `yaml:"id"`
	Name      string       `yaml:"name,omitempty"`
	API       string       `yaml:"api,omitempty"`
	BaseURL   string       `yaml:"baseUrl,omitempty"`
	Reasoning bool         `yaml:"reasoning"`
	Thinking  *ompThinking `yaml:"thinking,omitempty"`
	Context   int          `yaml:"contextWindow,omitempty"`
	MaxTokens int          `yaml:"maxTokens,omitempty"`
	Input     []string     `yaml:"input,omitempty"`
	// Cost is the model's price in USD per million tokens, the unit omp's
	// calculateUsageCost divides by (omp 18.9.1, pi-catalog
	// src/models.ts). omp's custom provider prices a call by it alone;
	// without it omp takes a bundled model's its fuzzy id match finds, and
	// a group/* id matches none, so its calls cost 0 (taotao7 on Discord).
	Cost *ompCost `yaml:"cost,omitempty"`
}

// ompCost is a model's cost in omp's models.yml (omp 18.9.1,
// models-config-schema-bundle.ts: input, output, cacheRead, cacheWrite).
type ompCost struct {
	Input      float64 `yaml:"input"`
	Output     float64 `yaml:"output"`
	CacheRead  float64 `yaml:"cacheRead"`
	CacheWrite float64 `yaml:"cacheWrite"`
}

type ompThinking struct {
	Mode    string   `yaml:"mode"`
	Efforts []string `yaml:"efforts"`
}

type ompProviderEntry struct {
	BaseURL string `yaml:"baseUrl"`
	API     string `yaml:"api"`
	// Auth is none on loopback, where the gateway asks for no key; APIKey,
	// for an omp that reaches it from beyond (a WSL distro under NAT), the
	// key the gateway is shared on the network with
	Auth   string `yaml:"auth,omitempty"`
	APIKey string `yaml:"apiKey,omitempty"`
	// Headers name omp to the gateway: omp 16.x asks with Bun's User-Agent
	// and only a later one with its own (omp/18.4.4), so its requests went
	// to "Bun" in usage and past omp's own rules and stand-ins
	Headers map[string]string `yaml:"headers,omitempty"`
	// Compat names the header omp sends its session in. omp names the
	// conversation on Anthropic's Messages API (X-Claude-Code-Session-Id),
	// and on Chat and Responses only to OpenAI itself, so a turn of omp's on
	// Chat or Responses reached the gateway with no session: its usage had
	// none and "Record gateway conversations" kept none of it (akic404 on
	// Discord). promptCacheSessionHeader has omp send its session id in it.
	Compat map[string]string `yaml:"compat,omitempty"`
	Models []ompModel        `yaml:"models"`
}

// ompSessionSince is the first omp whose compat takes
// promptCacheSessionHeader (pi-catalog 16.0.6).
const ompSessionSince = "16.0.6"

// ompMaxSince is the first omp whose models.yml takes max as a thinking
// effort (pi-ai 16.4.0); 16.3.5's schema stops at xhigh and turns the whole
// file away over a max.
const ompMaxSince = "16.4.0"

// ompVersion is the version of the omp on PATH, "" when not known; a var so
// tests can fake it.
var ompVersion = func() string { return (&Agent{ID: "omp", Bin: "omp"}).InstalledVersion() }

// ompTakesMax says whether omp at version v takes max in models.yml. One
// whose version isn't known is taken for an older one: a max it refuses
// costs every model magpie gives it, an xhigh in its place only the top
// level of a model that has both.
func ompTakesMax(v string) bool {
	return v != "" && !Newer(ompMaxSince, v)
}

// ompProvider is magpie's entry in models.yml. The thinking efforts are the
// levels omp offers for the model; on Chat it sends them as
// reasoning_effort.
//
// As for Pi, each model is asked on the API its provider speaks natively, so
// the gateway relays what omp sent as it is instead of translating Chat: a
// Claude over Chat lost its thinking's signatures between tool turns, as
// Chat has no place for them. One served on OpenAI's Responses API goes to
// baseUrl/responses; one on Anthropic's Messages API to the gateway's
// /v1/messages (omp adds the /v1). That one thinks adaptively when it takes
// nothing else, else on a budget: omp's anthropic-budget-effort would also
// send output_config.effort, which Sonnet 4.5 and Haiku 4.5 refuse. One on
// Gemini's API goes to the gateway's /v1beta on google-generative-ai
// (#4277, ompGeminiMode).
func ompProvider() ompProviderEntry { return ompProviderAt(gateway.URL(), ompVersion()) }

// ompProviderAt is ompProvider for an omp of version (as ompVersion) that
// reaches the gateway at gw.
func ompProviderAt(gw, version string) ompProviderEntry {
	takesMax := ompTakesMax(version)
	ms := []ompModel{}
	for _, m := range sharedModels("omp") {
		e := ompModel{ID: m.ID, Name: m.Name, Context: m.Context, MaxTokens: maxTokens(m)}
		if p := m.Price; p != nil {
			e.Cost = &ompCost{Input: p.Input, Output: p.Output, CacheRead: p.CacheRead, CacheWrite: p.CacheWrite}
		}
		mode := "effort"
		switch {
		case slices.Contains(m.APIs, string(provider.Responses)):
			e.API = "openai-responses"
		case slices.Contains(m.APIs, string(provider.Anthropic)):
			e.API, e.BaseURL = "anthropic-messages", gw
			mode = "budget"
			if gateway.AdaptiveThinking(m.ID) {
				mode = "anthropic-adaptive"
			}
		case slices.Contains(m.APIs, string(provider.Gemini)):
			// #4277: a Gemini model on Gemini's own API at the gateway's
			// /v1beta, as OmO's; omp posts to baseUrl/models/<id>:
			// streamGenerateContent?alt=sse with x-goog-api-key, the key
			// being "N/A" (kNoAuth) under auth: none, which loopback takes
			// (pi-ai providers/google.ts, 16.0.0 to 18.14.0)
			e.API, e.BaseURL = "google-generative-ai", gw+"/v1beta"
			mode = ompGeminiMode(m.ID)
		}
		// Without input omp takes it from a bundled model its fuzzy id match
		// finds, else text only; a model the source never answered for is
		// left to that guess.
		switch {
		case m.Images:
			e.Input = []string{"text", "image"}
		case m.ImageInput != nil:
			e.Input = []string{"text"}
		}
		var efforts []string
		for _, x := range ompEfforts { // in omp's order
			// before omp 16.4.0 a model's efforts stop at xhigh (16.3.5
			// turns the whole file away over a max): a model whose top is
			// max offers xhigh, which the gateway fits to max when the
			// model has no xhigh of its own
			if x == "max" && !takesMax {
				continue
			}
			if slices.Contains(m.Efforts, x) || x == "xhigh" && !takesMax && slices.Contains(m.Efforts, "max") {
				efforts = append(efforts, x)
			}
		}
		if len(efforts) > 0 {
			e.Reasoning = true
			e.Thinking = &ompThinking{Mode: mode, Efforts: efforts}
		}
		ms = append(ms, e)
	}
	e := ompProviderEntry{BaseURL: gw + "/v1", API: "openai-completions", Auth: "none",
		Headers: map[string]string{"User-Agent": "omp"}, Models: ms}
	if version != "" && !Newer(ompSessionSince, version) {
		e.Compat = map[string]string{"promptCacheSessionHeader": "x-session-affinity"}
	}
	// whqtian on Discord: a WSL omp under NAT asks Windows' address, where
	// the gateway turns a request without a named key away
	if key := keyAt(gw); key != gateway.Token {
		e.Auth, e.APIKey = "", key
	}
	return e
}

// ompGeminiMode is how omp asks a Gemini model on google-generative-ai
// to think. Gemini 3 and later take thinkingConfig.thinkingLevel
// (google-level: omp sends minimal to high as MINIMAL to HIGH, xhigh and max
// as HIGH); Gemini 2.x and earlier take a thinkingBudget and refuse a level,
// so they go on budget, which omp sends as -1, the model's own (pi-ai
// stream.ts and pi-catalog's gemini.kdl, omp 18.14.0). "effort" there is a
// budget too, the same -1.
func ompGeminiMode(id string) string {
	name := strings.ToLower(id[strings.LastIndex(id, "/")+1:])
	if strings.HasPrefix(name, "gemini-1") || strings.HasPrefix(name, "gemini-2") {
		return "budget"
	}
	return "google-level"
}
