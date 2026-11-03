package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/sessions"
	"github.com/yetone/magpie/internal/settings"
)

func TestHistoryReferenceReachesSessionAPIs(t *testing.T) {
	h := sandboxHome(t)
	claude := filepath.Join(h, ".claude")
	t.Setenv("CLAUDE_CONFIG_DIR", claude)
	sessions.Reset()
	t.Cleanup(sessions.Reset)
	const id = "75627e81-0000-4000-8000-000000000001"
	f := filepath.Join(claude, "projects", "-work-app", id+".jsonl")
	if err := os.MkdirAll(filepath.Dir(f), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f, []byte(`{"type":"user","timestamp":"2026-10-18T10:00:00Z","message":{"role":"user","content":"reference fixture"},"cwd":"/work/app","sessionId":"`+id+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	sessionRoutes(mux, folderOnly{})
	sessionManageRoutes(mux, folderOnly{})
	request := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		return w
	}
	type row struct {
		ID        string `json:"id"`
		Reference string `json:"reference"`
	}
	var reference string
	for _, path := range []string{"/api/sessions", "/api/sessions/manage?agent=claude"} {
		w := request(path)
		var out struct{ Sessions []row }
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 || len(out.Sessions) != 1 {
			t.Fatalf("%s: %d %s (%v)", path, w.Code, w.Body, err)
		}
		r := out.Sessions[0]
		if r.ID != id || !strings.HasSuffix(r.Reference, ":claude:"+id) {
			t.Fatalf("%s: id/reference = %q/%q; want unchanged native id and agent-history reference", path, r.ID, r.Reference)
		}
		if reference != "" && r.Reference != reference {
			t.Fatalf("APIs disagree: %q != %q", r.Reference, reference)
		}
		reference = r.Reference
	}
	for _, key := range []string{reference, "[session " + strings.TrimSuffix(reference, id) + id[:8] + "]", "claude:" + id, id[:8]} {
		w := request("/api/sessions/one?key=" + url.QueryEscape(key))
		var out row
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || w.Code != 200 || out.ID != id || out.Reference != reference {
			t.Fatalf("lookup %q: %d %s (%v)", key, w.Code, w.Body, err)
		}
	}
	foreign := "rog"
	if strings.HasPrefix(reference, "rog:") {
		foreign = "omen"
	}
	w := request("/api/sessions/one?key=" + url.QueryEscape(foreign+":claude:"+id))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "agent-history-"+foreign) {
		t.Fatalf("foreign reference must name the owning server: %d %s", w.Code, w.Body)
	}
	// A prefix cannot silently select the newest of two sessions.
	secondID := strings.TrimSuffix(id, "1") + "2"
	b, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(f), secondID+".jsonl"), []byte(strings.ReplaceAll(string(b), id, secondID)), 0o600); err != nil {
		t.Fatal(err)
	}
	sessions.Reset()
	w = request("/api/sessions/one?key=" + url.QueryEscape(strings.TrimSuffix(reference, id)+id[:8]))
	if w.Code != http.StatusConflict {
		t.Fatalf("ambiguous prefix: %d %s", w.Code, w.Body)
	}
	w = request("/api/sessions/one?key=" + url.QueryEscape(reference))
	if w.Code != http.StatusOK {
		t.Fatalf("full reference after collision: %d %s", w.Code, w.Body)
	}
	if err := settings.SetSessionReads(func(*settings.SessionReads) *settings.SessionReads { return &settings.SessionReads{Asked: true} }); err != nil {
		t.Fatal(err)
	}
	sessions.Reset()
	w = request("/api/sessions/one?key=" + url.QueryEscape(reference))
	if w.Code != http.StatusNotFound || strings.Contains(w.Body.String(), "reference fixture") {
		t.Fatalf("reference bypassed consent: %d %s", w.Code, w.Body)
	}
}
