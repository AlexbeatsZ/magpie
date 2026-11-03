package agent

import (
	"encoding/json"
	"reflect"
	"slices"

	"github.com/yetone/magpie/internal/edit"
)

// Retained callers may still report a provider/catalog change. No agent files
// or background workers are maintained in this fork.
func SyncCatalog()        {}
func SyncCatalogSoon()    {}
func SyncCatalogSettled() {}

// syncJSON rewrites the value at key of a JSON file, if the file has one
// there and it differs: an agent that watches its files isn't told of a
// change that isn't one.
func syncJSON(path, key string, value func() any) error {
	cur, ok := edit.GetJSON(path, key)
	if !ok {
		return nil
	}
	v := value()
	if sameJSON(cur, v) {
		return nil
	}
	return edit.SetJSON(path, edit.KV{Path: key, Value: v})
}

// theirsKept is value, magpie's provider block for the agent, with what
// else the block at key of the JSON file holds kept: a key magpie doesn't
// write there, at any depth, is the user's or another tool's (Pi's
// providers.magpie.compat.sendSessionAffinityHeaders, which
// pi-cache-optimizer adds, #1103) and stays. The keys named in whole (the
// model list) are magpie's alone, so a model gone from the catalog leaves
// the agent's list.
func theirsKept(path, key string, value func() any, whole ...string) func() any {
	return func() any {
		v := value()
		cur, ok := edit.GetJSON(path, key)
		if !ok {
			return v
		}
		var have, want map[string]json.RawMessage
		b, err := json.Marshal(v)
		if err != nil || json.Unmarshal([]byte(cur), &have) != nil || have == nil || json.Unmarshal(b, &want) != nil || want == nil {
			return v
		}
		return mergeTheirs(have, want, whole)
	}
}

// mergeTheirs is want with each key of have that want lacks, objects in
// both merged alike; a key named in whole is want's as it is.
func mergeTheirs(have, want map[string]json.RawMessage, whole []string) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(have)+len(want))
	for k, w := range want {
		out[k] = w
	}
	for k, h := range have {
		w, ok := want[k]
		if !ok {
			out[k] = h
			continue
		}
		if slices.Contains(whole, k) {
			continue
		}
		var hm, wm map[string]json.RawMessage
		if json.Unmarshal(h, &hm) != nil || hm == nil || json.Unmarshal(w, &wm) != nil || wm == nil {
			continue
		}
		if b, err := json.Marshal(mergeTheirs(hm, wm, nil)); err == nil {
			out[k] = b
		}
	}
	return out
}

// sameJSON reports whether raw JSON says what v marshals to.
func sameJSON(raw string, v any) bool {
	b, err := json.Marshal(v)
	if err != nil {
		return false
	}
	var x, y any
	if json.Unmarshal([]byte(raw), &x) != nil || json.Unmarshal(b, &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
