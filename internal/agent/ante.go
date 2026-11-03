package agent

// Ante (antigma.ai) takes a provider of one's own in ~/.ante/catalog.json
// (ANTE_HOME moves the folder; the catalog is always $ANTE_HOME/catalog.json),
// merged on top of the providers it ships (docs.antigma.ai,
// reference/catalog-reference):
//
//	{"providers":{"<id>":{"display_name":…,"base_url":…,
//	  "wire_style":"OpenAiCompatible","auth":{"bearer":{"env_key":…}},
//	  "http_headers":{…},"extra_body":{…},
//	  "preferred_models":[{"id":…,"display_name":…,"context_limit":…,
//	    "max_tokens":…,"effort":…,"supported_efforts":[…],
//	    "support_vision":…}]}},
//	 "models":{…}}
//
// The providers map's key is what --provider takes; an id inside an entry is
// ignored. magpie adds the entry "magpie" at the gateway's /v1, on
// OpenAiCompatible — the wire style Ante speaks Chat Completions with, which
// is what the gateway takes every model on — and lists magpie's models in its
// preferred_models, Ante's model picker. The user's own providers and the
// models section are left as they are, and so is everything the user added
// inside magpie's own entry (an auth block, http_headers, extra_body,
// stream_idle_timeout_secs), which is what a gateway shared over the local
// network or reached from WSL needs to send a key (#1103, theirsKept).
//
// No auth of magpie's own: the gateway listens on loopback and lets any token
// in from this machine (gateway.Token's comment, lan.go's identifyCaller), and
// Ante counts a provider with no auth block as authenticated ("Providers with
// no `auth` block are always authenticated", catalog-reference). Writing a
// bearer env_key instead would mean magpie putting a key into Ante's
// environment, which the user would have to do themselves — the same decision
// Empryo's entry makes, which sends no key of the user's either.
//
// Being in the catalog is not enough for Ante to use it. Which provider a
// session starts on is the `provider` and `model` of ~/.ante/settings.json:
// "Explicit --provider and --model flags override the provider and model
// values in ~/.ante/settings.json. When neither source specifies a choice,
// Ante uses the auto-detection rules below… it uses the first authenticated
// provider in catalog order" (catalog-reference). Providers it ships come
// first, so magpie's is one auto-detection never picks while the user has any
// other signed in, and a user who set nothing is on Ante's own provider. So
// magpie sets those two as well and saves the pair it found there, which is
// what steps out puts back — the same pair Empryo's defaultModel and MiniMax
// Code's defaultModel are. A pick the user makes in Ante while magpie is
// wired is theirs and stands; magpie hands them what they moved off, not the
// one it read when it wired in.
//
// Ante reads the catalog once per process (restart to pick an edit up) and
// parses it leniently: an entry that doesn't deserialize is skipped with a
// notice and the rest still load, so a bad file magpie wrote costs the user
// magpie's provider, not their own.

import (
	"path/filepath"
	"strings"
)

// anteDir is the folder Ante keeps catalog.json and settings.json in:
// ANTE_HOME, else ~/.ante (catalog-reference: "~/.ante can be relocated with
// the ANTE_HOME environment variable; the catalog is always
// $ANTE_HOME/catalog.json"). A distro's ANTE_HOME isn't one magpie can read,
// so there it is ~/.ante, as MiniMax Code's mcode (place.getenv).
func anteDir(at place) string {
	if d := strings.TrimSpace(at.getenv("ANTE_HOME")); d != "" {
		return d
	}
	return at.join(".ante")
}

func ante(home string) *Agent { return anteAt(here(home)) }

// anteIn is Ante in a WSL distro (see wsl.go).
func anteIn(at place) *Agent { return anteAt(at) }

// anteAt is Ante with its home at at.home, reaching the gateway as at does.
func anteAt(at place) *Agent {

	dir := anteDir(at)
	path := filepath.Join(dir, "catalog.json")

	return &Agent{
		ID: "ante", Name: "Ante", Icon: "generic",
		Bin: "ante", Dir: dir, Path: path,
	}
}
