package cli

import (
	"io/fs"
	"os"
	"path/filepath"

	"sort"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/provider"
)

// cliHome is groupsHome with every other folder an agent keeps files in
// (XDG_DATA_HOME, XDG_STATE_HOME, Windows' APPDATA) inside its home, so
// that homeFiles sees whatever a command writes, and with the gateway's
// address one nothing answers at. No agent runs: what a command prints
// doesn't gain "restart the Codex app" where Codex is open (and on Windows,
// where anything may be).
func cliHome(t *testing.T) {
	t.Helper()
	groupsHome(t)
	was := agent.Running
	agent.Running = func(...string) bool { return false }
	t.Cleanup(func() { agent.Running = was })
	home := os.Getenv("HOME")
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("MAGPIE_ADDR", "127.0.0.1:1")
}

// homeFiles is every folder and file in the test's home, config and cache
// folders, with each file's contents. Not magpie's cli-identity.json: who
// Cursor's and Devin's CLIs say is signed in, asked again behind a look a
// minute after the last answer and kept when it comes, which may be during
// a later command, or in a later test's home.
func homeFiles(t *testing.T) map[string]string {
	t.Helper()
	identities := filepath.Join(filepath.Dir(provider.Path()), "cli-identity.json")
	m := map[string]string{}
	for _, root := range []string{os.Getenv("HOME"), os.Getenv("XDG_CONFIG_HOME"), os.Getenv("XDG_CACHE_HOME")} {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if p == identities {
				return nil
			}
			if d.IsDir() {
				m[p] = "(folder)"
				return nil
			}
			b, err := os.ReadFile(p)
			m[p] = string(b)
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return m
}

// filesChanged names what differs between two homeFiles, "" for nothing.
func filesChanged(before, after map[string]string) string {
	var out []string
	for p, b := range before {
		if a, ok := after[p]; !ok {
			out = append(out, "removed "+p)
		} else if a != b {
			out = append(out, "changed "+p+":\n"+a)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			out = append(out, "made "+p)
		}
	}
	sort.Strings(out)
	return strings.Join(out, "\n")
}

func writeHomeFile(t *testing.T, rel, s string) string {
	t.Helper()
	p := filepath.Join(os.Getenv("HOME"), rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// codexOnAdvanced is the reporter's machine: Codex's own config.toml, a
// routing group advanced, and Codex put on it with magpie codex
// group/advanced. It gives the config's path.
func codexOnAdvanced(t *testing.T) string {
	t.Helper()
	cliHome(t)
	path := writeHomeFile(t, filepath.Join(".codex", "config.toml"),
		"model = \"gpt-5.5\"\nmodel_reasoning_effort = \"high\"\n\n[projects.\"/Users/me/work\"]\ntrust_level = \"trusted\"\n")
	for _, args := range [][]string{{"group", "add", "advanced", "models=a/m,b/gpt-5.5"}, {"codex", "group/advanced"}} {
		if _, err := printed(t, func() error { return run(args) }); err != nil {
			t.Fatalf("magpie %s: %v", strings.Join(args, " "), err)
		}
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), `"group/advanced"`) || !strings.Contains(string(b), "trust_level") {
		t.Fatalf("Codex isn't on group/advanced:\n%s", b)
	}
	return path
}
