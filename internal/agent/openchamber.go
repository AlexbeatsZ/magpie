package agent

// OpenChamber (a desktop, web and VS Code front end for OpenCode, #321) runs
// an OpenCode of its own — the desktop app bundles OpenCode 2 — on the
// user's OpenCode config: $OPENCODE_CONFIG_DIR, else
// $XDG_CONFIG_HOME/opencode (openCodeDir). Its own settings are under
// $OPENCHAMBER_DATA_DIR, ~/.config/openchamber by default (not moved by
// XDG_CONFIG_HOME):
//
//	preferences.json  {"version":1,"fields":{"<key>":{"value":…,"updatedAt":<ms>}}}: the profile
//	settings.json     the instance's facts, plus a copy of the profile's values for older builds
//
// Each new session is sent the model it picks, in this order: the project's
// defaultModel (settings.json projects[]), the profile's defaultModel, the
// agent's own, and only then OpenCode's model — so a defaultModel of its own
// wins over the one magpie sets for OpenCode. Its small model (titles, commit
// messages) is smallModelOverride once smallModelUseDefault is false, else
// OpenCode's small_model. A model is "provider/model" as OpenCode names it,
// so one of magpie's needs magpie's provider in the OpenCode config.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/edit"
)

// openChamberDir is where OpenChamber keeps its settings.
func openChamberDir(home string) string {
	if d := appdir.Getenv("OPENCHAMBER_DATA_DIR"); d != "" {
		if abs, err := filepath.Abs(d); err == nil {
			return abs
		}
	}
	return filepath.Join(home, ".config", "openchamber")
}

// ocStore reads and writes OpenChamber's profile as it does: a key in
// preferences.json wins over its copy in settings.json, which is all a
// build from before the split reads, and seeds preferences.json when that
// is missing.
type ocStore struct{ dir string }

func (s ocStore) prefs() string    { return filepath.Join(s.dir, "preferences.json") }
func (s ocStore) settings() string { return filepath.Join(s.dir, "settings.json") }

// hasPrefs says whether preferences.json is there; an error when it is
// but isn't a version-1 document, which OpenChamber doesn't write to either.
func (s ocStore) hasPrefs() (bool, error) {
	b, err := os.ReadFile(s.prefs())
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var doc struct {
		Version int             `json:"version"`
		Fields  json.RawMessage `json:"fields"`
	}
	if json.Unmarshal(b, &doc) != nil || doc.Version != 1 || len(doc.Fields) == 0 || doc.Fields[0] != '{' {
		return false, fmt.Errorf("OpenChamber's %s isn't a preferences file it can read; fix or remove it first", s.prefs())
	}
	return true, nil
}

func (s ocStore) get(key string) string {
	if v, ok := edit.GetJSON(s.prefs(), "fields."+key+".value"); ok {
		return v
	}
	v, _ := edit.GetJSON(s.settings(), key)
	return v
}

// set writes key = v, or takes key out for a nil v, in both files: the
// copy in settings.json only where that file is, or the profile has
// nowhere else to go.
func (s ocStore) set(kvs map[string]any) error {
	has, err := s.hasPrefs()
	if err != nil {
		return err
	}
	_, statErr := os.Stat(s.settings())
	for key, v := range kvs {
		if has {
			if v == nil {
				err = edit.DelJSON(s.prefs(), "fields."+key)
			} else {
				err = edit.SetJSON(s.prefs(), edit.KV{Path: "fields." + key + ".value", Value: v},
					edit.KV{Path: "fields." + key + ".updatedAt", Value: time.Now().UnixMilli()})
			}
			if err != nil {
				return err
			}
		}
		if has && statErr != nil {
			continue
		}
		if v == nil {
			err = edit.DelJSON(s.settings(), key)
		} else {
			err = edit.SetJSON(s.settings(), edit.KV{Path: key, Value: v})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// model and small are the models OpenChamber sends, "" when it leaves them
// to OpenCode's config.
func (s ocStore) model() string { return s.get("defaultModel") }

func openChamber(home, _ string) *Agent {
	s := ocStore{openChamberDir(home)}

	return &Agent{
		ID: "openchamber", Name: "OpenChamber", Icon: "openchamber", Aliases: []string{"chamber"},
		// its model requests are its OpenCode's, and counted as OpenCode's
		Bin: "openchamber", Dir: s.dir, Path: s.prefs(),
	}
}
