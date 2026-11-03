package agent

import (
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// Every agent's model picker has two halves: the models the agent reaches
// on its own (its sign-in, its own keys), and magpie's catalog — every model
// of every provider the user added, reached through the local gateway.
// A catalog model is spelled "provider/model"; an agent that sends that to
// the gateway gets the vendor's reply in whichever API the agent speaks.

// magpieID is the provider id agents know the gateway by.
const magpieID = "magpie"

// RoutingGroups is the picker's group of routing groups.
const RoutingGroups = "Routing groups"

// isMagpie reports whether a model value is a catalog reference.
func isMagpie(v string) bool {
	if pid, _, ok := strings.Cut(v, "/"); ok && dryProviders[pid] {
		return true
	}
	if !strings.Contains(v, "/") {
		return false
	}
	return magpieRefs.get(v, func() bool {
		_, _, ok := provider.Resolve(v)
		return ok
	})
}

// magpieRefs keeps isMagpie's answers while provider.StampNow stays the
// same, for a minute at most. Outside a request that holds the catalog,
// each Resolve built every provider's entries again: the window's drift
// poll, every 15s, asked it for each agent's values, ~6 MB a poll (the
// procmem heap profile: 235 MB of 1.37 GB in 10 minutes of GUI use).
var magpieRefs stampedRefs

// stampedRefs is a model value → answer, dropped when the stamp moves,
// after a minute (what the stamp can't see: an account another program
// signed in to), or once it holds more values than any config has.
type stampedRefs struct {
	mu    sync.Mutex
	stamp provider.Stamp
	at    time.Time
	of    map[string]bool
}

const (
	stampedRefsFor = time.Minute
	stampedRefsMax = 4096
)

func (s *stampedRefs) get(v string, build func() bool) bool {
	st := provider.StampNow()
	s.mu.Lock()
	if s.stamp != st || time.Since(s.at) >= stampedRefsFor || len(s.of) >= stampedRefsMax {
		s.stamp, s.at, s.of = st, time.Now(), nil
	}
	if ok, found := s.of[v]; found {
		s.mu.Unlock()
		return ok
	}
	s.mu.Unlock()
	// built from what is there after st was taken: never older than st
	ok := build()
	s.mu.Lock()
	if s.stamp == st {
		if s.of == nil {
			s.of = map[string]bool{}
		}
		s.of[v] = ok
	}
	s.mu.Unlock()
	return ok
}

func firstOf(xs []string) string {
	if len(xs) > 0 {
		return xs[0]
	}
	return ""
}

// magpieModels is the catalog as agent is shown it, as catalog.Models, for
// agents that keep their own model files: built again only when
// provider.StampNow moved, or magpieModelsFor after it was, each caller
// given its own copy of the list.
func magpieModels(agent string) []catalog.Model {
	return builtModels.get(agent, func() []catalog.Model { return buildModels(agent) })
}

// sharedModels is magpieModels without the caller's own copy, for one
// that only reads it: a range over it copies each model, so it can't
// write into the list kept. A copy per call was 37 MB of what forty polls
// of the window allocated with real accounts, most of it Pi's and Claude
// Code's thinking levels, read per field.
func sharedModels(agent string) []catalog.Model {
	return builtModels.shared(agent, func() []catalog.Model { return buildModels(agent) })
}

// builtModels is what magpieModels built last, per agent. The window asks
// /api/drift every 15s, and each agent's Drift asked for its list again
// per field, each built whole with its prices and Vision's ranking of the
// models that see: 14.5 GB of the 33 GB a live magpie allocated in 53
// minutes, the garbage its footprint peaks on.
var builtModels stampedLists[catalog.Model]

// stampedLists keeps a list built per key while provider.StampNow stays
// the same, magpieModelsFor at most, and gives each caller a copy.
type stampedLists[T any] struct {
	sync.Mutex
	m map[string]stampedList[T]
}

type stampedList[T any] struct {
	list  []T
	stamp provider.Stamp
	at    time.Time
}

func (s *stampedLists[T]) get(key string, build func() []T) []T {
	return slices.Clone(s.shared(key, build))
}

// shared is get without the caller's own copy: the list kept, read only.
func (s *stampedLists[T]) shared(key string, build func() []T) []T {
	s.Lock()
	b, ok := s.m[key]
	s.Unlock()
	// the stamp is read before the list is built: a change while it is
	// moves it past the one kept, and the next call builds it again
	stamp := provider.StampNow()
	if !ok || b.stamp != stamp || time.Since(b.at) >= magpieModelsFor {
		b = stampedList[T]{build(), stamp, time.Now()}
		s.Lock()
		if s.m == nil {
			s.m = map[string]stampedList[T]{}
		}
		s.m[key] = b
		s.Unlock()
	}
	return slices.Clip(b.list)
}

// magpieModelsFor is how long a list is kept while provider.StampNow says
// nothing changed: what the stamp can't see (another app rewriting an
// account's file, a model retired by its date, a provider resting, which
// Vision's pick of a model that sees follows) reaches the lists within it.
const magpieModelsFor = time.Minute

// buildModels is magpieModels built afresh. A model's own lists are
// clipped, so a caller appending to one in its copy makes a new one and
// never writes into the list kept.
func buildModels(agent string) []catalog.Model {
	// the providers built once for all the models' prices, not for each:
	// dsh's round at every start built them for each of hundreds
	defer provider.Hold()()
	var out []catalog.Model
	shown, _ := provider.CatalogFor(agent)
	labels := provider.Labels(shown)
	// a model magpie describes images to takes them (provider.Described)
	seen := describedNow()
	st, find := settings.Load(), provider.GroupFinder()
	priceOf := provider.AgentPricer(st, find)
	members := func(id string) []provider.Member {
		_, ms, _ := find(id)
		return ms
	}
	for i, e := range shown {
		efforts := e.Efforts
		if agent == "codex" {
			// the levels Codex's picker offers, a model magpie has no
			// word on included (provider.CodexEfforts)
			efforts = provider.CodexEfforts(e)
		}
		m := catalog.Model{ID: e.ID, Name: labels[i], Provider: firstOf(e.Provider.Catalogs()), Efforts: efforts, Images: e.Images || seen, ImageInput: e.ImageInput, Context: e.Context, Output: e.Output, OutputSet: e.OutputSet, AgentsV2: e.AgentsV2, Reasoning: e.Thinks(),
			AlwaysThinks: e.AlwaysThinks, DynamicTools: e.DynamicTools, DefaultEffort: e.DefaultEffort}
		// where Codex and Claude Code compact it (#876): a threshold the
		// user set on the model or on its provider, else a window they
		// typed whole (#1809), else the one for every model, which codexcat
		// answers from the settings it is given
		m.Compact = provider.CompactFor(st, e, find)
		// what a call costs the user (Pi's cost, #781)
		if pr, ok := priceOf(e); ok {
			m.Price = &pr
		}
		if seen && !e.Images {
			yes := true
			m.ImageInput = &yes
		}
		// APIs is the preferred client protocol, preserving the vendor's
		// own content where possible. A group's members may differ, so a group has
		// one only when every member is served best on Responses (#1594,
		// ObsidianArch02): Pi, omp, droid and dsh then ask it there, and a
		// GPT group's tool-call arguments stream as the member sends them
		// instead of being held for a Chat answer
		if e.Group == "" {
			if n := e.Provider.Native(e.Model); n != "" {
				m.APIs = []string{string(n)}
			} else {
				// Gemini clients also reach Code Assist and Google plugins
				// through the gateway's Gemini API, with translation as needed.
				apis := e.Provider.APIs(e.Model)
				for _, pr := range e.Provider.Speaks() {
					if (pr == provider.Gemini || pr == provider.CodeAssist) && (apis == nil || slices.Contains(apis, pr)) {
						m.APIs = []string{string(provider.Gemini)}
						break
					}
				}
			}
		} else if allResponses(members(e.ID)) {
			m.APIs = []string{string(provider.Responses)}
		}
		// the service tiers Codex is offered on it, as the gateway's own
		// lists offer them: Fast in the Codex CLI magpie writes
		// magpie-models.json for (#1530)
		provider.CodexTiers(&m, e, members)
		if agent == "codex" {
			// the entry its vendor lists the model under in Codex's own
			// shape, handed on as it is (codexcat)
			m.Codex = e.Codex
		}
		// the gateway paths it goes on untranslated, as /v1/models says
		// them: what an agent plugin's connect is told (cleary5779)
		m.Endpoints = provider.NativeEndpoints(e)
		// Codex's /fast on a ChatGPT account's own GPT model, which
		// OpenCode has no command for and takes as variants (#2379); a
		// group is sent fast as the group says, as a fast pick is
		if a := e.Provider.Account; e.Group == "" && a != nil && a.Agent == "codex" {
			m.CodexFast = provider.CanFast(e.Provider, e.Model)
		}
		m.Efforts, m.Keys, m.APIs, m.Endpoints = slices.Clip(m.Efforts), slices.Clip(m.Keys), slices.Clip(m.APIs), slices.Clip(m.Endpoints)
		m.InputModalities, m.OutputModalities, m.Tiers, m.Codex = slices.Clip(m.InputModalities), slices.Clip(m.OutputModalities), slices.Clip(m.Tiers), slices.Clip(m.Codex)
		out = append(out, m)
	}
	return out
}

// describedNow is provider.Described, asked once for every agent's list
// while provider.StampNow stays the same, magpieModelsFor at most, as the
// lists are kept: each list built asked the gateway's Vision pick again,
// which ranks every provider's models that see, 553 MB of what a live
// magpie allocated in 37 minutes for the same answer a dozen times over.
func describedNow() bool {
	return builtSeen.shared("", func() []bool {
		return []bool{provider.Described != nil && provider.Described()}
	})[0]
}

var builtSeen stampedLists[bool]

// allResponses reports whether a group's members are all served best on
// Responses, so the gateway relays a Responses request to whichever it
// picks. A group with none, or one member on another API, is false: a
// Messages-only group stays on Chat too, as the thinking a Claude member
// takes is known by its own model, not by the group's id.
func allResponses(ms []provider.Member) bool {
	for _, m := range ms {
		if m.Provider.Native(m.Model) != provider.Responses {
			return false
		}
	}
	return len(ms) > 0
}

// maxTokens is the output limit an agent is handed for m, kept within the
// context window it is handed with it: models.dev lists some models' output
// above their window (deepseek-chat's 384000 against 128000). An unknown
// window leaves the output as it is.
func maxTokens(m catalog.Model) int { return provider.OutputWithin(m.Context, m.Output) }

// group tags every option with a group name.
func group(name string, opts []Option) []Option {
	for i := range opts {
		opts[i].Group = name
	}
	return opts
}
