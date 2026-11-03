package agent

// Qoder's CLI (1.1) keeps its settings in $QODER_CONFIG_DIR/settings.json,
// ~/.qoder by default. A provider of one's own is an entry of "providers",
// its models picked as "<provider>/<model>":
//
//	{"providers":{"magpie":{"displayName":"magpie","protocol":"openai",
//	   "baseUrl":"http://127.0.0.1:3425/v1","apiKey":"magpie","model":…,
//	   "models":[{"model":…,"displayName":…,"contextWindow":…,"maxOutputTokens":…,
//	     "capabilities":{"tools":true,"vision":…,"thinking":{"modes":["enabled"],
//	       "supportsEffort":true,"supportedEffortLevels":[…]}}}]}},
//	 "model":{"name":"magpie/<model>","reasoningEffort":…,
//	   "preferences":{"magpie/<model>":{"reasoning":{"effort":…}}}}}
//
// Each model's effort is its own, in model.preferences (1.1.6x); the one in
// model.reasoningEffort is asked for only by a model with none there, and
// Qoder moves it into the preferences of the model it starts on, and drops
// it, so it counts once. magpie sets the model's preference, and keeps
// reasoningEffort for the Qoders before preferences.
//
// An openai provider is asked at baseUrl + /chat/completions, with
// reasoning_effort when the model says it takes one. Qoder offers custom
// providers only to a signed-in account whose plan has BYOK. Its requests
// say only undici, so they are not told apart from other clients'.
//
// Qoder CN (qoderclicn, @qodercn-ai/qoderclicn) is the same CLI for the
// China site, its own accounts and settings in $QODERCN_CONFIG_DIR,
// ~/.qoder-cn by default.

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func qoder(home string) *Agent {
	return qoderSite(home, qoderGlobal)
}

func qoderCN(home string) *Agent {
	return qoderSite(home, qoderChina)
}

// qoderBuild is what tells Qoder's two builds apart.
type qoderBuild struct {
	id, name, env, dir, bin string
	aliases, procs          []string
}

var (
	qoderGlobal = qoderBuild{id: "qoder", name: "Qoder", env: "QODER_CONFIG_DIR", dir: ".qoder", bin: "qodercli",
		aliases: []string{"qodercli", "qoder-cli"}, procs: []string{`(^|/)qodercli( |$)`, `(^|/)qoder( |$)`}}
	qoderChina = qoderBuild{id: "qoder-cn", name: "Qoder CN", env: "QODERCN_CONFIG_DIR", dir: ".qoder-cn", bin: "qoderclicn",
		aliases: []string{"qoderclicn", "qodercn", "qodercn-cli"}, procs: []string{`(^|/)qoderclicn( |$)`, `(^|/)qodercn( |$)`}}
)

// qoderIn and qoderCNIn are Qoder's two builds in a WSL distro (see
// wsl.go): at their default folders, as the distro's variables aren't read.
func qoderIn(at place) *Agent   { return qoderAt(at, qoderGlobal) }
func qoderCNIn(at place) *Agent { return qoderAt(at, qoderChina) }

func qoderSite(home string, b qoderBuild) *Agent { return qoderAt(here(home), b) }

// qoderAt is a build of Qoder at a place, its provider naming the gateway
// as it reaches it from there.
func qoderAt(at place, b qoderBuild) *Agent {
	dir := at.getenv(b.env)
	if dir == "" {
		dir = at.join(b.dir)
	}
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: b.id, Name: b.name, Icon: "qoder", Aliases: b.aliases,
		Bin: b.bin, Dir: dir, Path: path,
		// Qoder's JetBrains plugin keeps its indexer, app/ and plugins/ in
		// the same folder, and writes a settings.json of enabledPlugins
		// alone, with no CLI installed (#1495): neither is a sign of the
		// CLI. Its command, or a settings.json with any other key, is.
		// The Qoder app (Qoder-win-x64-user.exe, Qoder.app, not the IDE)
		// runs the CLI as its agent with the same folder, so magpie's
		// provider reaches it too; its settings.json can hold only
		// enabledPlugins (its bundled qoder-context plugin), and its status
		// file says it is there (#2138).
		dirShared: true,
		pathOwn:   func() bool { return !qoderPluginOnly(path) },
		sign:      func() bool { return qoderApp(dir) },
	}
}

// qoderApp reports whether the Qoder app keeps its status in dir: it
// writes .qoder-app-status.json (-canary-, -dev- for its other channels)
// into ~/.qoder, or ~/.qoder-cn for the China build, as it signs in, and
// runs Qoder's CLI as its agent with that folder's settings.
func qoderApp(dir string) bool {
	for _, f := range []string{".qoder-app-status.json", ".qoder-app-canary-status.json", ".qoder-app-dev-status.json"} {
		if fi, err := os.Stat(filepath.Join(dir, f)); err == nil && fi.Mode().IsRegular() {
			return true
		}
	}
	return false
}

// qoderPluginOnly reports whether the settings file at path is the one
// Qoder's JetBrains plugin writes, an object holding enabledPlugins and
// nothing else. A file that can't be read or parsed isn't known to be
// that, so it says false and the file still counts as the CLI's.
func qoderPluginOnly(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(b, &m) != nil || len(m) == 0 {
		return false
	}
	for k := range m {
		if k != "enabledPlugins" {
			return false
		}
	}
	return true
}
