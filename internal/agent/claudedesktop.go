package agent

// Claude Desktop runs on a third-party gateway in its "3p" deployment mode:
// no Anthropic sign-in, and its Code and Cowork tabs work on any plan (a
// Free account signed in sees only "Get Claude Code" there). The mode is
// claude_desktop_config.json's deploymentMode, in its own folder and in
// Claude-3p beside it; the gateway is a profile in Claude-3p/configLibrary,
// the one _meta.json applies. magpie writes "3p" in both, but Desktop reads
// only Claude-3p's, and not even that one when the profile disables its
// chooser (desktop3p):
//
//	Claude/claude_desktop_config.json      {"deploymentMode":"3p",…}
//	Claude-3p/claude_desktop_config.json   {"deploymentMode":"3p",…}
//	Claude-3p/configLibrary/<id>.json      {"inferenceProvider":"gateway",
//	  "inferenceGatewayBaseUrl":"http://127.0.0.1:3425",
//	  "inferenceGatewayApiKey":"magpie-claude-desktop",
//	  "inferenceGatewayAuthScheme":"bearer",
//	  "disableDeploymentModeChooser":true,"coworkEgressAllowedHosts":["*"]}
//	Claude-3p/configLibrary/_meta.json     {"appliedId":"<id>",
//	  "entries":[…,{"id":"<id>","name":"magpie"}]}
//
// as CC Switch writes them (claude_desktop_config.rs). The folders are in
// ~/Library/Application Support on macOS, %LOCALAPPDATA% on Windows (the
// MSIX package's LocalCache\Local for a packaged Desktop) and
// $XDG_CONFIG_HOME on Linux, as internal/desktopdir finds them. A
// packaged Desktop gets all of it twice, in the package's folders and the
// real ones: up to 2.x it reads the package's copy, from 2.31226 the real
// %LOCALAPPDATA%\Claude-3p (desktopdir.Sets). With no inferenceModels Desktop lists the
// gateway's /v1/models, which the gateway gives it by ids it keeps (see
// gateway/desktop.go). Desktop reads all this at start-up only.

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yetone/magpie/internal/desktopdir"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/gateway"
)

// desktopProfileID is magpie's profile in Desktop's configLibrary: a
// UUID like the ones Desktop makes, its last part "magpie" in hex.
const desktopProfileID = "00000000-0000-4000-8000-6d6167706965"

// desktopPaths are the files and folders magpie writes for Claude Desktop.
type desktopPaths struct {
	key                 string // the stash's keys end with it: "" or ".package"
	dir, dir3p          string // Claude, Claude-3p
	config, config3p    string // their claude_desktop_config.json
	library, meta, prof string // configLibrary, its _meta.json, magpie's profile
	// Desktop's own settings (config.json, its "locale" among them) in its
	// userData signed in with Anthropic (desktopdir's Data: Claude, or
	// %APPDATA%\Claude on Windows) and in Claude-3p, its userData in 3p
	// mode. magpie only ever adds the language to the second.
	app, app3p string
}

func desktopPathsOf(dir, dir3p string) desktopPaths {
	lib := filepath.Join(dir3p, "configLibrary")
	return desktopPaths{
		dir: dir, dir3p: dir3p,
		config:   filepath.Join(dir, "claude_desktop_config.json"),
		config3p: filepath.Join(dir3p, "claude_desktop_config.json"),
		library:  lib, meta: filepath.Join(lib, "_meta.json"), prof: filepath.Join(lib, desktopProfileID+".json"),
		app: filepath.Join(dir, "config.json"), app3p: filepath.Join(dir3p, "config.json"),
	}
}

// desktopSets are those of every set of Desktop's folders, Find's first:
// for a packaged Desktop the package's and the real ones, each with its
// own stash.
func desktopSets(goos, home string, getenv func(string) string) []desktopPaths {
	var out []desktopPaths
	for _, d := range desktopdir.Sets(goos, home, getenv) {
		p := desktopPathsOf(d.Mode, d.ThreeP)
		p.app = filepath.Join(d.Data, "config.json")
		if d.Package != "" {
			p.key = ".package"
		}
		out = append(out, p)
	}
	return out
}

