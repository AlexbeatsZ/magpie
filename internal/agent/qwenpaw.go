package agent

// QwenPaw (agentscope-ai/QwenPaw, formerly CoPaw) is a personal assistant
// app with agents of its own. Everything is read from its working folder
// (constant.py): QWENPAW_WORKING_DIR, else ~/.copaw when a legacy one is
// there, else ~/.qwenpaw. Its secret folder is QWENPAW_SECRET_DIR, which
// <working folder>/.env may set too (loaded after the working folder is
// known), else "<working folder>.secret" (~/.qwenpaw.secret).
//
// <working folder>/config.json lists the agents, agents.profiles
// {id: {id, workspace_dir, enabled}} in agents.agent_order. "default" is
// always there, in <working folder>/workspaces/default. Each agent's
// <workspace_dir>/agent.json names its model:
//
//	"active_model": {"provider_id": "…", "model": "…"}
//
// beside fallback_models and subagent_model, which magpie leaves alone,
// and thinking_level, which is the agent's effort field. Only an agent
// whose backend is "qwenpaw" (the default) runs that model
// (app/routers/agents.py); one on an ACP backend runs another program.
// QwenPaw reads agent.json again when it changes
// (load_agent_config's fingerprint), and asks for the model on every turn
// (agents/model_factory.py create_model_and_formatter), so a model picked
// there is used from the next message.
//
// Providers are files under <secret folder>/providers/: builtin/<id>.json,
// custom/<id>.json, and active_model.json, the model of an agent with no
// active_model (provider_manager_persistence.py). A custom provider is the
// Provider model as QwenPaw saves it from its console's "Add provider":
// id, name, base_url, api_key, chat_model ("OpenAIChatModel" is OpenAI's
// Chat Completions), is_custom, and the models the user added in
// extra_models. magpie writes custom/magpie.json, OpenAI Chat at the
// gateway's /v1 with a key naming QwenPaw and one model of extra_models
// for each of magpie's, and sets an agent's active_model to
// {"provider_id": "magpie", "model": "<provider>/<model>"}. QwenPaw keeps
// api_key encrypted (security/secret_store.py, "ENC:…") and takes one
// written in plain text, encrypting it at its next start
// (_maybe_migrate_plaintext): that is how its own files hold a key, so the
// key goes nowhere QwenPaw doesn't keep it. What the user changes in
// magpie's provider in QwenPaw's console (a model's window, its generation
// settings, the provider's headers) stays.
//
// QwenPaw reads the provider files once, at start (ProviderManager), so the
// first connect needs QwenPaw restarted; until then a turn on magpie's
// model fails with "Provider 'magpie' not found".
//
// Each agent is a row: "qwenpaw" is the default agent, any other
// qwenpaw#<id>, as omp's named profiles are.

import (
	"cmp"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/yetone/magpie/internal/edit"
)

// qwenpawDefault is the agent QwenPaw always has, which can't be removed.
const qwenpawDefault = "default"

