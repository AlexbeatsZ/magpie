// Package profile preserves legacy profile documents for backup and sync.
package profile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/lastgood"
	"github.com/yetone/magpie/internal/library"
	"github.com/yetone/magpie/internal/provider"
)

// Profile is every agent's fields, and the library's setup.
type Profile struct {
	// Fields maps "agent.field" to a value.
	Fields map[string]string
	// Library is which agents got which servers and skills, and the
	// instructions; nil for a profile saved while the library was empty,
	// or before profiles kept it, which leaves the library as it is.
	Library *library.Setup
	// Models are the agents' model lists (#1368), by the id their lists are
	// kept under (agent.ListsFor): which of magpie's models each shows and
	// in what order. An agent not named, every one in a profile saved
	// before profiles kept them, keeps the list it has.
	Models map[string]provider.ModelList
}

// libraryKey holds the library's setup among a profile's "agent.field"
// keys: it has no dot, so it is no agent's field, and a profile saved
// before it is read as it always was.
const libraryKey = "library"

// modelsKey holds the agents' model lists, as libraryKey the library's.
const modelsKey = "models"

// MarshalJSON writes the fields and the library's setup side by side.
func (p Profile) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, len(p.Fields)+1)
	for k, v := range p.Fields {
		m[k] = v
	}
	if p.Library != nil {
		m[libraryKey] = p.Library
	}
	if p.Models != nil {
		m[modelsKey] = p.Models
	}
	return json.Marshal(m)
}

// UnmarshalJSON reads what MarshalJSON writes, and a profile of before,
// which was the fields alone.
func (p *Profile) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*p = Profile{Fields: map[string]string{}}
	for k, raw := range m {
		if k == libraryKey {
			if string(raw) == "null" {
				continue
			}
			p.Library = &library.Setup{}
			if err := json.Unmarshal(raw, p.Library); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
			continue
		}
		if k == modelsKey {
			if string(raw) == "null" {
				continue
			}
			if err := json.Unmarshal(raw, &p.Models); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
			continue
		}
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
		p.Fields[k] = v
	}
	return nil
}

// Path is the profiles file.
func Path() string { return filepath.Join(appdir.Config(), "profiles.json") }

// Load reads every profile.
func Load() (map[string]Profile, error) {
	b, err := edit.Read(Path())
	if err != nil {
		return nil, err
	}
	out := map[string]Profile{}
	if len(b) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(b, &out); err != nil {
		// one a crash left all zero (#1505) is its last good generation
		bak, ferr := lastgood.Fallback(Path(), b, validProfiles)
		if ferr != nil {
			return nil, fmt.Errorf("%s: %w", Path(), err)
		}
		out = map[string]Profile{}
		if err := json.Unmarshal(bak, &out); err != nil {
			return nil, fmt.Errorf("%s: %w", Path(), err)
		}
	}
	return out, nil
}

func validProfiles(b []byte) bool {
	var m map[string]Profile
	return json.Unmarshal(b, &m) == nil
}

// Names lists profiles alphabetically.
func Names(ps map[string]Profile) []string {
	names := make([]string, 0, len(ps))
	for n := range ps {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func store(ps map[string]Profile) error {
	b, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return err
	}
	if err := lastgood.Keep(Path(), validProfiles, 0o600); err != nil {
		return err
	}
	return edit.WriteAtomic(Path(), append(b, '\n'))
}

// Save stores p under name, replacing any existing profile.
func Save(name string, p Profile) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("profile name is empty")
	}
	ps, err := Load()
	if err != nil {
		return err
	}
	ps[name] = p
	return store(ps)
}

// Delete removes a profile.
func Delete(name string) error {
	ps, err := Load()
	if err != nil {
		return err
	}
	if _, ok := ps[name]; !ok {
		return fmt.Errorf("no profile named %q", name)
	}
	delete(ps, name)
	return store(ps)
}

// Applied is what applying a profile did.
type Applied struct {
	Changed int // agent fields changed, and model lists
	// Library is what bringing the library's setup back wrote into the
	// agents; nil for a profile that carries none.
	Library *library.Result
}
