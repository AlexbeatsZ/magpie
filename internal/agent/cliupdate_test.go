package agent

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func TestNewerVersions(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"2.1.3", "2.1.2", true},
		{"2.1.2", "2.1.3", false},
		{"2.1.3", "2.1.3", false},
		{"0.10.0", "0.9.9", true},
		{"1.0", "1.0.0", false},
		{"1.0.1", "1.0", true},
		{"v1.2.0", "1.1.9", true},
		{"1.0.89", "1.0.87-0", true},
		{"1.0.87", "1.0.87-0", true},
		{"1.0.87-0", "1.0.87", false},
		{"1.0.87-beta.10", "1.0.87-beta.9", true},
		{"1.0.87-beta.2", "1.0.87-alpha.9", true},
		{"0.74.0", "0.3.0", true},
	} {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestParseCLIVersion(t *testing.T) {
	for out, want := range map[string]string{
		"2.1.284 (Claude Code)\n": "2.1.284",
		"codex-cli 0.155.1":       "0.155.1",
		"0.61.0":                  "0.61.0",
		"GitHub Copilot CLI 1.0.87-0.\nRun 'copilot update' to check for updates.": "1.0.87-0",
		"crush version v0.3.0":  "0.3.0",
		"omp/16.3.5":            "16.3.5",
		"\x1b[1m1.18.32\x1b[0m": "1.18.32",
		"command not found":     "",
		"":                      "",
	} {
		if got := parseVersion(out); got != want {
			t.Errorf("parseVersion(%q) = %q, want %q", out, got, want)
		}
	}
}

// file makes a file (and its folders) in a temp tree, executable.
func file(t *testing.T, path, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(body, "#!") {
		testenv.Program(t, path, body)
	} else if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// An agent's CLI installed with bun is updated by bun add -g, which can't
// use a SOCKS5 proxy (#1409): it is given the bridge in front of it. npm,
// which can, is given the proxy as it is.
func TestUpdateWithBunBridgesSOCKS(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	const socks = "socks5://127.0.0.1:4401"
	names := []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy", "ALL_PROXY", "all_proxy"}
	for _, k := range names {
		t.Setenv(k, socks)
	}
	// the proxy URLs only: NO_PROXY's hosts or Node's NODE_USE_ENV_PROXY=1,
	// set on the machine running the test, name no proxy
	proxies := func(env []string) []string {
		var out []string
		for _, kv := range env {
			if k, v, _ := strings.Cut(kv, "="); slices.Contains(names, k) && v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	bun := proxies(updateEnv(&updater{via: "bun", cmd: []string{"bun", "add", "-g", "x@latest"}}))
	if len(bun) == 0 {
		t.Fatal("bun is given no proxy")
	}
	for _, v := range bun {
		if !strings.HasPrefix(v, "http://127.0.0.1:") {
			t.Fatalf("bun is given %s", v)
		}
	}
	for _, v := range proxies(updateEnv(&updater{via: "npm", cmd: []string{"npm", "i", "-g", "x@latest"}})) {
		if v != socks {
			t.Fatalf("npm is given %s", v)
		}
	}
}
