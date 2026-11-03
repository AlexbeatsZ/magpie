package agent

// Hermes Agent (Nous Research) keeps its settings in $HERMES_HOME/config.yaml,
// by default %LOCALAPPDATA%\hermes\config.yaml on Windows and
// ~/.hermes/config.yaml elsewhere (place.hermesHome): the model under model.default, served by
// model.provider, which may name an entry of providers. magpie adds itself
// there as the provider "magpie" (the gateway, spoken to as chat completions,
// with the catalog as its models) and points model.provider at it; a model
// through magpie is "magpie/<provider>/<model>" here and model.default holds
// "<provider>/<model>". The provider and model the user had are stashed and
// put back when magpie steps out. A model of the user's own picked after
// that, in Hermes or on the Agents page, leaves magpie's provider in
// providers: Hermes stays connected, its session model picker still offers
// magpie's models beside its own (lijiho96940561 on X: Hermes went to Not
// connected and magpie's models left its picker), and only Disconnect
// takes the provider out.
//
// Each of Hermes's bots is a profile with its own config.yaml (see
// hermesProfiles), which doesn't read the main one's: magpie's provider goes
// into every profile's too, so magpie is offered in every bot's picker, not
// the main one's alone (lijiho96940561 on X). A profile's own model and
// provider are left as its user set them, and Disconnect takes the provider
// out of each. A profile still on the model magpie last set in the
// connected config (one Hermes seeded or cloned from it) follows the next
// one magpie sets there (hermesFollow; tjhhhh_84378 on Discord: another
// bot kept the old model). The config magpie connects is the profile
// `hermes` starts in (hermesActive), and Disconnect leaves each bot on a
// model it can start with, which connecting again puts back (hermesBack).
// With more than one profile, each other one has a model picker of its own
// on the Agents page (hermesProfileFields; dongshanfengdongshizhcn on
// Discord), and a bot picked there no longer follows the connected one.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func hermes(home string) *Agent { return hermesIn(here(home)) }

// hermesIn is Hermes at a place: this machine's home, or a WSL distro's
// (see wsl.go), where HERMES_HOME isn't read and its provider names the
// gateway as the distro reaches it.
func hermesIn(at place) *Agent {
	dir := hermesActive(at)
	path := filepath.Join(dir, "config.yaml")

	a := &Agent{
		ID: "hermes", Name: "Hermes Agent", Icon: "hermes", Aliases: []string{"hermes-agent"},
		UA:  []string{"hermes-agent"},
		Bin: "hermes", Dir: dir, Path: path,
	}

	return a
}

// hermesProfileID is the name a Hermes profile may have
// (hermes_constants.PROFILE_ID_RE).
var hermesProfileID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// hermesRoot is the folder Hermes's profiles are under
// (get_default_hermes_root): its default folder, which a HERMES_HOME under
// it is a profile of; a HERMES_HOME elsewhere is the root itself, or, as
// <root>/profiles/<name>, a profile of <root>.
func hermesRoot(at place) string {
	root := at.hermesDefault()
	if env := at.getenv("HERMES_HOME"); env != "" {
		if rel, err := filepath.Rel(root, env); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			root = env
			if filepath.Base(filepath.Dir(env)) == "profiles" {
				root = filepath.Dir(filepath.Dir(env))
			}
		}
	}
	return root
}

// hermesLive is whether name is a profile of root Hermes runs: a name it
// allows, not "default" (the root's own), not deleted (a tombstone at
// profiles/.deleted/<name>)
func hermesLive(root, name string) bool {
	if name == "default" || !hermesProfileID.MatchString(name) {
		return false
	}
	_, err := os.Stat(filepath.Join(root, "profiles", ".deleted", name))
	return err != nil
}

// hermesActive is the folder of the profile `hermes` starts in, whose
// config.yaml magpie connects: HERMES_HOME when it names a profile
// (<root>/profiles/<name>); else the profile `hermes profile use <name>`
// made sticky in <root>/active_profile, when it is there; else HERMES_HOME
// or Hermes's default folder (hermes_cli/main.py _apply_profile_override).
// `hermes model` writes the same profile's config. With a sticky profile,
// magpie wrote the root's config, which `hermes` doesn't read, so a model
// picked in magpie didn't reach the bot the user runs (tjhhhh on Discord).
func hermesActive(at place) string {
	home := at.hermesHome()
	if at.getenv("HERMES_HOME") != "" && filepath.Base(filepath.Dir(home)) == "profiles" {
		return home
	}
	// a stopped distro's file is never opened: that starts it (#3896)
	if at.cold {
		return home
	}
	root := hermesRoot(at)
	b, err := os.ReadFile(filepath.Join(root, "active_profile"))
	if err != nil {
		return home
	}
	name := strings.TrimSpace(strings.TrimPrefix(string(b), "\ufeff"))
	dir := filepath.Join(root, "profiles", name)
	if !hermesLive(root, name) {
		return home
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return home
	}
	return dir
}
