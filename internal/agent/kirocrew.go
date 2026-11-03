package agent

// Kiro Crew (kirodotdev/KiroCrew, v0.9.0: the KiroCrew.app desktop app, the
// kirocrew CLI, its web dashboard and the Gateway they share; Vincent on
// Discord) is a persistent agent that runs each session on a harness: an
// ACP agent it starts and talks to. It has no model provider of its own
// to point at magpie; its models are its harness's. So what magpie sets is
// which harness it runs: one already connected to magpie, whose own config
// then sends its requests to the gateway.
//
// Its settings are one JSON file in its home, $KIROCREW_HOME (made
// absolute; a root, /usr, /System or /etc is refused and the default used)
// or ~/.kiro/crew, the desktop app's too (config/paths.py config_dir,
// _valid_override_home). config.local.json beside it is merged over it
// key by key and wins (config/loader.py). The keys (config/sections.py):
//
//	{"agent": {"acp_backend": "codex",          // "" is kiro-cli, the default
//	           "member_acp_backend": "codex"}}  // crew members', default "kas"
//
// acp_backend is one of "", kas, claude, codex, opencode, pi, goose and
// deepseek; member_acp_backend one that can dispatch (claude, kas, codex,
// opencode, goose: ACP_BACKENDS_MEMBER_DISPATCH in agent_sdk/backends.py).
// The Gateway reloads the files when they change; a session already open
// keeps its harness.
//
// The harnesses magpie offers are those it wires itself, through the files
// the harness reads under Crew too: Codex (codex-acp reads ~/.codex, or
// CODEX_HOME), Claude Code (claude-agent-acp runs the user's claude, which
// reads ~/.claude), OpenCode (opencode acp), Pi (pi-acp runs pi) and Goose
// (goose acp). Not deepseek: Crew runs it as dsh --profile acp, a profile
// of its own that magpie's dsh wiring doesn't reach. Codex, Claude Code
// and Pi are reached through an npm adapter Crew doesn't ship
// (agent_sdk/backend_install.py); magpie looks for it where Crew does and
// says how to install it when it isn't there, and installs nothing.
//
// Connecting keeps the harness Kiro Crew is set to when it is one of
// these, and connects that agent to magpie too when it isn't yet (Crew's
// first run picks one, Claude Code on a Windows without kiro-cli: wuping9525
// on Discord had to connect it by hand first). Disconnect takes that agent
// off magpie again only when this connect is what put it on, and the user
// hasn't disconnected it since. On a harness magpie doesn't wire (KAS,
// kiro-cli, deepseek) it picks the first harness connected to magpie
// (Codex, Claude Code, OpenCode, Pi, Goose), one whose adapter is here
// first. It keeps what the two keys said before, which Disconnect puts
// back. A harness set in config.local.json is the user's: magpie writes
// none over it.

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
)

// kiroCrewOS is the system Kiro Crew's files are looked for as; a var so
// a test on a Mac can say Windows.
var kiroCrewOS = runtime.GOOS

// kiroCrewApp is Kiro Crew's macOS bundle.
var kiroCrewApp = desktopApp{id: "com.amazon.kiro.crew", names: []string{"KiroCrew.app", "Kiro Crew.app"}}

// kiroCrewHome is Kiro Crew's home, as config_dir finds it.
func kiroCrewHome(home string, getenv func(string) string) string {
	if o := getenv("KIROCREW_HOME"); o != "" {
		if strings.HasPrefix(o, "~") && (len(o) == 1 || o[1] == '/' || o[1] == filepath.Separator) {
			o = filepath.Join(home, o[1:])
		}
		if p, err := filepath.Abs(o); err == nil {
			if r, err := filepath.EvalSymlinks(p); err == nil {
				p = r
			}
			if !kiroCrewUnsafe(p) {
				return p
			}
		}
	}
	return filepath.Join(home, ".kiro", "crew")
}

// kiroCrewUnsafe is a home Kiro Crew refuses (_is_unsafe_home): a root, or
// a folder in /usr, /System or /etc (/private/etc, as a Mac resolves it).
func kiroCrewUnsafe(p string) bool {
	p = filepath.Clean(p)
	if filepath.Dir(p) == p {
		return true
	}
	if kiroCrewOS == "windows" {
		return false
	}
	for _, d := range []string{"/usr", "/System", "/etc", "/private/etc"} {
		if p == d || strings.HasPrefix(p, d+"/") {
			return true
		}
	}
	return false
}

func kiroCrew(home string) *Agent {
	dir := kiroCrewHome(home, appdir.Getenv)
	path := filepath.Join(dir, "config.json")

	return &Agent{
		ID: "kirocrew", Name: "Kiro Crew", Icon: "generic", Aliases: []string{"kiro-crew", "crew"},
		Dir: dir, Path: path,
		// its app, its command, or its home with its settings in it; never
		// ~/.kiro alone, which is Kiro's and kiro-cli's
		detect: func() bool {
			return isFile(path) || macBundle(kiroCrewApp) || binAt("kirocrew") != "" ||
				isFile(filepath.Join(home, ".kirocrew-app", ".venv", "bin", "kirocrew")) ||
				kiroCrewOS == "windows" && isFile(filepath.Join(home, ".kirocrew-app", ".venv", "Scripts", "kirocrew.exe"))
		},
	}
}
