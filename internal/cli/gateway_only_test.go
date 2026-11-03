package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/agent"
)

func TestGatewayOnlyCLIRetiresConfigCommands(t *testing.T) {
	cliHome(t)
	cfg := filepath.Join(os.Getenv("USERPROFILE"), ".codex", "config.toml")
	const original = "# user setting\nmodel = \"own-model\"\n"
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	// Exercise the real startup as well as the dispatch. These commands
	// used to migrate and rewrite an installed client's config.
	for _, args := range [][]string{
		{"codex", "a/m"}, {"claude", "a/m"}, {"agy", "env"},
		{"agy", "run", "--version"}, {"agents"}, {"ls"},
		{"use", "old-profile"},
		{"visible", "codex", "all"},
	} {
		if err := run(args); err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%v: %v", args, err)
		}
	}
	agent.SyncCatalogSettled()
	got, err := os.ReadFile(cfg)
	if err != nil || string(got) != original {
		t.Fatalf("client config changed: %q, %v", got, err)
	}
	for _, removed := range []string{"magpie <agent>", "magpie use ", "magpie agents "} {
		if strings.Contains(usage, removed) {
			t.Errorf("help still advertises %s", removed)
		}
	}
}
