package agent

import (
	"testing"
)

// TestClaudeDatedAlias (#496): models.dev lists an alias and its dated id
// as two models (claude-opus-4-5, claude-opus-4-5-20251101), and the
// picker showed both. The dated one now names its alias, for the picker to
// show one row; through magpie too, the [1m] mark and all. Two dated ids
// of one name, or a dated one with no alias beside it, are left alone.
func TestClaudeDatedAlias(t *testing.T) {
	opts := claudeDated([]Option{
		{Value: "claude-opus-4-5", Group: "Claude Code"},
		{Value: "claude-opus-4-5-20251101", Group: "Claude Code"},
		{Value: "claude/claude-sonnet-4-5[1m]", Group: "Claude Code"},
		{Value: "claude/claude-sonnet-4-5-20250929[1m]", Group: "Claude Code"},
		{Value: "claude-3-5-sonnet-20240620", Group: "Claude Code"},
		{Value: "claude-3-7-sonnet", Group: "Other"},
		{Value: "claude-3-7-sonnet-20250219", Group: "Other"},
		{Value: "claude-3-7-sonnet-20250301", Group: "Other"},
		{Value: "claude-haiku-4-5-20251001", Group: "Other"},
		{Value: "claude-haiku-4-5", Group: "Elsewhere"},
	})
	want := map[string]string{
		"claude-opus-4-5-20251101":              "claude-opus-4-5",
		"claude/claude-sonnet-4-5-20250929[1m]": "claude/claude-sonnet-4-5[1m]",
	}
	for _, o := range opts {
		if o.Alias != want[o.Value] {
			t.Errorf("%s: alias %q, want %q", o.Value, o.Alias, want[o.Value])
		}
	}
}
