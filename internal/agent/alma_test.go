package agent

import (
	"sync"
	"testing"
)

// fakeAlma is Alma's API as its spec has it: providers and settings in
// memory, and every write counted.
type fakeAlma struct {
	mu        sync.Mutex
	providers []map[string]any
	settings  map[string]any
	writes    []string
	keys      []string // the API keys sent
	n         int
}

func (f *fakeAlma) provider(id string) map[string]any {
	for _, p := range f.providers {
		if p["id"] == id {
			return p
		}
	}
	return nil
}

// Under test no Alma is reached unless a fake one is set up.
func TestAlmaNoneUnderTest(t *testing.T) {
	if almaAPI != "" {
		t.Fatalf("almaAPI under test: %q", almaAPI)
	}
}
