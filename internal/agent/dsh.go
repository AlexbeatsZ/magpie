package agent

// DeepSeek Harness (dsh) boots from its shipped rows and then applies a
// personal patch list: a YAML list of {id, config} entries, each replacing
// the whole config of the row with that id. Since 0.1.5 every profile has
// its own, ~/.dsh/profiles/<name>/cordis.patch.yml (web and desktop read it
// live); before that there was one, ~/.dsh/config.yaml.
//
// Since 0.1.5 magpie is one of dsh's custom model providers: a route named
// magpie in the llm-pi-ai entry's providers (what dsh's Models page writes
// for "Add model provider"), with the gateway as its endpoint, the catalog
// as its models and, as the key, a credential named in apiKeyEnv, which
// magpie puts in ~/.dsh/.env and dsh's own key store (the desktop app reads
// only the store); agent-default-model names the model new
// sessions start on. dsh's own DeepSeek row (llm-deepseek) is left to dsh:
// taken over, it read as DeepSeek in dsh's Models page, and a DeepSeek key
// entered there went to the gateway under magpie's credential name, over
// magpie's own (Discord: 集成进 dsh 会被 dsh 覆写). A route of the user's in
// the same entry stays as it is, and so does one dsh adds beside magpie's.
//
// Before 0.1.5, which has no custom providers, magpie took over
// llm-deepseek, with the key in the entry, and set agent-loop's main agent
// (the TUI) and api-gateway's route (`dsh -p`, `dsh web`). Entries magpie
// writes whole are marked as magpie's, the user's own entries stay as they
// are, and one magpie replaces is stashed and put back when it steps out.

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/yetone/magpie/internal/edit"
)

const dshMark = "# magpie"

func dsh(home string) *Agent { return dshAt(here(home)) }

// dshIn is DeepSeek Harness in a WSL distro (see wsl.go).
func dshIn(at place) *Agent { return dshAt(at) }

// dshAt is DeepSeek Harness with its home at at.home, reaching the gateway
// as at does. A distro's DSH_HOME can't be read, so there it is ~/.dsh.
func dshAt(at place) *Agent {

	dir := at.getenv("DSH_HOME")
	if dir == "" {
		dir = at.join(".dsh")
	}
	// DeepSeek Harness Desktop, on this machine (a distro has none)
	desktop := func() bool { return at.spell == nil && (&Agent{ID: "dsh"}).HasApp() }
	// the patch list magpie makes for a dsh with no profile yet: the desktop
	// app's, else the web profile's; "" for a dsh before profiles, which
	// reads config.yaml (liby233: a dsh 0.2 in WSL that hadn't booted a
	// profile yet was wired in config.yaml, which it never reads)
	fresh := func() string {
		if desktop() {
			return dshDesktopFile(dir)
		}
		version := at.version
		if at.spell == nil {
			version = dshVersion()
		}
		if dshBeforeProfiles(dir, version, at.cold) {
			return ""
		}
		return dshWebFile(dir)
	}

	path := filepath.Join(dir, "config.yaml")
	// a stopped distro's folder is never listed: that starts it (#3896)
	var files []string
	if !at.cold {
		files = dshProfiles(dir)
	}
	if len(files) > 0 {
		path = files[0]
	} else if f := fresh(); f != "" {
		path = f
	}
	return &Agent{
		ID: "dsh", Name: "DeepSeek Harness", Icon: "deepseek-color", Aliases: []string{"deepseek-harness"},
		UA:  []string{"deepseek-harness"},
		Bin: "dsh", Dir: dir, Path: path,
		// the desktop app installed and not opened yet has no ~/.dsh (#2279)
		sign: desktop,
	}
}

// dshProfiles are the profiles' patch lists (dsh 0.1.5 on), web's first.
func dshProfiles(dir string) []string {
	files, _ := filepath.Glob(filepath.Join(dir, "profiles", "*", "cordis.patch.yml"))
	var web, rest []string
	for _, f := range files {
		if filepath.Base(filepath.Dir(f)) == "web" {
			web = append(web, f)
		} else {
			rest = append(rest, f)
		}
	}
	return append(web, rest...)
}

// dshDesktopFile is the desktop app's profile's patch list: the app's
// profile is $DSH_HOME/profiles/desktop (resolveDesktopPaths).
func dshDesktopFile(dir string) string {
	return filepath.Join(dir, "profiles", "desktop", "cordis.patch.yml")
}

// dshWebFile is the web profile's patch list, $DSH_HOME/profiles/web: the
// profile `dsh web` boots, the first of dsh's shipped ones (web, headless,
// acp) that magpie makes for a dsh CLI that has booted none yet. A profile
// dsh makes later is given magpie's route by the sync (dshFillNewProfiles).
func dshWebFile(dir string) string {
	return filepath.Join(dir, "profiles", "web", "cordis.patch.yml")
}

// dshProfilesSince is the first dsh whose patch lists are its profiles'.
// That is every dsh on npm: 0.0.1-rc.1, the first, already boots
// $DSH_HOME/profiles/<name>, and none of 0.0.1-rc.1, 0.0.1-rc.5,
// 0.1.0-rc.8, 0.1.1-rc.1/rc.2, 0.1.2-rc.1, 0.1.3-alpha.2 or 0.2.1-rc.5
// reads a config.yaml. A dsh saying an older version is one from before.
const dshProfilesSince = "0.0.1-rc.1"

