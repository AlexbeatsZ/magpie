package agent

import "errors"

// ErrWiringRemoved keeps model and gateway setup manual, even when an old
// profile still names an agent. Library resource management is independent.
var ErrWiringRemoved = errors.New("automatic model and gateway configuration is disabled; configure your agent's model endpoint and key manually")
