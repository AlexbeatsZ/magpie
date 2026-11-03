package tui

import "testing"

func TestGatewayOnlyTerminalPages(t *testing.T) {
	m := newModel()
	if m.page != pageProviders {
		t.Fatalf("terminal starts on %v", m.page)
	}
	for _, key := range []string{"1", "2", "3", "4", "5", "[", "]"} {
		p, ok := m.pageKey(key)
		if !ok || p == pageAgents {
			t.Fatalf("%q exposes removed page %v", key, p)
		}
	}
	for _, p := range []page{pageAgents} {
		m.goTo(p)
		if m.page != pageProviders {
			t.Fatal("old navigation exposes configuration page")
		}
	}
}
