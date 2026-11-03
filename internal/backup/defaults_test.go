package backup

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Legacy model settings survive encryption, but restoring them never writes
// a manually configured agent. New backups do not collect those settings.
func TestLegacyAgentSettingsArePreservedAndSkipped(t *testing.T) {
	home(t)
	path := filepath.Join(os.Getenv("HOME"), ".claude", "settings.json")
	const own = `{"model":"own-model","theme":"dark"}`
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(own), 0600); err != nil {
		t.Fatal(err)
	}
	current, err := Collect(false, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(current.Agents) != 0 {
		t.Fatalf("captured manual models: %v", current.Agents)
	}
	legacy := Bundle{Version: 1, Agents: map[string]string{"claude.model": "old-model", "claude.effort": "low"}}
	sealed, err := Seal(legacy, "pw")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := Open(sealed, "pw")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.Agents, legacy.Agents) {
		t.Fatal("lost legacy fields")
	}
	result, err := Restore(restored, Parts{Agents: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Agents != 0 || len(result.Skipped) != 2 {
		t.Fatalf("unexpected restore: %+v", result)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != own {
		t.Fatalf("manual settings changed: %q %v", got, err)
	}
}
