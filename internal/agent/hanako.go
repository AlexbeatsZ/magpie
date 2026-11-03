package agent

// OpenHanako (HanaAgent, a desktop personal agent) keeps what it knows under
// $HANA_HOME, ~/.hanako by default:
//
//	provider-catalog.json                  {"catalogVersion":2,"providers":{"<id>":{…}},…}: the providers it reaches
//	provider-plugins/<id>/providers/<id>.json  a provider added by hand, moved out of the catalog at start-up
//	                                       (all but its key and headers, which stay in the catalog)
//	user/preferences.json                  primaryAgent: the agent it opens on
//	agents/<id>/config.yaml                that agent's settings; models.chat is {id, provider}
//	server-info.json                       {port, token, …}, while it runs
//
// Since 1.0 OpenHanako seals the catalog under a key the desktop app keeps
// in the system keychain: the file is then "HANA-SECRET-1.AES-256-GCM.…",
// not JSON. magpie can't read it nor write it, and never writes over it. A
// provider's definition is in its plugin file, out of the catalog, so that
// is still magpie's to keep up to date; its key stays sealed, and adding or
// removing magpie needs OpenHanako running, to do it through its API.
//
// magpie is one provider there, magpie, spoken to as chat completions at the
// gateway's /v1 with the catalog as its models; choosing one of them makes it
// the primary agent's chat model. While OpenHanako runs, both go through its
// local API, which saves them and reloads its models at once:
//
//	PUT /api/config               {"providers":{"magpie":{…}|null}}
//	PUT /api/agents/:id/config    {"models":{"chat":{"id","provider"}}}
//	GET /api/models               {"models":[{"id","provider",…}]}
//
// Otherwise magpie writes the files, which it reads when it starts. The
// model the agent had is stashed and put back when magpie steps out; since
// 1.0 the API takes no agent without a model, so with none stashed, or one
// it no longer has, the agent goes to the first of OpenHanako's own.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/yetone/magpie/internal/appdir"
)

// hanakoClient waits longer: a provider change reloads every model.
var hanakoClient = &http.Client{Timeout: 15 * time.Second}

func hanakoHome(home string) string {
	dir := appdir.Getenv("HANA_HOME")
	if dir == "" {
		return filepath.Join(home, ".hanako")
	}
	if rest, ok := strings.CutPrefix(dir, "~"); ok && (rest == "" || rest[0] == '/' || rest[0] == filepath.Separator) {
		dir = home + rest
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return dir
}

func hanako(home string) *Agent {
	dir := hanakoHome(home)
	path := filepath.Join(dir, "provider-catalog.json")

	return &Agent{
		ID: "hanako", Name: "OpenHanako", Icon: "hanako", Aliases: []string{"openhanako", "hana", "hanaagent"},
		// what its provider client sends, though the key names it first
		UA:  []string{"hanaagent"},
		Dir: dir, Path: path,
	}
}

type hanakoAvail struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
}

// models lists the models OpenHanako can chat with, as its model picker
// does: GET /api/models.
func (s *hanakoServer) models() ([]hanakoAvail, error) {
	req, err := http.NewRequest("GET", s.url("/api/models"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := hanakoClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenHanako: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		Models []hanakoAvail `json:"models"`
		Error  string        `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&out); err != nil || resp.StatusCode/100 != 2 {
		if out.Error == "" && err != nil {
			out.Error = err.Error()
		}
		return nil, fmt.Errorf("OpenHanako: %s /api/models: %s", resp.Status, out.Error)
	}
	return out.Models, nil
}

// hanakoServer is a running OpenHanako's API, as server-info.json gives it.
type hanakoServer struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
}

func (s *hanakoServer) url(path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", s.Port, path)
}

func (s *hanakoServer) put(path string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", s.url(path), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := hanakoClient.Do(req)
	if err != nil {
		return fmt.Errorf("OpenHanako: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 == 2 {
		return nil
	}
	out, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	msg := strings.TrimSpace(string(out))
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(out, &e) == nil && e.Error != "" {
		msg = e.Error
	}
	return fmt.Errorf("OpenHanako: %s %s: %s", resp.Status, path, msg)
}
