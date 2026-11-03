package agent

import (
	"path/filepath"
	"strings"
)

// StepCode keeps its configuration in ~/.stepcode/config.toml (or
// the parent of an absolute STEP_CODING_AGENT_DIR) and its models in
// models.json using Pi's provider format.
// magpie manages defaultProvider and defaultModel in config.toml, and one
// providers.magpie entry in models.json.

func stepcode(home string) *Agent { return stepcodeIn(here(home)) }

func stepcodeDir(at place) string {
	if d := strings.TrimSpace(at.getenv("STEP_CODING_AGENT_DIR")); filepath.IsAbs(d) {
		return filepath.Dir(filepath.Clean(d))
	}
	name := ".stepcode"
	if at.spell == nil {
		if custom := strings.TrimSpace(at.getenv("STEPCODE_CONFIG_DIR")); custom != "" {
			name = custom
		}
	}
	return at.join(name)
}

func stepcodeIn(at place) *Agent {
	dir := stepcodeDir(at)
	path := filepath.Join(dir, "config.toml")

	a := &Agent{
		ID:      "stepcode",
		Name:    "StepCode",
		Icon:    "stepfun-color",
		Bin:     "step",
		Aliases: []string{"step"},
		Dir:     dir,
		Path:    path,

		UA: []string{"stepcode"},
		detect: func() bool {
			return agentDir(dir)
		},
	}

	return a
}
