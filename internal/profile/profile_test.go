package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

// Retired profile actions must not discard older backup documents.
func TestLegacyProfilesRoundTrip(t *testing.T) {
	testenv.SetHome(t, t.TempDir())
	const document = `{"work":{"codex.model":"relay/manual","claude.model":"opus"}}`
	if err := os.MkdirAll(filepath.Dir(Path()), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(), []byte(document), 0600); err != nil {
		t.Fatal(err)
	}
	ps, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if ps["work"].Fields["codex.model"] != "relay/manual" {
		t.Fatalf("lost fields: %+v", ps)
	}
	if err := Save("copy", ps["work"]); err != nil {
		t.Fatal(err)
	}
	ps, err = Load()
	if err != nil || ps["copy"].Fields["claude.model"] != "opus" {
		t.Fatalf("lost copy: %+v %v", ps, err)
	}
	b, err := json.Marshal(ps["copy"])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]string
	if err := json.Unmarshal(b, &fields); err != nil || fields["codex.model"] != "relay/manual" {
		t.Fatalf("format changed: %s %v", b, err)
	}
	if err := Delete("copy"); err != nil {
		t.Fatal(err)
	}
}
