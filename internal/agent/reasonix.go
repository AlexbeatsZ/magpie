package agent

// Reasonix Studio 2.x and its CLI share config.toml and .env under REASONIX_HOME,
// ~/.reasonix, or %APPDATA%/reasonix on Windows. Magpie owns one [[providers]]
// entry, never the user's other providers. Project/session overrides stay theirs.

import (
	"os"
	"path/filepath"
	"runtime"
)

type reasonixConfig struct {
	DefaultModel *string `toml:"default_model"`
	Agent        struct {
		Planner *string `toml:"planner_model"`
	} `toml:"agent"`
	Providers []reasonixProvider `toml:"providers"`
	Desktop   struct {
		Access []string `toml:"provider_access"`
	} `toml:"desktop"`
}

type reasonixProvider struct {
	Name      string                   `toml:"name"`
	Kind      string                   `toml:"kind"`
	BaseURL   string                   `toml:"base_url"`
	APIKeyEnv string                   `toml:"api_key_env"`
	Model     string                   `toml:"model,omitempty"`
	Models    []string                 `toml:"models,omitempty"`
	Default   string                   `toml:"default,omitempty"`
	Effort    string                   `toml:"effort,omitempty"`
	Protocol  string                   `toml:"reasoning_protocol,omitempty"`
	NoProxy   bool                     `toml:"no_proxy,omitempty"`
	Overrides map[string]reasonixModel `toml:"model_overrides,omitempty"`
}

type reasonixModel struct {
	Protocol string   `toml:"reasoning_protocol,omitempty"`
	Efforts  []string `toml:"supported_efforts,omitempty"`
	Vision   *bool    `toml:"vision,omitempty"`
	Context  int      `toml:"context_window,omitempty"`
	Output   int      `toml:"max_output_tokens,omitempty"`
}

func reasonix(home string) *Agent { return reasonixIn(here(home)) }

// reasonixIn is Reasonix at a place: this machine's home, or a WSL
// distro's (see wsl.go), where its CLI keeps Linux's ~/.reasonix
// (REASONIX_HOME isn't read) and its provider names the gateway as the
// distro reaches it, with the key it takes from there.
func reasonixIn(at place) *Agent {
	goos, getenv := runtime.GOOS, os.Getenv
	if at.spell != nil {
		goos, getenv = "linux", func(string) string { return "" }
	}
	dir := at.open(reasonixHome(goos, at.home, getenv))

	if absolute, err := filepath.Abs(dir); err == nil {
		dir = absolute
	}
	path, _ := filepath.Join(dir, "config.toml"), filepath.Join(dir, ".env")

	a := &Agent{
		ID: "reasonix", Name: "Reasonix Studio", Icon: "reasonix-color", Aliases: []string{"reasonix-studio"},
		Bin: "reasonix", Dir: dir, Path: path, UA: []string{"reasonix"},
		detect: func() bool { return reasonixDetected(at.home) },
	}

	return a
}

func reasonixHome(goos, home string, getenv func(string) string) string {
	if dir := getenv("REASONIX_HOME"); dir != "" {
		return filepath.Clean(dir)
	}
	if goos == "windows" && getenv("APPDATA") != "" {
		return filepath.Join(getenv("APPDATA"), "reasonix")
	}
	return filepath.Join(home, ".reasonix")
}

func (cfg reasonixConfig) magpie() reasonixProvider {
	for _, p := range cfg.Providers {
		if p.Name == magpieID {
			return p
		}
	}
	return reasonixProvider{}
}
