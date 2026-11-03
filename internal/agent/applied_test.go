package agent

import (
	"testing"
)

// sameGroup compares two values of a field, one of them read back from the
// agent's own settings where Claude Code's [1m] mark rides on the group's
// id: the marked one is the group magpie set, not another group — the
// gateway serves either (#750's comparison, with the mark on one side).
func TestSameGroupTakesTheOneMMarkOff(t *testing.T) {
	for _, pair := range [][2]string{
		{"group/auto-gpt-5", "group/auto-gpt-5[1m]"},
		{"group/auto-gpt-5[1m]", "group/auto-gpt-5"},
		{"group/auto-gpt-5[1m]", "magpie/group/auto-gpt-5"},
	} {
		if !sameGroup(pair[0], pair[1]) {
			t.Errorf("sameGroup(%q, %q) = false", pair[0], pair[1])
		}
	}
	if sameGroup("group/auto-gpt-5", "group/auto-other") {
		t.Error("two different groups are the same")
	}
	if sameGroup("group/auto-gpt-5", "a/gpt-5") {
		t.Error("a group and a model are the same")
	}
}
