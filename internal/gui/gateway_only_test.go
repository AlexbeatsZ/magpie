package gui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/testenv"
)

func TestGatewayOnlyAPIRetiresConfigRoutes(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	cfg := filepath.Join(home, ".codex", "config.toml")
	const original = "# user's settings\nmodel = \"own-model\"\n"
	if err := os.MkdirAll(filepath.Dir(cfg), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	h := Handler(nil, nil)
	for _, route := range []string{
		"/api/set", "/api/agents/connect/codex", "/api/agents/reapply/codex",
		"/api/agents/stage/aside", "/api/agents/disconnect/codex",
		"/api/profile/use", "/api/tune/apply",
		"/api/tune/undo", "/api/settings/codex-async-compact",
		"/api/settings/codex-compact-held/drop", "/api/settings/gateway-mode",
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, route, strings.NewReader(`{"agent":"codex","field":"model","value":"relay/test","mode":"off"}`)))
		if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s is still available: %d %s", route, w.Code, w.Body.String())
		}
	}
	got, err := os.ReadFile(cfg)
	if err != nil || string(got) != original {
		t.Fatalf("client config changed: %q, %v", got, err)
	}
}

func TestGatewayOnlyDesktopAndWebBoot(t *testing.T) {
	testenv.SetHome(t, t.TempDir())
	s := settings.Load()
	s.GatewayMode = "off" // a saved upstream preference cannot bring wiring back
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
	for _, host := range []Windows{nil, webHost{}} {
		w := httptest.NewRecorder()
		Handler(host, nil).ServeHTTP(w, httptest.NewRequest("GET", "/boot.js", nil))
		for _, want := range []string{`"manualModels":true`} {
			if !strings.Contains(w.Body.String(), want) {
				t.Errorf("boot lacks %s: %s", want, w.Body.String())
			}
		}
		if strings.Contains(w.Body.String(), `"gatewayOnly":true`) || strings.Contains(w.Body.String(), `"gateway":true`) {
			t.Errorf("resource pages hidden: %s", w.Body.String())
		}
	}
}
