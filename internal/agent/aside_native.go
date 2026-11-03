package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/proc"
)

var asideRead = func(account string) (map[string]json.RawMessage, error) {
	if testing.Testing() {
		return nil, fmt.Errorf("Aside runtime is isolated in tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), asideTimeout)
	defer cancel()
	out, err := proc.CommandContext(ctx, "aside", "repl", "--account", account, "--host", "local",
		"console.log('MAGPIE_ASIDE_STATE '+JSON.stringify({defaultModel:aside.settings.get('defaultModel'),modelCategories:aside.settings.get('modelCategories'),imageGenerationModel:aside.settings.get('imageGenerationModel')}))").Output()
	if err != nil {
		return nil, fmt.Errorf("Aside settings could not be read: %w", err)
	}
	_, tail, ok := bytes.Cut(out, []byte("MAGPIE_ASIDE_STATE "))
	if !ok {
		return nil, fmt.Errorf("Aside returned no settings snapshot")
	}
	line, _, _ := bytes.Cut(tail, []byte("\n"))
	var values map[string]json.RawMessage
	err = json.Unmarshal(line, &values)
	return values, err
}

var asideMu sync.Mutex

type asideConnection struct {
	at                   place
	account              int // the Aside account, ~/.aside/u/<account>
	path, models, record string
	snapshot             map[string]json.RawMessage
	readErr              error
	read                 bool
}

func newAsideConnection(at place) *asideConnection {
	id := asideChoice(at)
	dir := asideAccountDir(at, id)
	// what magpie remembers of an account's settings is that account's:
	// u0's keeps the name it had before there was a choice
	record := "aside-state.json"
	if id != 0 {
		record = "aside-state-" + asideAccountName(id) + ".json"
	}
	return &asideConnection{at: at, account: id, path: filepath.Join(dir, "settings.json"), models: filepath.Join(dir, "models.json"), record: filepath.Join(filepath.Dir(stashPath()), record)}
}

func (c *asideConnection) accountID() string { return asideAccountName(c.account) }

func (c *asideConnection) runtime() (map[string]json.RawMessage, error) {
	return asideRead(c.accountID())
}

func (c *asideConnection) settings() map[string]json.RawMessage {
	if !c.read {
		c.read = true
		b, err := edit.Read(c.path)
		c.readErr = err
		c.snapshot = map[string]json.RawMessage{}
		if err == nil && len(b) > 0 {
			c.readErr = json.Unmarshal(b, &c.snapshot)
		}
	}
	return c.snapshot
}

func asideSelection(raw json.RawMessage) string {
	var m struct {
		Provider string
		ModelID  string `json:"modelId"`
	}
	_ = json.Unmarshal(raw, &m)
	if m.Provider == "" || m.ModelID == "" {
		return ""
	}
	return m.Provider + "/" + m.ModelID
}

func asideFieldPath(key string) string {
	switch key {
	case "model", "effort":
		return "defaultModel"
	case "image":
		return asideImageKey
	}
	return "modelCategories." + key
}

func asideRaw(s map[string]json.RawMessage, key string) json.RawMessage {
	if strings.HasPrefix(key, "modelCategories.") {
		var roles map[string]json.RawMessage
		_ = json.Unmarshal(s["modelCategories"], &roles)
		return roles[strings.TrimPrefix(key, "modelCategories.")]
	}
	return s[key]
}

func (c *asideConnection) value(key string) string {
	raw := asideRaw(c.settings(), asideFieldPath(key))
	if key == "effort" {
		var m struct {
			ThinkingLevel string `json:"thinkingLevel"`
		}
		_ = json.Unmarshal(raw, &m)
		return m.ThinkingLevel
	}
	return asideSelection(raw)
}

func (c *asideConnection) provider() (string, string) {
	b, err := edit.Read(c.models)
	if err != nil {
		return "foreign", "Aside's provider configuration could not be read"
	}
	if len(bytes.TrimSpace(b)) != 0 && !json.Valid(b) {
		return "foreign", "Aside's provider configuration could not be read"
	}
	raw, ok := edit.GetJSON(c.models, "providers.magpie")
	if !ok {
		return "disconnected", ""
	}
	var p struct {
		Key    string          `json:"apiKey"`
		URL    string          `json:"baseUrl"`
		API    string          `json:"api"`
		Models json.RawMessage `json:"models"`
	}
	if json.Unmarshal([]byte(raw), &p) != nil {
		return "invalid", "Aside's provider configuration could not be read"
	}
	if p.Key == "" || p.URL == "" {
		return "foreign", "Aside's magpie provider has no ownership credential; it will not be overwritten"
	}
	if p.Key != gateway.TokenFor("aside") && p.Key != agentKeyAt("aside", c.at.gw()) {
		return "foreign", "Aside's magpie provider belongs to another configuration; it will not be overwritten"
	}
	if !sameGateway(p.URL, c.at.v1()) {
		return "invalid", "Aside's magpie provider points to another gateway; reconnect explicitly to use this one"
	}
	if p.API == "" || len(p.Models) == 0 {
		return "invalid", "Aside's magpie provider has an incomplete model catalog; reconnect the provider"
	}
	return "connected", ""
}

func (c *asideConnection) block() any {
	p := magpieProviderJSONAt("pi", "aside", c.at.gw()).(map[string]any)
	p["apiKey"] = agentKeyAt("aside", c.at.gw())
	return p
}

func (c *asideConnection) sync() error {
	asideMu.Lock()
	defer asideMu.Unlock()
	status, _ := c.provider()
	if status != "connected" {
		return nil
	}
	return syncJSON(c.models, "providers.magpie", c.kept())
}

// kept is the block with what else Aside's providers.magpie holds kept
// (theirsKept).
func (c *asideConnection) kept() func() any {
	return piBlockKept(c.models, c.block)
}
