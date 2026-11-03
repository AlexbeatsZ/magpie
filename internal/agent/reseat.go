package agent

import (
	"regexp"
)

// A provider switched off or removed, or the last account of a sign-in
// signed out, takes its models out of the catalog, but an agent picked one
// of them keeps asking for it, and its next request is refused (#200:
// Copilot switched off, an agent still on copilot/…). Reseat moves such
// agents along with the change: to the same model from a provider still on,
// else back to the agent's own default.

// Move is one of an agent's fields moved off a model magpie no longer serves.
type Move struct {
	Agent string `json:"agent"` // the agent's name
	Field string `json:"field"` // model, or a tier (Claude Code's haiku …)
	From  string `json:"from"`  // the catalog model it was on
	To    string `json:"to"`    // the one it is on now; "" its own default
	// Error is why the agent couldn't be moved (its file unwritable, say):
	// it stays on From, which magpie no longer serves
	Error string `json:"error,omitempty"`
}

// String is the move as the CLI and the TUI say it.
func (m Move) String() string {
	if m.Error != "" {
		return m.Agent + " " + m.Field + ": still on " + m.From + ", not moved: " + m.Error
	}
	to := m.To
	if to == "" {
		to = "its default"
	}
	return m.Agent + " " + m.Field + ": " + m.From + " → " + to
}

// Reseat makes a change to the providers (change) and moves every agent
// whose model it stopped serving: to that model from another provider
// that serves it, else to the agent's own default. It answers the moves.
// Its error is the change's alone: once the change is made, an agent that
// can't be moved is a move with its Error and the others are still moved.
// An error then would say the change failed when it was made — the panel,
// told so, kept a removed provider's editor open and listed, and a second
// Remove said there was no such provider.
func Reseat(change func() error) ([]Move, error) {
	return nil, change()
}

var dated = regexp.MustCompile(`-(\d{8}|\d{4}-\d{2}-\d{2})$`)
