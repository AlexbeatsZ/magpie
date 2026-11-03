package agent

// T3 Code (pingdotgg/t3code, a GUI that drives Claude Code, Codex, Cursor
// and others; KevinXC on Discord) keeps its settings in
// ~/.t3/userdata/settings.json ($T3CODE_HOME/userdata), a sparse JSON file
// its server watches and reloads on an outside edit. Each provider it lists
// is an instance of a driver, and besides the built-in ones (keyed by the
// driver, "claudeAgent", "codex", …) it takes instances of one's own:
//
//	{"providerInstances":{"<id>":{"driver":"claudeAgent","displayName":…,
//	  "enabled":true,"environment":[{"name":…,"value":…,"sensitive":false}],
//	  "config":{"customModels":[{"slug":…,"name":…,"capabilities":
//	    {"optionDescriptors":[{"id":"effort","type":"select",…}]}}]}}}}
//
// A claudeAgent instance runs Claude Code (Anthropic's Agent SDK, the
// user's ~/.claude settings read as Claude Code reads them) with the
// instance's environment added, and lists Claude's models and its custom
// models, each sent as Claude Code's model. So magpie is an instance of its
// own, "magpie": Claude Code pointed at the gateway (ANTHROPIC_BASE_URL and
// the token Claude Code is routed with, so its requests are Claude Code's
// to the gateway, tiers and all), every magpie model one of its custom
// models, a 1M one marked [1m] as magpie marks it for Claude Code, with
// a Reasoning pick of the levels it takes and a Thinking switch
// (t3Capabilities). It is
// magpie's alone: the user's own instances, their custom models and T3's
// other keys stay as they are, and off takes only it out. A binary path or
// Claude home the user gave T3's own Claude is carried over, so it runs the
// same Claude Code.
//
// T3 Code runs each thread on the harness of the instance its model is
// picked from, so with magpie on Claude Code alone every magpie model ran
// there (karissasel_ on X: "can't pick the harness, it's Claude Code").
// So magpie has a Codex instance too, "magpie-codex" (magpie · Codex): a
// codex instance whose launchArgs, which T3 passes to `codex app-server`
// and `codex exec` (codexLaunchArgs.ts), are Codex's own -c overrides
// making magpie its model provider and magpie's model catalog for T3
// (magpie-codex-models.json beside settings.json) its catalog, so T3's
// picker lists every magpie model under it from Codex's model/list. The
// user's ~/.codex config.toml isn't written: the overrides are this
// instance's alone, and T3's own Codex stays as it is. It is written with
// magpie's instance; one the user removed in T3 isn't brought back by a
// sync (t3CodexOffered), only by turning magpie on again. Its name, whether
// it is on, and the keys magpie doesn't write are the user's; the launch
// arguments and the catalog are magpie's.
//
// The settings environment (~/.claude/settings.json's env) is applied over
// the process', so a Claude Code routed elsewhere in its own settings
// takes its requests there; one routed through magpie or not routed at all
// asks the gateway. A provider switch on there (CLAUDE_CODE_USE_BEDROCK, …)
// takes it to that cloud, which Check says.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
)

func t3code(home string) *Agent {
	base := appdir.Getenv("T3CODE_HOME")
	if base == "" {
		base = filepath.Join(home, ".t3")
	} else if rest, ok := strings.CutPrefix(base, "~"); ok {
		base = filepath.Join(home, rest)
	}
	path := filepath.Join(base, "userdata", "settings.json")

	return &Agent{
		ID: "t3code", Name: "T3 Code", Icon: "t3code", Aliases: []string{"t3", "t3-code"},
		Dir: base, Path: path,
		// ~/.t3 is made at its first start; a Mac app never opened yet is
		// found by its bundle. No command: `t3` is a common name.
		detect: func() bool { return isDir(base) || t3App(home) },
	}
}

// t3App reports whether T3 Code's app is in a Mac's Applications, by the
// names its builds take.
func t3App(home string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	for _, d := range macAppFolders(home) {
		for _, n := range []string{"T3 Code (Alpha).app", "T3 Code.app"} {
			if _, err := os.Stat(filepath.Join(d, n)); err == nil {
				return true
			}
		}
	}
	return false
}
