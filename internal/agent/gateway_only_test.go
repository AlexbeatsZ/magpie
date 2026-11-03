package agent

import (
	"errors"
	"testing"
)

func TestResourceRegistryBuildsReadOnlyDescriptors(t *testing.T) {
	home, cfg := homes()
	for _, a := range onMachine(home, cfg) {
		if len(a.Fields) != 0 || a.Native != nil || a.Sync != nil || a.Follow != nil || a.Routed != nil || a.Unwire != nil {
			t.Fatalf("registry constructed model integration hooks for %s", a.ID)
		}
	}
}

func TestGatewayOnlyHasNoConfigWriters(t *testing.T) {
	if len(All()) != 0 || len(Detected()) != 0 {
		t.Fatal("gateway exposes another program's configuration")
	}
	for _, a := range Inventory() {
		if len(a.Fields) != 0 || a.Native != nil || a.Sync != nil || a.Join != nil || a.Follow != nil || a.Unwire != nil || a.Launch != nil || a.Import != nil {
			t.Fatalf("resource inventory exposes a model integration: %s", a.ID)
		}
	}
	if len(Inventory()) == 0 {
		t.Fatal("resource inventory is empty")
	}
	found := map[string]bool{}
	for _, a := range Clients() {
		found[a.ID] = true
		if len(a.Fields) != 0 || a.Native != nil || a.Sync != nil || a.Join != nil || a.Launch != nil {
			t.Fatalf("%s exposes a config writer through request attribution", a.ID)
		}
	}
	for _, id := range []string{"claude", "codex", "opencode"} {
		if !found[id] {
			t.Errorf("request attribution lost %s", id)
		}
	}
}

func TestGatewayOnlyRejectsOldConfigOperations(t *testing.T) {
	writes := 0
	a := &Agent{ID: "old-profile", Fields: []Field{{
		Key: "model", Get: func() string { return "own" },
		Set: func(string) error { writes++; return nil },
	}}}
	for _, op := range []func() error{
		func() error { return a.Apply("model", "relay/test") },
		func() error { return a.Pick("model", "relay/test") },
		a.Connect, a.Reapply, a.Disconnect, a.DisconnectOffline,
		func() error { return TuneApply("", "codex", TuneCompact, "", "1000") },
		func() error { return TuneUndo("", "codex", TuneCompact, "") },
	} {
		if err := op(); !errors.Is(err, ErrWiringRemoved) {
			t.Errorf("old operation returned %v", err)
		}
	}
	SyncCatalog()
	SyncCatalogSoon()
	SyncCatalogSettled()
	if writes != 0 {
		t.Fatalf("removed operations wrote config %d times", writes)
	}
}