// dshVersion is the version of the dsh CLI on this machine, "" when there is
// none or it didn't say; a var for tests.
var dshVersion = func() string { return (&Agent{ID: "dsh", Bin: "dsh"}).InstalledVersion() }

// dshBeforeProfiles says a dsh with no profile at dir is one from before
// profiles, which reads ~/.dsh/config.yaml: by its version where that is
// known, else only where config.yaml holds an entry that isn't magpie's
// (what such a dsh, or its user, wrote there). A dsh since profiles makes
// none until one is first booted, and reads no config.yaml at all. cold is
// a stopped distro's, whose files are not read.
func dshBeforeProfiles(dir, version string, cold bool) bool {
	if version != "" {
		return Newer(dshProfilesSince, version)
	}
	if cold {
		return false
	}
	_, items, err := dshRead(filepath.Join(dir, "config.yaml"))
	return err == nil && slices.ContainsFunc(items, func(it dshItem) bool { return !it.magpie })
}

// dshItem is one entry of the patch list, as its lines.
type dshItem struct {
	id     string
	magpie bool
	lines  []string
}

var dshIDLine = regexp.MustCompile(`^(?:- |  )id:\s*['"]?([^'"#\s]+)['"]?\s*(#.*)?$`)

// dshParse splits the patch list at path into what comes before its first
// entry and the entries. A list written in flow style ([ {id: a} ], which
// dsh's own writers keep a profile's [] in) is read as the same list in block
// style, and written back so; a file that is no list is left alone.
func dshParse(path, raw string) (head []string, items []dshItem, err error) {
	if b, ok := edit.BlockList(raw); ok {
		raw = b
	}
	opened := false // a "---" or "[]" came before the first entry
	for _, l := range splitLinesKeep(raw) {
		// magpie's notes are left out, and dshWrite puts them back: a
		// route is compared with the lines magpie would write
		if edit.IsNote(l) {
			continue
		}
		t := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "- ") || l == "-":
			items = append(items, dshItem{lines: []string{l}})
		case len(items) == 0:
			// "---" is YAML's document start marker: a patch list may open
			// with one (#2667), and it stays in the head it is written back
			// with, so the file keeps it. Only at the left edge and before
			// anything but comments: a second one starts a second
			// document, which is no list magpie can edit.
			marker := dshDocStart(l) && !opened
			if t != "" && !strings.HasPrefix(t, "#") && t != "[]" && !marker {
				return nil, nil, fmt.Errorf("%s is not a list of entries magpie can edit", path)
			}
			if t != "" && !strings.HasPrefix(t, "#") {
				opened = true
			}
			if t != "[]" {
				head = append(head, l)
			}
			continue
		case t != "" && !strings.HasPrefix(t, "#") && !strings.HasPrefix(l, " "):
			return nil, nil, fmt.Errorf("%s is not a list of entries magpie can edit", path)
		default:
			items[len(items)-1].lines = append(items[len(items)-1].lines, l)
		}
		it := &items[len(items)-1]
		if m := dshIDLine.FindStringSubmatch(l); m != nil && it.id == "" {
			it.id = m[1]
			it.magpie = strings.TrimSpace(m[2]) == dshMark
		}
	}
	return head, items, nil
}

func splitLinesKeep(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func dshRead(path string) ([]string, []dshItem, error) {
	b, err := edit.Read(path)
	if err != nil {
		return nil, nil, err
	}
	return dshParse(path, string(b))
}

// dshProfileList reports whether f is a profile's own patch list rather than
// the home layer, which applies over every profile's.
func dshProfileList(f string) bool {
	return filepath.Base(filepath.Dir(filepath.Dir(f))) == "profiles"
}

// dirOf is the dsh home a patch list belongs to.
func dirOf(f string) string {
	if dshProfileList(f) {
		return filepath.Dir(filepath.Dir(filepath.Dir(f)))
	}
	return filepath.Dir(f)
}

// dshSaidOnce is what a round could not write, as it was last said: a round
// repeating it says nothing, so a profile magpie cannot write to is reported
// once rather than every 30s.
type dshSaidOnce struct{ last string }

// first reports whether this trouble has not been said yet, and is to be said.
func (s *dshSaidOnce) first(trouble string) bool {
	if trouble == s.last {
		return false
	}
	s.last = trouble
	return trouble != ""
}

// dshFiles are the patch lists magpie writes: every profile's, else
// config.yaml.
func dshFiles(dir string) []string {
	if files := dshProfiles(dir); len(files) > 0 {
		return files
	}
	return []string{filepath.Join(dir, "config.yaml")}
}

// DshPatchFiles are the patch lists of the dsh whose home is dir, as magpie
// writes them: every profile's (web's first), else config.yaml.
func DshPatchFiles(dir string) []string { return dshFiles(dir) }

// dshDocStart says whether l is YAML's document start marker, "---" at the
// left edge with at most a comment after it.
func dshDocStart(l string) bool {
	r, ok := strings.CutPrefix(l, "---")
	if !ok {
		return false
	}
	if r == "" {
		return true
	}
	if r[0] != ' ' && r[0] != '\t' {
		return false
	}
	r = strings.TrimSpace(r)
	return r == "" || strings.HasPrefix(r, "#")
}
