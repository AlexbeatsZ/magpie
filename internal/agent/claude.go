package agent

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yetone/magpie/internal/provider"
)

// claudeNameRe finds the Claude model an id names, as Claude Code matches it
// to its settings: claude-opus-5-5 in claude-opus-5-5[1m], in a dated id or
// under a provider's prefix (anthropic/…).
var claudeNameRe = regexp.MustCompile(`claude-(opus|sonnet|haiku|fable)-(\d+)(?:-(\d{1,2}))?(?:[^0-9]|$)`)

// claudeName is the Claude model a model id names, "" for another vendor's,
// an alias (opus) or none. A version written with a dot (claude-opus-4.6)
// is the one Claude Code is given with dashes (provider.ClaudeSpelled).
func claudeName(model string) string {
	m := claudeNameRe.FindStringSubmatch(strings.ToLower(provider.ClaudeDashed(model)))
	if m == nil {
		return ""
	}
	name := "claude-" + m[1] + "-" + m[2]
	if m[3] != "" {
		name += "-" + m[3]
	}
	return name
}

// claudeCompactEnv is where Claude Code compacts a conversation, the smaller
// of it and the model's window, a [1m] one's too. magpie sets it to
// settings.WorkingWindow, so a 1M model isn't run to 1M with every turn
// sending all of it (X: Chen, turns of ~550K tokens waiting 70–90s for a
// first token); settings.FullContext leaves it out, and so does a Claude
// model, which Anthropic runs to its whole window.
const claudeCompactEnv = "CLAUDE_CODE_AUTO_COMPACT_WINDOW"

func claude(home string) *Agent { return claudeIn(here(home)) }

// claudeIn is Claude Code as it lives at a place: this machine's home, or
// a WSL distro's (see wsl.go), its settings.json naming the gateway as it
// reaches it from there.
func claudeIn(at place) *Agent {
	path := filepath.Join(at.claudeDir(), "settings.json")

	var self *Agent
	self = &Agent{
		ID: "claude", Name: "Claude Code", Icon: "claudecode-color", Aliases: []string{"cc", "claude-code"},
		UA:  []string{"claude-cli", "claude-code"},
		Bin: "claude", Dir: filepath.Dir(path), Path: path,
	}
	return self
}

// claudeDated marks each dated Claude id that another option of its group
// names undated (models.dev lists claude-opus-4-5 and
// claude-opus-4-5-20251101, the one model) as that one's Alias. Two dated
// ids of one name are two models, and left as they are.
func claudeDated(opts []Option) []Option {
	const mark = "[1m]"
	bare := func(v string) string { return strings.TrimSuffix(v, mark) }
	type key struct{ group, value string }
	alias := map[key]int{}
	for i, o := range opts {
		if !dated.MatchString(bare(o.Value)) {
			alias[key{o.Group, bare(o.Value)}] = i
		}
	}
	twins := map[key][]int{}
	for i, o := range opts {
		v := bare(o.Value)
		if base := dated.ReplaceAllString(v, ""); base != v && claudeName(base) != "" {
			if _, ok := alias[key{o.Group, base}]; ok {
				twins[key{o.Group, base}] = append(twins[key{o.Group, base}], i)
			}
		}
	}
	for k, is := range twins {
		if len(is) == 1 {
			opts[is[0]].Alias = opts[alias[k]].Value
		}
	}
	return opts
}
