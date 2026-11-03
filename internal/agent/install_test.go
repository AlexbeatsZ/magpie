package agent

import (
	"os/exec"
	"reflect"
	"testing"
)

// The install commands are the vendors' own (#727): the installer their
// docs give first, then npm's package, the one the update uses; per OS.
func TestInstallCommands(t *testing.T) {
	cmds := func(id, goos string) []string {
		var out []string
		for _, c := range installCommands(id, goos, true) {
			out = append(out, c.Via+": "+c.Command)
		}
		return out
	}
	for _, c := range []struct {
		id, goos string
		want     []string
	}{
		{"claude", "darwin", []string{"script: curl -fsSL https://claude.ai/install.sh | bash", "npm: npm install -g @anthropic-ai/claude-code"}},
		{"claude", "windows", []string{"powershell: irm https://claude.ai/install.ps1 | iex", "npm: npm install -g @anthropic-ai/claude-code"}},
		{"codex", "darwin", []string{"script: curl -fsSL https://chatgpt.com/codex/install.sh | sh", "brew: brew install --cask codex", "npm: npm install -g @openai/codex"}},
		{"codex", "linux", []string{"script: curl -fsSL https://chatgpt.com/codex/install.sh | sh", "npm: npm install -g @openai/codex"}},
		{"gemini", "linux", []string{"brew: brew install gemini-cli", "npm: npm install -g @google/gemini-cli"}},
		{"gemini", "windows", []string{"npm: npm install -g @google/gemini-cli"}},
		{"opencode", "darwin", []string{"script: curl -fsSL https://opencode.ai/install | bash", "npm: npm install -g opencode-ai"}},
		{"opencode", "windows", []string{"npm: npm install -g opencode-ai"}},
		{"deveco", "darwin", []string{"npm: npm install -g @deveco/deveco-code"}},
		{"deveco", "windows", []string{"npm: npm install -g @deveco/deveco-code"}},
		{"deveco", "linux", nil}, // no Linux build in its package
		{"pi", "linux", []string{"npm: npm install -g @earendil-works/pi-coding-agent"}},
		{"cursor", "darwin", nil}, // no command magpie knows
	} {
		if got := cmds(c.id, c.goos); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s on %s: %q, want %q", c.id, c.goos, got, c.want)
		}
	}
}

// The command without Node is one line every shell it's run in reads:
// bash and zsh on a Mac or Linux.
func TestInstallWithoutNodeParses(t *testing.T) {
	c := npmInstall("@earendil-works/pi-coding-agent", "linux", false).Command
	for _, sh := range []string{"bash", "zsh", "sh"} {
		p, err := exec.LookPath(sh)
		if err != nil {
			continue
		}
		if out, err := exec.Command(p, "-n", "-c", c).CombinedOutput(); err != nil {
			t.Errorf("%s -n: %v %s", sh, err, out)
		}
	}
}
