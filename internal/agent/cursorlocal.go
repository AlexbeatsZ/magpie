package agent

// Cursor Private Inference is a build of Cursor whose agent runs on this
// machine (its cursor-local-agent-runtime extension) and asks a model
// endpoint of the user's rather than Cursor's backend (#299). It is
// installed as Cursor is — the same app name, bundle id and ~/.cursor —
// and tells itself apart only by its product.json ("nameShort": "Cursor
// Private Inference"). The endpoint it takes from a model's own settings,
// else its Open configuration dialog's (kept in the state.vscdb regular
// Cursor shares), else from its environment: CURSOR_LOCAL_AGENT_BASE_URL
// and CURSOR_LOCAL_AGENT_API_KEY. magpie writes nothing of Cursor's — the
// Open configuration's base URL and key are regular Cursor's own override
// settings too, so writing them would move regular Cursor as well. It sets
// the two variables, which only this build reads, for the user instead
// (cursorLocalUserEnv: launchctl on the Mac, the registry's Environment on
// Windows, environment.d on Linux), so an app opened from the Dock or the
// Start menu has them, and the row is connected like any other agent's;
// the command that starts it with them (Agent.Launch) stays for a start
// from a shell. The gateway's /models tells it, for each model, the APIs
// it is served on (api_types) and its limits (capabilities), which it
// reads to pick Anthropic Messages, Responses or Chat and its context.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/proc"
)

// CursorLocalID is the agent's id, which its key names (TokenFor).
const CursorLocalID = "cursor-local"

// cursorLocalName is product.json's nameShort for the build.
const cursorLocalName = "Cursor Private Inference"

// cursorLocalApp is where the build is installed: its program, to start,
// and "" when it isn't. A var so tests can point it elsewhere.
var cursorLocalApp = func() string {
	cursorLocalSeen.Lock()
	defer cursorLocalSeen.Unlock()
	if time.Since(cursorLocalSeen.at) > 30*time.Second {
		cursorLocalSeen.app, cursorLocalSeen.at = findCursorLocal(cursorLocalRoots(), cursorLocalInstalls()...), time.Now()
	}
	return cursorLocalSeen.app
}

// cursorLocalSeen is where the build was last found, looked for again
// after half a minute: the apps folder is read every time the agents are.
var cursorLocalSeen struct {
	sync.Mutex
	app string
	at  time.Time
}

// cursorLocalRoots are the folders apps are installed in on this system.
// On Linux the build comes as an AppImage (Cursor_Private_Inference-<v>-
// x86_64.AppImage), kept wherever it was downloaded, so the folders it is
// usually kept in are looked in too.
func cursorLocalRoots() []string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return macAppFolders(home)
	case "windows":
		// LOCALAPPDATA is the user's (a test's lies under its home);
		// ProgramFiles is the machine's, so it goes through SystemDirs
		var out []string
		for _, d := range append([]string{os.Getenv("LOCALAPPDATA")}, proc.SystemDirs(os.Getenv("ProgramFiles"))...) {
			if d != "" {
				out = append(out, filepath.Join(d, "Programs"), d)
			}
		}
		return out
	}
	return append(proc.SystemDirs("/opt", "/usr/share", "/usr/lib"), filepath.Join(home, ".local", "share"), filepath.Join(home, "Applications"),
		filepath.Join(home, ".local", "bin"), filepath.Join(home, "bin"), appdir.Downloads(), filepath.Join(home, "Downloads"), home)
}

// findCursorLocal is the program of the Cursor Private Inference among the
// apps in roots, else the first of installs (folders it was installed in,
// as Windows' list of installed programs gives them, which may be any
// folder the user picked in its installer).
func findCursorLocal(roots []string, installs ...string) string {
	for _, root := range roots {
		ents, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if exe := cursorLocalAt(filepath.Join(root, e.Name())); exe != "" {
				return exe
			}
		}
	}
	for _, dir := range installs {
		if exe := cursorLocalAt(dir); exe != "" {
			return exe
		}
	}
	return ""
}

