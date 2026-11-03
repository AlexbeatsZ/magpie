package uninstall

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/library"
)

// fakes puts a fake in for everything Run asks of the system, and records
// what it was asked, in order.
func fakes(t *testing.T, list []*agent.Agent, isWired map[string]bool, failing map[string]error) *[]string {
	t.Helper()
	var did []string
	wasService, wasAutostart, wasShortcut, wasScheme, wasPII, wasExe := removeService, autostartOff, removeShortcut, removeScheme, endPII, executable
	t.Cleanup(func() {
		removeService, autostartOff, removeShortcut, removeScheme, endPII, executable = wasService, wasAutostart, wasShortcut, wasScheme, wasPII, wasExe
	})
	step := func(name string) func() error { return func() error { did = append(did, name); return nil } }
	removeService = step("service")
	autostartOff = step("autostart")
	removeShortcut = func(string) error { did = append(did, "shortcut"); return nil }
	removeScheme = func(string) error { did = append(did, "scheme"); return nil }
	endPII = step("pii")
	executable = func() (string, error) { return "/opt/magpie/magpie", nil }
	return &did
}

func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gone(t *testing.T, p string) bool {
	t.Helper()
	_, err := os.Stat(p)
	return errors.Is(err, os.ErrNotExist)
}

// Uninstalling stops the service before anything else (one still serving
// would wire agents again), disconnects only the agents magpie is in,
// restarts the Codex app's app-server when Codex was one of them, and
// removes magpie's files last.
func TestRunOrder(t *testing.T) {
	codex := &agent.Agent{ID: "codex", Name: "Codex"}
	claude := &agent.Agent{ID: "claude", Name: "Claude Code"}
	gemini := &agent.Agent{ID: "gemini", Name: "Gemini CLI"}
	did := fakes(t, []*agent.Agent{codex, claude, gemini}, map[string]bool{"codex": true, "claude": true}, nil)
	settings := filepath.Join(appdir.Config(), "settings.json")
	write(t, settings, "{}")
	write(t, filepath.Join(appdir.Cache(), "models.json"), "{}")

	r := Run(context.Background())

	want := []string{"service", "autostart", "shortcut", "scheme", "pii"}
	if !slices.Equal(*did, want) {
		t.Fatalf("did %q, want %q", *did, want)
	}
	if r.Failed() || r.CodexApp {
		t.Fatalf("report: %+v", r)
	}
	if !gone(t, appdir.Config()) || !gone(t, appdir.Cache()) {
		t.Fatal("magpie's folders left")
	}
	if r.Exe != "/opt/magpie/magpie" {
		t.Fatalf("exe %q", r.Exe)
	}
	if len(r.Agents) != 0 {
		t.Fatalf("agents %+v", r.Agents)
	}
}

// An agent that can't be disconnected is said, and the rest are still
// done; Codex not among them leaves the Codex app alone.
func TestRunSaysWhatFailed(t *testing.T) {
	fakes(t, nil, nil, nil)
	removeService = func() error { return errors.New("service failed") }
	r := Run(context.Background())
	if !r.Failed() {
		t.Fatal("failure lost")
	}
	if len(r.Agents) != 0 || r.CodexApp {
		t.Fatalf("touched an agent: %+v", r)
	}
}

// Skills and servers the user's agents use from the Library stay; the
// rest of magpie's folder goes.
func TestDataKeepsLibraryInUse(t *testing.T) {
	fakes(t, nil, map[string]bool{}, nil)
	skill := filepath.Join(library.Dir(), "skills", "docx", "SKILL.md")
	write(t, skill, "# docx")
	settings := filepath.Join(appdir.Config(), "settings.json")
	write(t, settings, "{}")
	write(t, filepath.Join(appdir.Config(), "usage", "2026-10.db"), "x")

	r := &Report{}
	Data(r)
	if !gone(t, settings) || !gone(t, filepath.Join(appdir.Config(), "usage")) {
		t.Fatal("magpie's own files left")
	}
	if gone(t, skill) {
		t.Fatal("a skill in use removed")
	}
	if len(r.Kept) != 1 || r.Kept[0].What != library.Dir() {
		t.Fatalf("kept %+v", r.Kept)
	}

	// an empty Library goes with the rest
	os.RemoveAll(filepath.Join(library.Dir(), "skills", "docx"))
	r = &Report{}
	Data(r)
	if !gone(t, appdir.Config()) || len(r.Kept) != 0 {
		t.Fatalf("empty library kept: %+v", r)
	}
}

// Never the home, a root or a relative path, from a variable pointing at
// one; a folder inside another is listed once.
func TestFoldersNeverHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	remove, _ := Folders()
	if len(remove) == 0 {
		t.Fatal("nothing to remove")
	}
	for _, p := range remove {
		if !filepath.IsAbs(p) || within(home, p) {
			t.Fatalf("would remove %s (home %s)", p, home)
		}
		for _, q := range remove {
			if p != q && within(p, q) {
				t.Fatalf("%s listed inside %s", p, q)
			}
		}
	}
	root := "/"
	if runtime.GOOS == "windows" {
		root = `C:\`
	}
	h := filepath.Join(root, "Users", "me")
	for p, ok := range map[string]bool{
		"":                                    false,
		"magpie":                              false,
		root:                                  false,
		h:                                     false,
		filepath.Join(root, "Users"):          false,
		filepath.Join(h, ".config", "magpie"): true,
		filepath.Join(root, "Users", "me2"):   true,
	} {
		if removable(p, h) != ok {
			t.Errorf("removable(%q) = %v", p, !ok)
		}
	}
}

// What the app starts as it quits removes magpie's files once the app is
// gone.
func TestAfterExit(t *testing.T) {
	fakes(t, nil, map[string]bool{}, nil)
	settings := filepath.Join(appdir.Config(), "settings.json")
	write(t, settings, "{}")
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	r := AfterExit(cmd.Process.Pid, time.Minute)
	if time.Since(start) > 30*time.Second {
		t.Fatal("waited on a process that had ended")
	}
	if r.Failed() || !gone(t, settings) {
		t.Fatalf("report %+v", r)
	}
}
