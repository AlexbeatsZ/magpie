package agent

// MiniMax Code — MiniMax's mcode CLI and the desktop app, which run the same
// runtime (#362) — keeps its settings in config.yaml in its data folder,
// $MINIMAX_DATA_DIR or ~/.minimax. The model sessions start on is
// defaultModel, "<provider>/<model>" split at the first slash, where a
// provider the user added (Add 3rd-party provider…, mcode provider add) is
// "custom_provider:<key>", an entry of custom_provider. magpie adds itself
// there as custom_provider.magpie — the gateway, spoken to as Anthropic
// messages, with the catalog as its models — so the catalog joins MiniMax
// Code's /model. A model there is keyed by its id's flat spelling,
// "bb-codex~gpt-6.1-sol" (provider.FlatID), which the gateway takes back to
// "bb-codex/gpt-6.1-sol": a SubAgent's model must be "<provider>/<model>"
// with one slash in the whole (#1387), so a model through magpie is
// "custom_provider:magpie/bb-codex~gpt-6.1-sol". A key written before
// that, with the slashes, has its keys moved to the flat one and stays
// beside it switched off, for the sessions that started on it, and the
// settings that named it name the flat one (mcodeRefs). The default the
// user had is stashed and put back when magpie steps out. MiniMax Code
// rewrites the file itself (its /model, its provider commands), so magpie's
// entry is edited in place: a key the user or MiniMax Code put on it, or on
// one of its models, stays.

import (
	"path/filepath"
	"strings"

	"github.com/yetone/magpie/internal/appdir"
)

// mcodeUA is the User-Agent magpie's entry sends: MiniMax Code 0.5's own was
// the Anthropic SDK's ("L/JS 0.91.1"), or node's for its token counts.
// 0.6.8 sends "MiniMaxCode" on a custom provider that sets none, which names
// it too (#1605: a provider pointed at magpie by hand).
const mcodeUA = "minimax-code"

// MiniMaxDir is the folder MiniMax Code keeps its config.yaml in.
func MiniMaxDir(home string) string {
	if d := strings.TrimSpace(appdir.Getenv("MINIMAX_DATA_DIR")); d != "" {
		return d
	}
	return filepath.Join(home, ".minimax")
}

func miniMax(home string) *Agent { return miniMaxAt(here(home), MiniMaxDir(home)) }

// miniMaxIn is MiniMax Code's mcode in a WSL distro (see wsl.go): ~/.minimax,
// as MINIMAX_DATA_DIR there isn't read.
func miniMaxIn(at place) *Agent { return miniMaxAt(at, at.join(".minimax")) }

// miniMaxAt is MiniMax Code with its data folder at dir, reaching the
// gateway as at does.
func miniMaxAt(_ place, dir string) *Agent {

	path := filepath.Join(dir, "config.yaml")

	return &Agent{
		ID: "minimax-code", Name: "MiniMax Code", Icon: "minimax-color", Aliases: []string{"mcode"},
		UA:  []string{mcodeUA, "minimaxcode"},
		Bin: "mcode", Dir: dir, Path: path,
	}
}
