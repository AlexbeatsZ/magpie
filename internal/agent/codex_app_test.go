package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/yetone/magpie/internal/provider"
)

// The Restart on Codex's "the Codex app is still signed in as" notice says
// what went wrong in a notice the GUI translates, and an error it doesn't
// know (a failed listing) is "couldn't tell", never success.
func TestRestartSwitchedCodexAppSays(t *testing.T) {
	old := restartSwitchedApp
	t.Cleanup(func() { restartSwitchedApp = old })
	for _, c := range []struct {
		err  error
		want notice
	}{
		{provider.ErrCodexAppGone, noticeSwitchAppGone},
		{provider.ErrCodexAppNoEnd, noticeSwitchAppNoEnd},
		{provider.ErrCodexAppNotBack, noticeSwitchAppNotBack},
		{errors.New("ps: timed out"), noticeSwitchAppUnknown},
	} {
		restartSwitchedApp = func(context.Context) error { return c.err }
		err := RestartSwitchedCodexApp(context.Background())
		if err == nil || err.Error() != c.want.say("agent", "Codex") {
			t.Errorf("%v: said %v, want %q", c.err, err, c.want.say("agent", "Codex"))
		}
	}
	restartSwitchedApp = func(context.Context) error { return nil }
	if err := RestartSwitchedCodexApp(context.Background()); err != nil {
		t.Fatal(err)
	}
}
