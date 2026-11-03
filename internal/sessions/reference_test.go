package sessions

import (
	"errors"
	"testing"
)

func TestHistoryReferenceDeviceAndHarness(t *testing.T) {
	for _, tc := range []struct{ os, host, want string }{
		{"windows", "META-OMEN", "omen"}, {"windows", "Meta-ROGAlly", "rog"},
		{"darwin", "MetadeMBP", "mac"}, {"linux", "debian.example", "debian"},
	} {
		if got := historyDevice(tc.os, tc.host); got != tc.want {
			t.Errorf("%+v: %q", tc, got)
		}
	}
	s := Session{Agent: "agy", ID: "75627e81-1234"}
	if got, want := Reference(s), referenceDevice()+":antigravity:"+s.ID; got != want {
		t.Fatalf("%q, want %q", got, want)
	}
	for _, other := range []Session{{Agent: "codex", ID: "abcd", Gateway: true}, {Agent: "codex", ID: "abcd", WSL: "Ubuntu"}} {
		if ref := Reference(other); ref != "" {
			t.Errorf("unknown source labeled local: %q", ref)
		}
	}
}

func TestHistoryReferenceNeverGuessesAPrefix(t *testing.T) {
	a := Session{Agent: "claude", ID: "75627e81-a"}
	b := Session{Agent: "claude", ID: "75627e81-b"}
	c := Session{Agent: "codex", ID: a.ID}
	if _, err := matchReference([]Session{a, b}, "75627e81"); !errors.Is(err, ErrReferenceAmbiguous) {
		t.Fatalf("collision: %v", err)
	}
	if _, err := matchReference([]Session{a, c}, a.ID); !errors.Is(err, ErrReferenceAmbiguous) {
		t.Fatalf("cross-agent collision: %v", err)
	}
	if got, err := matchReference([]Session{b, a}, a.ID); err != nil || got.ID != a.ID {
		t.Fatalf("exact lookup: %+v %v", got, err)
	}
	if _, err := matchReference([]Session{a}, "absent"); !errors.Is(err, ErrReferenceMissing) {
		t.Fatalf("missing: %v", err)
	}
}