// DesktopConfigs3p are the claude_desktop_config.json files Claude Desktop
// reads in its 3p mode: there its whole userData is Claude-3p, its MCP
// servers too (%LOCALAPPDATA%\Claude-3p on Windows, and the MSIX package's;
// Claude-3p beside Claude elsewhere).
func DesktopConfigs3p(home string) []string {
	var out []string
	for _, p := range desktopSets(desktopdir.OS, home, os.Getenv) {
		out = append(out, p.config3p)
	}
	return out
}

func claudeDesktop(home string) *Agent {
	sets := desktopdir.Sets(desktopdir.OS, home, os.Getenv)
	ps := desktopSets(desktopdir.OS, home, os.Getenv)
	p := ps[0]
	return &Agent{
		ID: "claude-desktop", Name: "Claude Desktop", Icon: "claude-color", Aliases: []string{"claude-app"},
		Dir: p.dir, Path: p.config,
		detect: func() bool {
			// its own data (%APPDATA%\Claude on Windows, or the MSIX
			// package's), the folder magpie writes or Claude-3p
			for _, s := range sets {
				for _, d := range s.All() {
					if isDir(d) {
						return true
					}
				}
			}
			// installed on Windows and not yet opened: none of them yet
			return desktopdir.Installed(desktopdir.OS, home, os.Getenv)
		},
	}
}

// desktop3p: Desktop starts in its 3p mode with prof, the gateway profile
// it applies, as Desktop decides it (m4 in 4.4212's and 2.7032's
// index.pre.js): with the profile's disableDeploymentModeChooser, whatever
// deploymentMode says, and otherwise unless Claude-3p's
// claude_desktop_config.json holds "1p" — none is 3p. The Claude folder's
// deploymentMode isn't read, and Desktop deletes Claude-3p's itself: a
// stored "1p" with no claude.ai sign-in is dropped at start and Desktop
// relaunches in 3p ("deploymentMode written { mode: undefined }"), so
// requiring "3p" in both files called a Desktop on magpie off it (dev_quj
// on X: "已不经过 magpie：配置已变更" after a restart).
func desktop3p(p desktopPaths, prof string) bool {
	if c, _ := edit.GetJSON(prof, "disableDeploymentModeChooser"); c == "true" {
		return true
	}
	m, _ := edit.GetJSON(p.config3p, "deploymentMode")
	return m != "1p"
}

// desktopAppliedGateway is the configLibrary profile Desktop applies when
// it is not magpie's own but a gateway profile for magpie's gateway, ""
// for none. Desktop's "Import configuration" makes one: it adds the file
// it is given as a new entry under an id of its own (configLibrary's
// createConfig, a random UUID) and writes the file's settings into it, so
// magpie's profile exported on one computer and imported on another is
// applied there under another id, with the connectors (managedMcpServers)
// and settings the user made in it (aa86222 on Discord: 配置已变更 while
// Desktop ran on magpie). It is the user's profile: magpie keeps it
// applied and sets nothing in it but a key the gateway refuses.
func desktopAppliedGateway(p desktopPaths) string {
	id, _ := edit.GetJSON(p.meta, "appliedId")
	if id == "" || id == desktopProfileID || filepath.Base(id) != id || strings.HasPrefix(id, ".") {
		return ""
	}
	f := filepath.Join(p.library, id+".json")
	b, err := os.ReadFile(f)
	if err != nil || !desktopObjectBytes(b) {
		return ""
	}
	var prof struct {
		Provider string `json:"inferenceProvider"`
		BaseURL  string `json:"inferenceGatewayBaseUrl"`
	}
	if json.Unmarshal(b, &prof) != nil || prof.Provider != "gateway" || !desktopIsOurGateway(prof.BaseURL) {
		return ""
	}
	return f
}

// desktopIsOurGateway: u is magpie's gateway, by its port on this machine
// (127.0.0.1, localhost or ::1, with or without /v1).
func desktopIsOurGateway(u string) bool {
	a, err := url.Parse(strings.TrimSpace(u))
	if err != nil || a.Port() == "" {
		return false
	}
	ours, err := url.Parse(gateway.URL())
	if err != nil || a.Port() != ours.Port() {
		return false
	}
	h := a.Hostname()
	if h == ours.Hostname() || h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// desktopObjectBytes: b is a JSON object.
func desktopObjectBytes(b []byte) bool {
	var m map[string]json.RawMessage
	return json.Unmarshal(b, &m) == nil && m != nil
}
