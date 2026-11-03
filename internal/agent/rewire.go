package agent

import (
	"fmt"
)

// The gateway's port is set in Settings (Magic_zero on Discord): every
// agent magpie connected has the gateway's URL written into its config, so
// moving the gateway moves them too. Before the port changes, OnGateway
// notes which agents reach the gateway where it is; once it has moved,
// Rewire sets each of them again, which writes the URL it has now.

// OnGateway are the agents connected to magpie whose configs reach the
// gateway at its address now: none of their wiring is off.
func OnGateway() []*Agent {
	var out []*Agent
	for _, a := range All() {
		if len(a.Fields) == 0 || !a.Wired() {
			continue
		}
		if d := a.Drift(); d != nil && d.Kind == "unwired" {
			continue // taken off magpie before: the user's to set again
		}
		out = append(out, a)
	}
	return out
}

// Rewire puts the gateway's address now into the configs of agents, taken
// by OnGateway before it moved: each one whose wiring is off now is set
// again, an agent joined to magpie beside its own models (Codex signed in
// with ChatGPT) joined again. It says the names of those it moved, and
// what one it couldn't move failed with. Cursor Private Inference's
// variables and dsh's routes, which magpie keeps on its own, follow too;
// Claude Science's daemon, started for the old address, is drift
// (scienceCheck) and is started again with the rest.
func Rewire(as []*Agent) (moved []string, err error) {
	return nil, nil
}

// Rekey sets again those of as, taken by OnGateway before LAN sharing was
// turned on or off or its key made anew, that reach the gateway from beyond
// loopback (an agent in a WSL distro under NAT): the key they send there is
// the sharing key while sharing is on (keyAt), so it changed with it. One
// connected while nothing was shared kept gateway.Token, which the gateway
// turns away from beyond loopback, after sharing was turned on for it.
func Rekey(as []*Agent) (moved []string, err error) {
	return nil, nil
}

// joinErrs is errs as one error, "; " between them; nil for none.
func joinErrs(errs []error) (err error) {
	if len(errs) > 0 {
		err = errs[0]
		for _, e := range errs[1:] {
			err = fmt.Errorf("%w; %w", err, e)
		}
	}
	return err
}

// wiredAt ends a stash key holding the gateway address magpie last wrote
// into an agent's config, which a move of the gateway updates.
const wiredAt = ".wired_at"
