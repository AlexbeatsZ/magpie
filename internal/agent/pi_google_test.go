package agent

import (
	"slices"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
)

// Raft runs Pi's SDK on Pi's own agent folder and hands each request a
// fetch of its own, which pi-ai's Google adapter refuses before sending
// ("Custom fetch is not supported by the Google Generative AI adapter",
// yunfei9958 on X), so a Gemini model in Pi's models.json stays on Chat,
// and Pi offers it the levels Chat has; OmO and Prime Agent, whose
// folders are their own, keep Gemini's API.
func TestPiFolderKeepsGeminiOnChatForRaft(t *testing.T) {
	m := catalog.Model{ID: "gem/gemini-3.8-flash", Name: "Gemini 3.8 Flash", APIs: []string{"gemini"}, Efforts: []string{"low", "high", "xhigh"}}
	const gw = "http://127.0.0.1:1"
	for id, want := range map[string]any{"pi": nil, "stepcode": nil, "omo": "google-generative-ai", "prime-agent": "google-generative-ai"} {
		e := piModelJSON(piAsked(id, m), gw, true)
		if e["api"] != want {
			t.Errorf("%s: api %v, want %v", id, e["api"], want)
		}
		if want == nil && (e["baseUrl"] != nil || e["thinkingLevelMap"].(map[string]any)["xhigh"] != "xhigh") {
			t.Errorf("%s: not Chat's entry: %v", id, e)
		}
	}
	if !slices.Equal(m.APIs, []string{"gemini"}) {
		t.Errorf("the catalog's model changed: %v", m.APIs)
	}
}