// qwenpawWorking is QwenPaw's working folder: QWENPAW_WORKING_DIR, else a
// legacy ~/.copaw, else ~/.qwenpaw. QwenPaw resolves it (Path.resolve), so
// a link is followed: the secret folder is named after where it leads.
func qwenpawWorking(at place) string {
	d := strings.TrimSpace(at.getenv("QWENPAW_WORKING_DIR"))
	switch {
	case d == "~":
		d = at.home
	case strings.HasPrefix(d, "~/"), strings.HasPrefix(d, `~\`):
		d = filepath.Join(at.home, d[2:])
	case d == "":
		d = at.join(".qwenpaw")
		if legacy := at.join(".copaw"); at.exists(legacy) {
			d = legacy
		}
	}
	if r, err := filepath.EvalSymlinks(d); err == nil {
		return r
	}
	return d
}

// qwenpawProfile is one of QwenPaw's agents: its id and its workspace.
type qwenpawProfile struct{ id, dir string }

// qwenpawProfiles are the agents config.json lists, in agent_order, the
// default first; those switched off are left out. With no config.json yet
// it is the default alone, in its default workspace.
func qwenpawProfiles(at place, work string) []qwenpawProfile {
	def := qwenpawProfile{qwenpawDefault, filepath.Join(work, "workspaces", qwenpawDefault)}
	raw, err := os.ReadFile(filepath.Join(work, "config.json"))
	if err != nil || !json.Valid(raw) {
		return []qwenpawProfile{def}
	}
	profiles := gjson.GetBytes(raw, "agents.profiles")
	dirOf := func(id string) string {
		d := strings.TrimSpace(profiles.Get(gjson.Escape(id) + ".workspace_dir").String())
		switch {
		case d == "":
			return filepath.Join(work, "workspaces", id)
		case strings.HasPrefix(d, "~/"), strings.HasPrefix(d, `~\`):
			return filepath.Join(at.home, d[2:])
		}
		return d
	}
	var ids []string
	for _, id := range gjson.GetBytes(raw, "agents.agent_order").Array() {
		ids = append(ids, id.String())
	}
	var rest []string
	profiles.ForEach(func(k, _ gjson.Result) bool {
		if !slices.Contains(ids, k.String()) {
			rest = append(rest, k.String())
		}
		return true
	})
	sort.Strings(rest)
	ids = append(ids, rest...)
	out := []qwenpawProfile{{qwenpawDefault, dirOf(qwenpawDefault)}}
	for _, id := range ids {
		if id == qwenpawDefault || !qwenpawIDOK(id) || slices.ContainsFunc(out, func(p qwenpawProfile) bool { return p.id == id }) {
			continue
		}
		if e := profiles.Get(gjson.Escape(id) + ".enabled"); e.Exists() && !e.Bool() {
			continue
		}
		out = append(out, qwenpawProfile{id, dirOf(id)})
	}
	return out
}

// qwenpawIDOK is an agent id as QwenPaw takes one (agents.py's
// _AGENT_ID_RE): letters, digits, _ and -, 2 to 64 of them, neither end a
// _ or -.
func qwenpawIDOK(id string) bool {
	if len(id) < 2 || len(id) > 64 {
		return false
	}
	for i, r := range id {
		alnum := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
		if !alnum && (i == 0 || i == len(id)-1 || r != '_' && r != '-') {
			return false
		}
	}
	return true
}

func qwenpaw(home string) *Agent {
	at := here(home)
	work := qwenpawWorking(at)
	return qwenpawAt(at, work, qwenpawProfiles(at, work)[0])
}

// qwenpawAgents are the rows of QwenPaw's agents other than the default.
func qwenpawAgents(home string) []*Agent {
	at := here(home)
	work := qwenpawWorking(at)
	if !isDir(work) {
		return nil
	}
	var out []*Agent
	for _, p := range qwenpawProfiles(at, work)[1:] {
		agentJSON := filepath.Join(p.dir, "agent.json")
		if !isFile(agentJSON) {
			continue
		}
		if b := qwenpawBackend(agentJSON); b != "qwenpaw" {
			continue
		}
		a := qwenpawAt(at, work, p)
		name, _ := edit.GetJSON(agentJSON, "name")
		a.ID, a.Name = "qwenpaw#"+p.id, "QwenPaw · "+cmp.Or(strings.TrimSpace(name), p.id)
		a.Aliases, a.UA, a.Bin = nil, nil, ""
		a.detect = func() bool { return isFile(agentJSON) }
		out = append(out, a)
	}
	return out
}

// qwenpawBackend is the backend an agent.json names, "qwenpaw" when none.
func qwenpawBackend(path string) string {
	b, _ := edit.GetJSON(path, "backend")
	return cmp.Or(strings.TrimSpace(b), "qwenpaw")
}

func qwenpawAt(_ place, work string, p qwenpawProfile) *Agent {

	path := filepath.Join(p.dir, "agent.json")

	return &Agent{
		ID: "qwenpaw", Name: "QwenPaw", Icon: "qwenpaw",
		Bin: "qwenpaw", Dir: work, Path: path,
	}
}
