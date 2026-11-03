package agent

// AtomCode (atomgit.com/atomgit_atomcode/atomcode, a Rust terminal agent)
// keeps its settings in ~/.atomcode/config.toml ($ATOMCODE_HOME). A model is
// a [models."<key>"] table naming a [provider_accounts.<account>] table —
// what its own CodingPlan sign-in writes since 5.0.7: the account holds the
// protocol, the gateway's /v1 and the key; the model table holds the model
// string, the context window (AtomCode compacts as it nears it), the output
// limit, whether it sees images, the reasoning levels its picker offers and
// the one in use. magpie adds the account "magpie" and one model table per
// catalog model, keyed "magpie/<provider>/<model>" — the model string goes
// to the gateway as it is, for it to resolve — and points default_provider
// and default_model at the one picked, default_model winning when both are
// set; the defaults the user had are stashed and put back when magpie steps
// out. Requests are Chat Completions with Bearer auth and a User-Agent of
// atomcode/<version>, which the gateway knows apart; the key is AtomCode's
// own, so a request is its agent's even when the header is overridden.

import (
	"path/filepath"
)

func atomcode(home string) *Agent { return atomcodeIn(here(home)) }

// atomcodeIn is AtomCode at a place: this machine's home, or a WSL distro's
// (see wsl.go), where ATOMCODE_HOME isn't read and its account names the
// gateway as the distro reaches it, with the key it takes from there.
func atomcodeIn(at place) *Agent {
	dir := at.join(".atomcode")
	if h := at.getenv("ATOMCODE_HOME"); filepath.IsAbs(h) {
		dir = filepath.Clean(h)
	}
	path := filepath.Join(dir, "config.toml")

	return &Agent{
		ID: "atomcode", Name: "AtomCode", Icon: "atomcode",
		UA:  []string{"atomcode"},
		Bin: "atomcode", Dir: dir, Path: path,
	}
}
