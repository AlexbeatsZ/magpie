package agent

// Kimi Code, Moonshot's kimi CLI, keeps its settings in config.toml — the
// new one's (TypeScript, 2.x) in ~/.kimi-code ($KIMI_CODE_HOME), the old
// Python kimi-cli's in ~/.kimi ($KIMI_SHARE_DIR), in the same shape (see
// KimiDir): the model sessions start with as default_model, a key
// of its [models."<key>"] tables, each naming a [providers.<name>] table and
// the model to ask it for. magpie adds itself as the provider "magpie" (the
// gateway, spoken to as Kimi's own chat completions, so the thinking goes
// both ways as reasoning_content) and one model table per catalog model,
// keyed "magpie/<provider>/<model>", so the catalog joins Kimi's /model
// picker. The default the user had is stashed and put back when magpie steps
// out. Kimi refuses a config whose default_model or a model's provider is
// missing, so the provider goes in before the models and the default after,
// and out the other way round.

import (
	"path/filepath"
)

// KimiDir is the folder Kimi Code keeps its config.toml in, and whether it is
// the old kimi-cli's. kimi-cli (1.50 on) hands over to the new Kimi Code,
// which reads only ~/.kimi-code: it offers to copy ~/.kimi's config over
// when it first starts, and once that's done never reads ~/.kimi again, so a
// model written there is one it never sees (#290). ~/.kimi is kimi-cli's only where
// the new one isn't: no ~/.kimi-code, nor $KIMI_CODE_HOME.
func KimiDir(home string) (dir string, legacy bool) { return kimiDir(here(home)) }

// kimiDir is KimiDir at a place: in a WSL distro its variables aren't
// read, and a stopped one's is ~/.kimi-code until it is looked at.
func kimiDir(at place) (dir string, legacy bool) {
	if d := at.getenv("KIMI_CODE_HOME"); d != "" {
		return d, false
	}
	code := at.join(".kimi-code")
	if at.isDir(code) {
		return code, false
	}
	if d := at.getenv("KIMI_SHARE_DIR"); d != "" {
		return d, true
	}
	if d := at.join(".kimi"); at.isDir(d) {
		return d, true
	}
	return code, false
}

func kimi(home string) *Agent { return kimiIn(here(home)) }

// kimiIn is Kimi Code at a place: this machine's home, or a WSL distro's
// (see wsl.go), its provider naming the gateway as it reaches it from there.
func kimiIn(at place) *Agent {
	dir, _ := kimiDir(at)
	path := filepath.Join(dir, "config.toml")

	return &Agent{
		ID: "kimi", Name: "Kimi Code", Icon: "kimi", Aliases: []string{"kimi-code", "kimi-cli"},
		// kimi-cli (Python) says KimiCLI/…; Kimi Code 2.x, kimi-code-cli/…
		// (jiabubu on Discord: its requests showed the raw name)
		UA:  []string{"kimicli", "kimi-code-cli"},
		Bin: "kimi", Dir: dir, Path: path,
	}
}