// cursorLocalAt is the program of the build installed at dir, "" when dir
// isn't it: one whose product.json names it, a Mac's bundle by its
// Info.plist's CFBundleExecutable, else by the applicationName product.json
// gives (cursor, Cursor.exe); or its AppImage on Linux, by the name it is
// downloaded under.
func cursorLocalAt(dir string) string {
	name := filepath.Base(dir)
	if strings.HasSuffix(name, ".app") {
		res := filepath.Join(dir, "Contents", "Resources", "app")
		if p, ok := cursorLocalProduct(res); ok {
			if exe := plistExecutable(filepath.Join(dir, "Contents", "Info.plist")); exe != "" {
				return filepath.Join(dir, "Contents", "MacOS", exe)
			}
			return filepath.Join(dir, "Contents", "MacOS", p.macExe())
		}
		return ""
	}
	if cursorLocalAppImage(name) && isFile(dir) {
		return dir
	}
	if p, ok := cursorLocalProduct(filepath.Join(dir, "resources", "app")); ok {
		exe := p.exe()
		if runtime.GOOS == "windows" {
			exe = p.NameShort + ".exe"
			if !isFile(filepath.Join(dir, exe)) {
				exe = p.exe() + ".exe"
			}
		}
		return filepath.Join(dir, exe)
	}
	return ""
}

// cursorLocalAppImage says a file's name is the build's AppImage as it is
// downloaded (Cursor_Private_Inference-3.24.9-x86_64.AppImage), which
// regular Cursor's (Cursor-3.24.9-x86_64.AppImage) isn't.
func cursorLocalAppImage(name string) bool {
	n := strings.ToLower(name)
	if !strings.HasSuffix(n, ".appimage") {
		return false
	}
	return strings.HasPrefix(strings.NewReplacer("_", "", "-", "", " ", "").Replace(n), "cursorprivateinference")
}

type cursorProduct struct {
	NameShort       string `json:"nameShort"`
	ApplicationName string `json:"applicationName"`
}

// exe is the program's name as product.json gives it: Cursor for cursor.
func (p cursorProduct) exe() string {
	n := p.ApplicationName
	if n == "" {
		return "Cursor"
	}
	if runtime.GOOS == "linux" {
		return n
	}
	return p.macExe()
}

// macExe is the program's name in a .app, which is a Mac's on any system.
func (p cursorProduct) macExe() string {
	n := p.ApplicationName
	if n == "" {
		return "Cursor"
	}
	return strings.ToUpper(n[:1]) + n[1:]
}

// cursorLocalProduct reads the product.json in an app's resources, which
// is the build's when it names it.
func cursorLocalProduct(res string) (cursorProduct, bool) {
	var p cursorProduct
	b, err := os.ReadFile(filepath.Join(res, "product.json"))
	if err != nil || json.Unmarshal(b, &p) != nil {
		return p, false
	}
	return p, strings.EqualFold(strings.TrimSpace(p.NameShort), cursorLocalName)
}

var plistExeRe = regexp.MustCompile(`<key>CFBundleExecutable</key>\s*<string>([^<]+)</string>`)

// plistExecutable is CFBundleExecutable of an XML Info.plist, "" when it
// can't be read.
func plistExecutable(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if m := plistExeRe.FindSubmatch(b); m != nil {
		return strings.TrimSpace(string(m[1]))
	}
	return ""
}

// cursorLocalVars are the variables the build reads its endpoint from.
var cursorLocalVars = []string{"CURSOR_LOCAL_AGENT_BASE_URL", "CURSOR_LOCAL_AGENT_API_KEY"}

func cursorLocal() *Agent {
	return &Agent{
		ID: CursorLocalID, Name: cursorLocalName, Icon: "cursor", Aliases: []string{"cursor-private-inference"},

		detect: func() bool { return cursorLocalApp() != "" },
	}
}
