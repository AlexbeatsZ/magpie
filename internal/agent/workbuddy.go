package agent

// WorkBuddy (Tencent's CodeBuddy desktop app) takes models of one's own
// from ~/.workbuddy/models.json ($WORKBUDDY_CONFIG_DIR if set): a list of
// OpenAI chat-completions models, or {"models":[…],"availableModels":[…]}
// as CodeBuddy Code's CLI writes it. Its own "local models" are entries of
// the same list (local: true).
//
//	[{"id":"deepseek/pro","name":…,"vendor":"magpie","apiKey":"magpie-workbuddy",
//	  "url":"http://127.0.0.1:3425/v1/chat/completions",
//	  "maxInputTokens":…,"maxOutputTokens":…,
//	  "contextWindow":{"supportedLengths":[200000,…],"defaultLength":…},
//	  "supportsToolCall":true,"supportsImages":…,"supportsReasoning":true,
//	  "reasoning":{"supportedEfforts":[…],"canDisableThinking":…,"defaultEffort":…}}]
//
// WorkBuddy shows them in its picker as "custom-local:<id>", and asks url
// for model <id>, with reasoning_effort, under Bearer apiKey. 5.6's
// app.asar watches the file only when it was there at start, and the watch
// is on the file itself, so a rewrite can go unseen until a restart; the
// Agents row offers one (apprestart.go). The model is picked per
// task in WorkBuddy's window, not in a file, so what magpie sets is whether
// its models are in that picker. Its requests say "CLI/<ver> WorkBuddy/<ver>",
// so they are told apart by their key.
//
// WorkBuddy AI, the international build (www.workbuddy.ai, bundle
// com.workbuddy.workbuddy-ai), is the same app with its data in
// ~/.workbuddy-ai: its cli/product.json sets dataFolderName ".workbuddy-ai",
// which resolveWorkbuddyConfigDir (5.6.2's app.asar) takes when
// WORKBUDDY_CONFIG_DIR isn't set. magpie wires it as its own agent, its
// models under its own key, so each build's picker is written in its own
// folder (#1494).

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/edit"
)

func workbuddy(home string) *Agent {
	dir := appdir.Getenv("WORKBUDDY_CONFIG_DIR")
	if dir == "" {
		dir = filepath.Join(home, ".workbuddy")
	}
	return workbuddyBuild("workbuddy", "WorkBuddy", dir, []string{"work-buddy"}, workbuddyApp)
}

// workbuddyAI is WorkBuddy AI, the international build, in ~/.workbuddy-ai.
// WORKBUDDY_CONFIG_DIR is WorkBuddy's (workbuddy above): both builds would
// read a folder it names, which that agent writes already.
func workbuddyAI(home string) *Agent {
	return workbuddyBuild("workbuddy-ai", "WorkBuddy AI", filepath.Join(home, ".workbuddy-ai"), []string{"workbuddy-intl", "workbuddyai"}, workbuddyAIApp)
}

// workbuddyApp and workbuddyAIApp are the two builds' bundles.
var (
	workbuddyApp   = desktopApp{id: "com.tencent.workbuddy.mac", names: []string{"WorkBuddy.app"}}
	workbuddyAIApp = desktopApp{id: "com.workbuddy.workbuddy-ai", names: []string{"WorkBuddy AI.app"}}
)

// workbuddyOnMac says whether WorkBuddy is found by its app bundle, as on a
// Mac; a var so a test forces either OS.
var workbuddyOnMac = func() bool { return runtime.GOOS == "darwin" }

// workbuddyHere reports whether a WorkBuddy build is installed. On a Mac it
// is its app bundle (app): its data folder is no sign of it. Dragging
// WorkBuddy.app to the Trash leaves ~/.workbuddy behind, and magpie writes
// there itself (models.json when wired, mcp.json and skills/ from the
// library), so the folder alone kept an uninstalled WorkBuddy on every
// library row for good (@jiahao_jayden on X: 没安装的 WorkBuddy 也在技能
// 那一排). Elsewhere, where magpie doesn't know its install folder, the
// folder still counts, as before.
func workbuddyHere(dir, path string, app desktopApp) bool {
	if Taken(dir) {
		return false
	}
	if workbuddyOnMac() {
		return macBundle(app)
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return agentDir(dir)
}

// workbuddyBuild is one of WorkBuddy's builds, its models.json in dir and
// app its macOS bundle.
func workbuddyBuild(id, name, dir string, aliases []string, app desktopApp) *Agent {
	path := filepath.Join(dir, "models.json")

	return &Agent{
		ID: id, Name: name, Icon: "workbuddy-color", Aliases: aliases,
		UA:  []string{id},
		Dir: dir, Path: path,
		detect: func() bool { return workbuddyHere(dir, path, app) },
	}
}

// workbuddyDoc is models.json as read: its entries, and the rest of the
// object when it is one (nil for a bare list).
type workbuddyDoc struct {
	models []json.RawMessage
	rest   map[string]json.RawMessage
}

func workbuddyRead(path string) (workbuddyDoc, error) {
	var d workbuddyDoc
	b, err := edit.Read(path)
	if err != nil || len(bytes.TrimSpace(b)) == 0 {
		return d, err
	}
	if bytes.TrimSpace(b)[0] == '[' {
		return d, json.Unmarshal(b, &d.models)
	}
	if err := json.Unmarshal(b, &d.rest); err != nil {
		return d, err
	}
	if raw, ok := d.rest["models"]; ok {
		if err := json.Unmarshal(raw, &d.models); err != nil {
			return d, err
		}
	}
	return d, nil
}

// workbuddyEntry is what magpie reads of an entry.
type workbuddyEntry struct {
	ID        string `json:"id"`
	Vendor    string `json:"vendor"`
	Disabled  *bool  `json:"disabled"`
	URL       string `json:"url"`
	APIKey    string `json:"apiKey"`
	Reasoning *struct {
		DefaultEffort string `json:"defaultEffort"`
		Effort        string `json:"effort"` // the older key WorkBuddy still reads
	} `json:"reasoning"`
}

func workbuddyMine(raw json.RawMessage) (workbuddyEntry, bool) {
	var e workbuddyEntry
	return e, json.Unmarshal(raw, &e) == nil && e.Vendor == magpieID
}

// workbuddyPart is magpie's part of models.json, in one form whoever wrote
// the file: magpie's entries, each with its keys sorted and without its
// apiKey. WorkBuddy rewrites models.json itself seconds after every start
// (it purges local models a computer can't run, and saves the file again
// even when nothing was purged): without magpie's trailing newline, as a
// bare list on 5.6, and with the apiKey encrypted where its credential
// policy says so. Its own write is no change of magpie's, so its row is not
// told to restart WorkBuddy for it. The file as it is when it can't be
// read as models.json.
func workbuddyPart(path string) string {
	d, err := workbuddyRead(path)
	if err != nil {
		c, _ := edit.Read(path)
		return string(c)
	}
	var b strings.Builder
	for _, raw := range d.models {
		if _, ok := workbuddyMine(raw); !ok {
			continue
		}
		var e map[string]any
		if json.Unmarshal(raw, &e) != nil {
			continue
		}
		delete(e, "apiKey")
		c, _ := json.Marshal(e)
		b.Write(c)
		b.WriteByte('\n')
	}
	return b.String()
}
