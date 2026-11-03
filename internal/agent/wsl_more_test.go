package agent

import (
	"strings"
	"testing"
)

// wslMore is each CLI #515 brings into WSL distros: the file under the
// distro's home that carries magpie's provider once one of magpie's models
// is picked, and what of the gateway's address it names (its /v1, or the
// bare address for one spoken to as Anthropic messages).
var wslMore = []struct {
	id, file, suffix string
}{
	{"opencode", ".config/opencode/opencode.json", "/v1"},
	{"mimocode", ".config/mimocode/mimocode.json", "/v1"},
	{"kimi", ".kimi-code/config.toml", "/v1"},
	{"omp", ".omp/agent/models.yml", "/v1"},
	{"crush", ".config/crush/crush.json", "/v1"},
	{"hermes", ".hermes/config.yaml", "/v1"},
	{"morph", ".morph/config.yaml", "/v1"},
	{"grok", ".grok/config.toml", "/v1"},
	{"droid", ".factory/settings.json", "/v1"},
	{"fx", ".fx/settings.json", "/v1"},
	{"commandcode", ".commandcode/providers.json", "/v1"},
	{"minimax-code", ".minimax/config.yaml", ""},
	{"dsh", ".dsh/profiles/web/cordis.patch.yml", "/v1"}, // liby233: no dsh reads config.yaml
	{"muse", ".config/muse/settings.json", "/v1"},
	{"qoder", ".qoder/settings.json", "/v1"},
	{"qoder-cn", ".qoder-cn/settings.json", "/v1"},
	// lgtm on Discord: a CodeBuddy Code in WSL wasn't found, nor were the
	// other CLIs magpie set up only on this machine
	{"codebuddy", ".codebuddy/models.json", "/v1/chat/completions"},
	{"cline", ".cline/data/settings/providers.json", "/v1"},
	{"atomcode", ".atomcode/config.toml", "/v1"},
	{"snow", ".snow/profiles/magpie.json", "/v1"},
	{"goose", ".config/goose/custom_providers/magpie.json", "/v1"},
	{"mothx", ".mothx/settings.json", "/v1"},
	{"penguin", ".penguin/data/default_project/.project_config.toml", "/v1"},
	{"reasonix", ".reasonix/config.toml", "/v1"},
}

// wslFinds is what a probe of a distro with the agent of kind k prints,
// and the distro it makes: its folder, or for one found by its command
// alone, the command and the version it gives.
func wslFinds(k wslKind) (probe string, has map[string]bool, versions map[string]string) {
	if k.dir != "" {
		return "dir:" + k.dir + "\n", map[string]bool{"dir:" + k.dir: true}, nil
	}
	probe = "bin:" + k.bin + " /home/me/.local/bin/" + k.bin + "\n"
	has = map[string]bool{"bin:" + k.bin: true}
	if k.version {
		probe += "ver:" + k.id + " " + k.bin + " v2.1.0\n"
		versions = map[string]string{k.id: "2.1.0"}
	}
	return probe, has, versions
}

// The probe asks after each of them by its folder and, but for grok and
// fx, whose names other tools have too, its command; and finds each by
// either.
func TestWSLProbeFindsMoreAgents(t *testing.T) {
	for _, c := range wslMore {
		k := wslKindOf(c.id)
		if k.dir != "" && !strings.Contains(wslProbeScript, `[ -d "$HOME/`+k.dir+`" ] && echo dir:`+k.dir+`; `) {
			t.Errorf("%s: probe lacks its folder %s", c.id, k.dir)
		}
		if k.dir != "" && !strings.HasPrefix(c.file, k.dir+"/") {
			t.Errorf("%s: its file %s isn't under %s", c.id, c.file, k.dir)
		}
		if k.bin != "" && !strings.Contains(wslProbeScript, `command -v `+k.bin+` `) {
			t.Errorf("%s: probe lacks its command %s", c.id, k.bin)
		}
		if p, _, _ := wslFinds(k); k.dir != "" || k.version {
			if d := parseProbe("U", "home:/home/me\n"+p); d == nil || !k.found(*d) {
				t.Errorf("%s: not found by %q", c.id, p)
			}
		}
		if k.bin != "" && !k.version {
			if d := parseProbe("U", "home:/home/me\nbin:"+k.bin+" /usr/bin/"+k.bin+"\n"); d == nil || !k.found(*d) {
				t.Errorf("%s: not found by its command", c.id)
			}
		}
	}
	for _, bin := range []string{"fx", "grok", "goose"} {
		if strings.Contains(wslProbeScript, `command -v `+bin+` `) {
			t.Errorf("probe takes any %s command for the agent", bin)
		}
	}
	// Kimi Code's old kimi-cli, in ~/.kimi, is found by its command
	if d := parseProbe("U", "home:/home/me\nbin:kimi /home/me/.local/bin/kimi\n"); d == nil || !wslKindOf("kimi").found(*d) {
		t.Error("kimi by its command")
	}
}
