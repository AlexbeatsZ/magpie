package agent

import (
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/yetone/magpie/internal/netproxy"
	"github.com/yetone/magpie/internal/proc"
)

// Codex sends its requests to magpie's gateway, http://127.0.0.1:3425, the
// way it sends every other: through a proxy when one is set (Muma102kMichael
// on X: every project and every new chat 502, "a local gateway config and a
// proxy switch, two pitfalls stacked"). Its HTTP client is reqwest with its
// default features (codex-rs Cargo.toml), whose proxy is hyper-util's
// Matcher::from_system: HTTP_PROXY, else the system's proxy for http://
// (Windows' Internet Options when it names one proxy for all, macOS' "Web
// proxy (HTTP)" entry alone: never its HTTPS or SOCKS one), else ALL_PROXY,
// and nothing goes
// around it but what NO_PROXY names — or, with no NO_PROXY, Windows'
// ProxyOverride — as an IP, a CIDR or "*" for an address: "127.*" and
// "localhost" there don't keep 127.0.0.1 off it, and macOS' exceptions
// aren't read at all. A proxy that sends 127.0.0.1 on to a server of its own
// (Clash in global mode, a rule set with no LAN-direct rule) answers 502,
// and Codex is down on every model, its own included once magpie carries
// them. With [features] respect_system_proxy (codex-rs outbound_proxy),
// Codex asks the system alone: Windows' bypass list is matched with its
// wildcards and <local>, macOS' exceptions likewise, and with no proxy set
// there it goes direct; a SOCKS one alone on macOS it can't use, and takes
// HTTP_PROXY/ALL_PROXY then. reqwest is built without SOCKS (codex-rs'
// Cargo.lock has no tokio-socks), so a socks5:// proxy in ALL_PROXY is
// spoken to as an HTTP one.
//
// The environment is the one Codex runs with: a terminal's, or, for the
// ChatGPT app's Codex, the login shell's, which the app takes as it opens.
// magpie opened from the Dock has neither, so the login shell's are asked
// (proc.LoginEnv), over magpie's own.
//
// So before magpie hands Codex's requests to the gateway uninvited (account
// failover), and while it has them, it asks the way Codex would go: through
// which proxy, and whether a request sent through it reaches magpie (Drift
// kind "proxied").

// codexProxyEnv, codexSystemProxy and codexProxyOS are what codexProxyFor
// reads, and loginEnv what codexEnv does; vars so tests can say.
var (
	codexProxyEnv    = codexEnv
	codexSystemProxy = netproxy.System
	codexProxyOS     = runtime.GOOS
	loginEnv         = proc.LoginEnv
)

// codexEnv is k as Codex has it: the login shell's, else magpie's own.
func codexEnv(k string) string {
	if v := loginEnv(k); v != "" {
		return v
	}
	return os.Getenv(k)
}

func realProxyProbe(base, proxy string) bool {
	p, err := netproxy.Parse(proxy)
	if err != nil {
		return false
	}
	if strings.HasPrefix(strings.ToLower(p.Scheme), "socks") {
		p.Scheme = "http"
	}
	c := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(p)}}
	res, err := c.Get(strings.TrimSuffix(base, "/") + "/")
	if err != nil {
		return false
	}
	return answersAsMagpie(res)
}

// codexProxiedDetail says what keeps Codex off magpie through proxy, and
// how to get it there.
func codexProxiedDetail(base, proxy, from string) string {
	host := strings.TrimPrefix(base, "http://")
	d := "Codex sends its requests for magpie (" + host + ") through the proxy at " + netproxy.Redact(proxy) + ", set in " + from + ", and that proxy doesn't pass them on to magpie: every turn fails, with 502 Bad Gateway, whichever model it is on. " +
		"Have the proxy send 127.0.0.1 direct (a DIRECT rule for 127.0.0.0/8, or rule mode rather than global), "
	switch {
	case strings.Contains(strings.ToUpper(from), "_PROXY"):
		d += "or add 127.0.0.1 to NO_PROXY where " + from + " is set (your shell profile, such as ~/.zshrc, if it's there), "
	case codexProxyOS == "windows":
		d += "or add 127.0.0.1 to \"Don't use the proxy server for addresses beginning with\" in Windows' proxy settings (127.* there is not enough for Codex), "
	default:
		d += "or add export NO_PROXY=127.0.0.1 to your shell profile (~/.zshrc): Codex in a terminal and the ChatGPT app's Codex both read it, "
	}
	return d + "then quit and reopen Codex. Until then, Disconnect from magpie puts Codex's config back as it was."
}
