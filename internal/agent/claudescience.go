package agent

// Claude Science is Anthropic's research app (0.1.62, "operon" inside): a
// daemon, claude-science, that serves its UI on localhost, and a window
// that shows it (a Swift launcher on the Mac, Electron on Windows, a
// browser on Linux) and attaches to the daemon running, else starts one.
// The daemon asks Anthropic's Messages API at its environment's
// ANTHROPIC_BASE_URL (ky() in its bundle: https, or http on a loopback
// address) and has no setting for it, so magpie connects it by starting
// the daemon itself with that variable set to the gateway — `stop`, then
// `serve --detached --no-browser`, on the port it was on — and nothing
// else: no variable set for the user, no file of its written. Disconnect
// restarts it without the variable. The app still needs its own Claude
// sign-in, which it sends as its bearer and the gateway never passes on
// (see gateway/science.go); with none it asks nothing at all.
//
// The daemon lives till it is stopped or the computer restarts, so while
// it is connected magpie starts it again when it isn't running
// (KeepClaudeScience), and says so when another one runs in its place —
// the window, opened while none ran, starts its own without the variable
// (Check).

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/proc"
)

// ScienceID is the agent's id.
const ScienceID = gateway.ScienceID

const scienceName = "Claude Science"

// scienceBin is the daemon's program, "" when it isn't installed. A var so
// tests point it elsewhere.
var scienceBin = func() string {
	scienceSeen.Lock()
	defer scienceSeen.Unlock()
	if time.Since(scienceSeen.at) > 30*time.Second {
		home, _ := os.UserHomeDir()
		scienceSeen.bin, scienceSeen.at = findScience(home), time.Now()
	}
	return scienceSeen.bin
}

var scienceSeen struct {
	sync.Mutex
	bin string
	at  time.Time
}

// findScience is where claude-science is installed for home: in the Mac
// app's bundle, Windows' per-user Programs folder, or the ~/.local/bin the
// installers link it into (the only place on Linux), else on PATH.
func findScience(home string) string {
	var at []string
	switch runtime.GOOS {
	case "darwin":
		for _, d := range macAppFolders(home) {
			at = append(at, filepath.Join(d, scienceName+".app", "Contents", "Resources", "bin", "claude-science"))
		}
	case "windows":
		if d := os.Getenv("LOCALAPPDATA"); d != "" {
			at = append(at, filepath.Join(d, "Programs", "ClaudeScience", "claude-science.exe"))
		}
		at = append(at, filepath.Join(home, "AppData", "Local", "Programs", "ClaudeScience", "claude-science.exe"))
	}
	exe := "claude-science"
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	at = append(at, filepath.Join(home, ".local", "bin", exe))
	for _, p := range at {
		if isFile(p) {
			return p
		}
	}
	if p, err := proc.LookPath("claude-science"); err == nil {
		return p
	}
	return ""
}

func claudeScience() *Agent {
	return &Agent{
		ID: ScienceID, Name: scienceName, Icon: "claude-color", Aliases: []string{"operon", "science"},

		detect: func() bool { return scienceBin() != "" },
	}
}
