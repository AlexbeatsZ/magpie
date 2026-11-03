package agent

// Cindy (a desktop app that runs Claude Code, Codex and Pi) keeps its
// providers in its own database, their keys encrypted, and takes a new one
// only through its import link, which the user confirms in Cindy:
//
//	cindy://provider/import?v=1&data=<base64url of the provider's JSON>
//
// So magpie has nothing to set there: its row is the link, magpie as one
// custom provider with an endpoint for each of Cindy's runtimes, and
// whether Cindy has it already, read from Cindy's database and never
// written.

import (
	"os"
	"path/filepath"

	_ "github.com/yetone/magpie/internal/sqlitedb"
)

// cindyDirs are where Cindy keeps its data (~/Library/Application
// Support/Cindy on a Mac, CindyGlobal for its global build), those there
// first: there once Cindy was installed and opened.
func cindyDirs() []string {
	d, err := os.UserConfigDir()
	if err != nil {
		return nil
	}
	var have, not []string
	for _, n := range []string{"Cindy", "CindyGlobal"} {
		p := filepath.Join(d, n)
		if _, err := os.Stat(p); err == nil {
			have = append(have, p)
		} else {
			not = append(not, p)
		}
	}
	return append(have, not...)
}

func cindy() *Agent {
	dirs := cindyDirs()
	dir := ""
	if len(dirs) > 0 {
		dir = dirs[0]
	}
	return &Agent{
		ID: "cindy", Name: "Cindy", Icon: "cindy",
		UA:  []string{"cindy"},
		Dir: dir, Path: dir,
		detect: func() bool {
			for _, d := range dirs {
				if isDir(d) {
					return true
				}
			}
			return false
		},
	}
}
