package davsync

import (
	"context"
	"github.com/yetone/magpie/internal/profile"
	"os"
	"path/filepath"
	"testing"
)

// Old profile documents can travel with the backup; neither computer's
// manually chosen model is part of the synchronization lifecycle.
func TestSyncPreservesManualAgentSettings(t *testing.T) {
	f, srv := newFakeS3(t)
	cfg := f.config(srv)
	a, b := newComputer(t), newComputer(t)
	write := func(c computer, model string) string {
		t.Helper()
		c.use(t)
		p := filepath.Join(string(c), ".claude", "settings.json")
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(model), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	write(a, `{"model":"a-manual"}`)
	if err := profile.Save("legacy", profile.Profile{Fields: map[string]string{"claude.model": "old-model"}}); err != nil {
		t.Fatal(err)
	}
	sync := func() {
		t.Helper()
		if err := Configure(cfg); err != nil {
			t.Fatal(err)
		}
		if err := Now(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	sync()
	p := write(b, `{"model":"b-manual"}`)
	sync()
	got, err := os.ReadFile(p)
	if err != nil || string(got) != `{"model":"b-manual"}` {
		t.Fatalf("model changed: %q %v", got, err)
	}
	ps, err := profile.Load()
	if err != nil || ps["legacy"].Fields["claude.model"] != "old-model" {
		t.Fatalf("lost legacy document: %v %v", ps, err)
	}
}
