package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/library"
)

func TestResourceManagementPreservesManualModels(t *testing.T) {
	h := sandboxHome(t)
	codex := filepath.Join(h, ".codex")
	t.Setenv("CODEX_HOME", codex)
	if err := os.MkdirAll(codex, 0o700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(codex, "config.toml")
	const original = "model = \"user-model\"\nmodel_provider = \"user-relay\"\n\n[model_providers.user-relay]\nbase_url = \"https://example.invalid/v1\"\n\n[mcp_servers.own]\ncommand = \"user-command\"\n"
	if err := os.WriteFile(config, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	handler := Handler(folderOnly{}, nil)
	request := func(method, path, body string) []byte {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body)
		}
		return w.Body.Bytes()
	}
	var state struct {
		Agents []struct {
			ID        string
			Fields    []any
			Resources []string
		}
	}
	if err := json.Unmarshal(request("GET", "/api/state", ""), &state); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range state.Agents {
		if a.ID == "codex" {
			found = len(a.Fields) == 0 && len(a.Resources) == 3
		}
	}
	if !found {
		t.Fatalf("Codex resource inventory missing: %+v", state)
	}
	request("GET", "/api/library?glance=1", "")
	request("POST", "/api/library/instructions/save", `{"shared":"Use the project conventions.","agents":["codex"]}`)
	instructions, err := os.ReadFile(filepath.Join(codex, "AGENTS.md"))
	if err != nil || !strings.Contains(string(instructions), "Use the project conventions.") {
		t.Fatalf("instructions not distributed: %q %v", instructions, err)
	}
	request("POST", "/api/library/servers/save", `{"server":{"name":"resource-fixture","transport":"stdio","command":"fixture-command","agents":["codex"]}}`)
	source := filepath.Join(h, "skill-source")
	skillDir := filepath.Join(source, "resource-fixture")
	if err := os.MkdirAll(skillDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: resource-fixture\ndescription: Resource fixture\n---\nUse this fixture.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"source": source, "paths": []string{"resource-fixture"}, "agents": []string{"codex"}})
	request("POST", "/api/library/skills/install", string(body))
	if b, err := os.ReadFile(filepath.Join(codex, "skills", "resource-fixture", "SKILL.md")); err != nil || !strings.Contains(string(b), "Use this fixture.") {
		t.Fatalf("skill not distributed: %q %v", b, err)
	}
	b, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`model = "user-model"`, `model_provider = "user-relay"`, `base_url = "https://example.invalid/v1"`, `[mcp_servers.own]`, `[mcp_servers.resource-fixture]`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("config lost %q:\n%s", want, b)
		}
	}
	if strings.Contains(string(b), "magpie-codex") || strings.Contains(string(b), "model_providers.magpie") {
		t.Fatalf("resource sync wired the model:\n%s", b)
	}
	before := string(b)
	agent.SyncCatalog()
	if _, err := library.Sync(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(config)
	if err != nil || string(after) != before {
		t.Fatalf("background sync changed model/MCP configuration: %v\n%s", err, after)
	}
}
