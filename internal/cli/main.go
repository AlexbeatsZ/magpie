// Package cli is magpie's command line: every command, and the app, tray
// and TUI it opens (the root's main.go only calls Main).
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/yetone/magpie/internal/agent"
	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/browsermcp"
	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/claudebridge"
	"github.com/yetone/magpie/internal/claudecode"
	"github.com/yetone/magpie/internal/claudetui"
	"github.com/yetone/magpie/internal/codexrelay"
	"github.com/yetone/magpie/internal/davsync"
	"github.com/yetone/magpie/internal/gateway"
	"github.com/yetone/magpie/internal/imagemcp"
	"github.com/yetone/magpie/internal/netproxy"
	"github.com/yetone/magpie/internal/notices"
	"github.com/yetone/magpie/internal/pluginrt"
	"github.com/yetone/magpie/internal/proc"
	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/sessions"
	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/tui"
	"github.com/yetone/magpie/internal/update"
)

// version is the root main package's, set at build time
// (-ldflags "-X main.version=…") and handed over by Main.
var version = "dev"

const usage = `magpie — local API gateway, providers and model routing

  magpie                          open the app: a window plus a menu bar icon
  magpie tray                     start in the menu bar only
  magpie panel                    open the menu bar icon's quick panel, or close it
  magpie autostart [on|off]       open magpie (in the menu bar) when you log in, or say whether it does
  magpie tui                      the same thing, in the terminal (serves the gateway while open when no magpie does)
  magpie web [--addr host:port] [--lan] [--no-open] [--gateway] [--tls-cert <file> --tls-key <file>]
                                  the app's window in a browser, with the gateway (no desktop needed: WSL, a server over SSH)
                                  a new key each run; MAGPIE_WEB_KEY (16+ letters, digits, - . _ ~) keeps one, signed in for 400 days
                                  behind a reverse proxy, MAGPIE_WEB_URL=https://<the page there> prints the link through it
                                  --gateway: hide local Agent and Library management; Sessions remains available
                                  --tls-cert/--tls-key: the page and the gateway over https:// to other computers
  magpie <command> --help         what a command takes; -h or --help anywhere does nothing else
  magpie library                  manage instructions, MCP servers and skills (model setup stays manual)

  magpie backup [--no-keys] [--no-library] [file]    providers, keys, settings, profiles, agent models and the library in one file, sealed with a passphrase
  magpie restore [--no-agents] [--no-library] <file> put a backup in on this machine
  magpie uninstall [--yes]                           remove magpie's service and files
  magpie webdav [on <address>|set k=v…|now|off]      the same, kept the same on every computer through a WebDAV folder (magpie webdav help)
  magpie s3 [on s3://<bucket>[/<prefix>]|set k=v…|now|off]   the same through an S3-compatible bucket: AWS, R2, B2, MinIO… (magpie s3 help)
  magpie github [on github://<owner>/<repo>[/<folder>]|set k=v…|now|off]   the same in an encrypted GitHub repository (magpie github help)

  magpie providers                list your providers: host, key, models, who uses them
  magpie presets                  the vendors magpie knows: add one with just a key
  magpie provider add <preset> <key>   e.g. magpie provider add deepseek sk-…
  magpie provider add <name> k=v…      a custom vendor (magpie provider for the fields)
  magpie provider key|models|test|rm <id>
  magpie provider fallback <id> <provider/model>…   use these when it's out of quota or down
  magpie import [-y] <link>       add the provider a magpie://import?… link describes
  magpie models [<agent>]         every model agents can pick, as provider/model; an agent's, and why others aren't
  magpie model name <provider/model> <name>|--reset       the name a model goes by, everywhere
  magpie model efforts <provider/model> <l>,<l>|--reset   the reasoning levels a model offers (magpie model help)
  magpie search [add <api> <key>|rm <api>]   Tavily, Brave, Exa, Firecrawl, You.com, DeepSeek, Doubao Search (Volcengine), Yuanbao Search (Tencent Cloud) or SearXNG for web search when no provider can search
  magpie groups                   routing groups: several models agents pick as one, group/<id>
  magpie group add <name> models=<m1>,<m2> [routing=smart|order|rotate|usage|pace|weight|free|fast|cheap|drain] [stays=auto|session|turn|off]
  magpie group <id> | set <id> k=v… | rm <id>   show, change or remove one (magpie group help for more)
  magpie accounts [agent] [--json]  every subscription magpie knows, with each one's allowance used and when it resets
  magpie accounts add <agent>     sign in to one more Claude, ChatGPT or Google (Gemini CLI, Antigravity) subscription
  magpie accounts add copilot [--host <name>.ghe.com]   one more Copilot account, on github.com or an enterprise's GHE.com
  magpie accounts discovery [on|off]   read agents' sign-ins (default on; MAGPIE_NO_ACCOUNTS=1 disables too)
  magpie accounts import <codex|claude|antigravity|factory> <file>... [--yes]   bring in accounts from other tools' files (Codex's auth.json, Cockpit Tools, CLIProxyAPI, Sub2API, claude-swap); the tool a ChatGPT or Claude file came from is signed out of it
  magpie accounts switch <agent> <email|name>   sign the agent in to another of them, by its email or the name you gave it
  magpie accounts switch <claude|codex> next   to the next account in order with room, or the one with the most left
  magpie accounts refresh         renew the saved ChatGPT and Claude sign-ins now (the gateway does it every 8 days)
  magpie accounts checkin         WorkBuddy's daily check-in (签到) for each WorkBuddy account, now (Settings can do it daily)
  magpie accounts project <gemini|antigravity> <email> <project>   the Google Cloud project a Google account's requests go to
  magpie accounts alias <agent> <email> [<name>|--clear]   a name of your own for an account, shown on its card, in the menu bar and in magpie quota
  magpie accounts switch-only codex [on|off]   only switch Codex's own sign-in to an account with room; its requests go straight to OpenAI, never through magpie
  magpie accounts switch-only claude [on|off]  only switch Claude Code's own sign-in to an account with room; magpie serves no requests on the Claude accounts
  magpie claude-code [install [--yes]|remove]   the Claude Code a Claude subscription runs; install downloads Anthropic's own build, checked against its manifest, for a server or container without one
  magpie plugin [add <package>|rm|update|on|off|login <provider>|logout <provider>]
                                  OpenCode provider plugins and pi packages: subscriptions signed in to, and served, through a plugin
  magpie plugin move|migrate <subscription>   run a built-in subscription's accounts on its community plugin

  magpie serve [--web[=host:port]] [--tls-cert <file> --tls-key <file>]
                                  run the gateway alone (the app runs it too); --web serves magpie web's page too, from the same process (127.0.0.1:3430 unless told);
                                  with a certificate and key (PEM; MAGPIE_TLS_CERT, MAGPIE_TLS_KEY) other computers reach both over https://
  magpie daemon [install|uninstall|start|stop|restart|status]
                                  magpie serve as a service of yours, from boot or login and started again when it stops (systemd --user on Linux, launchd on the Mac)
  magpie daemon install [--web] [--lan] [--addr host:port] [--no-web] [--tls-cert <file> --tls-key <file>|--no-tls]
                                  --web: the service serves magpie web's page too, with a key kept across restarts; installed again, it keeps what it served
  magpie daemon web [on|off] [--lan] [--addr host:port]
                                  the service's page: its link, or turn it on or off
  magpie healthcheck              exit 0 when the gateway answers (a container's HEALTHCHECK)
  magpie heap [<file>]            a Go heap profile of the running magpie, with its memory numbers, for a report of it using too much memory
  magpie cpu [<file>] [--seconds N]   a Go CPU profile of the running magpie over 30s (5..120) while you use it, for a report of it using too much CPU
  magpie goroutines [<file>]      every goroutine of the running magpie, its stack and how long it waited, for a report of it stuck at something
  magpie log [--path] [-n N] [-f]  where magpie's logs are, to send with a report: the app's (app.log), the service's; -n prints the app log's last N lines, -f follows it
  magpie gateway-key list|add <name> [--read-only]|rotate <id>|remove <id>   manage the keys clients use to call a shared gateway
  magpie gateway-key limit <id> [list|off|reset [n]|remove <n>|[add] day|week|month|<N>d|<N>h --tokens N --cost USD --cache-reads]   a key's own limits, and what it used
  magpie gateway-key models <id> [all|<provider>/<model>|<provider>/* ...]   the models a key may use, every one unless it names some
  magpie gateway-key quotas <id> [on|off]   whether a key held to a budget, models or accounts is told the quotas of what it may use
  magpie gateway-key access <id> [read-only|full]   read-only: the key reads quotas and usage (GET /, /v1/magpie/quotas, /v1/magpie/usage) and uses no model, MCP server or write
  magpie gateway-key move <id> <place>   put a key at a place in the list (from 1), the order list and the app show them in
  magpie gateway-key issue --ttl 1h [--name n] [--models a,b] [--accounts a,b] [--tokens N] [--cost USD] [--read-only] [--from <id>] [--json]
                                         a short-lived key (1m to 30 days) that stops working by itself; --from holds it to a key that issues keys; prints the key
  magpie gateway-key issuer <id> [on|off]   let a key issue short-lived keys under itself (POST /v1/magpie/keys with it), each held to its models, accounts, limits and life
  magpie gateway-key oidc add --key <id> --repo owner/name [--ref refs/heads/main] [--environment name] [--claim name=value]… [--ttl 1h] [--issuer URL] [--audience aud]
                                         let a CI job's OIDC token (GitHub Actions by default) be traded for a key issued under <id> (POST /v1/magpie/keys/oidc); oidc list | remove <rule id>
  magpie mcp image                the image and video generation MCP server an agent is given from the library (stdio)
  magpie tools [<name>]           the tools magpie's MCP servers give agents (images, vision, web search, plugins'), with their parameters
  magpie tool <name> [--<param> <value>…] [<words>] [-o <file|folder>] [--json]   call one through the running magpie; an image or video it makes is saved
  magpie usage [today|7d|30d|all] tokens and cost per agent, model and subscription account (30d)
  magpie usage --csv [--account <name>] [today|7d|30d|all]   every request as CSV (or one account's): the model asked for, sent and served, tokens, cost, time, status, account
  magpie usage [--csv [--account <name>]] --last <dur>|--days <n>   the same for a window ending now (24h, 90m, 1.5 days), not from a midnight
  magpie sessions [--model <m>] [--folder <f>] [--json]   the latest Claude Code, Codex, OpenCode and Pi sessions, with what each cost
  magpie sessions --days N|today|all [--model <m>] [--folder <f>] [--json]
                                  what every session spent, day by day, with the top models and folders (7 days)
  magpie sessions read [on|off all|<agent>…]   which agents' session history magpie may read (none until allowed)
  magpie quota [<provider>] [--json]  what is left of every subscription, plan and key balance
  magpie quota wait <provider|account> [--timeout <d>] [--quiet]
                                  block until that subscription (any of its accounts) or account has allowance again
  magpie quota history [<provider|account>] [--days N] [--json]
                                  each window's readings over time, kept 45 days
  magpie statusline [<provider|account>]   one line for Claude Code's status line: the session's model and each account's windows
  magpie sync                     refresh the model catalog and vendor model lists
  magpie settings [list|get <key>|set <key> <value>|reset <key>] [--json]
                                  every setting the window's Settings page saves, read and set as it saves them (magpie settings help)
  magpie update [check] [--proxy <url>] [--mirror <prefix>]
                                  install the newest release (check: only say if there is one); --proxy: an
                                  http(s):// or socks5:// proxy for it; --mirror: a GitHub download mirror put
                                  before the github.com URL (none unless given; still checked against usemagpie.ai's SHA-256)
  magpie update mirror [<prefix>|off]  the mirror every update, the app's own too, is downloaded through
  magpie update channel [stable|every]  which versions updates offer: stable (one a week, an urgent one at
                                  once) or every version as it comes out (the default)
  magpie update auto [on|off] [download|notify] [30m|1h|6h|24h]
                                  whether the app looks for updates by itself, how often (6h), and whether
                                  it downloads one at once or only says it is out (notify: for metered networks)
  magpie licenses                 the third-party notices of what magpie is built with

`

var (
	bold  = lipgloss.NewStyle().Bold(true)
	muted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8B8F98", Dark: "#7C8290"})
	faint = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C4C7CE", Dark: "#4A4F5A"})
	green = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0F9D58", Dark: "#7EE2A8"})
)

// Main runs magpie with os.Args, at the version the build was stamped with.
func Main(v string) {
	version = v
	// before anything reads or writes a file: no home, or a relative one,
	// would put the agents' configs and magpie's keys under the working
	// folder. What needs no file still answers (magpie version in a
	// container or a script without HOME): run makes magpie's folders first.
	ignored, err := appdir.CheckEnv()
	if err != nil {
		if len(os.Args) > 1 {
			switch os.Args[1] {
			case "-v", "--version", "version":
				fmt.Println("magpie", version)
				return
			case "-h", "--help", "help":
				fmt.Print(usage)
				return
			case "licenses":
				fmt.Print(notices.Text)
				return
			}
		}
		fmt.Fprintln(os.Stderr, "magpie:", err)
		os.Exit(1)
	}
	slices.Sort(ignored)
	for _, v := range ignored {
		fmt.Fprintf(os.Stderr, "magpie: ignoring %s: not an absolute path (the programs magpie starts still get it)\n", v)
	}
	if provider.TookOpenedURL(os.Args[1:]) {
		// Claude Code, signing in for magpie, handed over the page to open
		return
	}
	endProbesOnSignal()
	gateway.Version = version
	netproxy.Install()
	update.GUI = hasGUI
	err = run(os.Args[1:])
	proc.EndProbes() // a CLI still being asked something isn't left to init
	sessions.Saved() // the session index kept, for the next run
	if err != nil {
		// a command with exit codes of its own (quota wait) says which
		code := 1
		var e exitError
		if errors.As(err, &e) {
			code = e.code
		}
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, "magpie:", msg)
		}
		os.Exit(code)
	}
}

// runTUI runs the TUI, which quits on Ctrl+C and SIGTERM itself once it
// has started; it asks CLIs first, and a signal then ends those.
func runTUI() error {
	return tuiRun(ownSignals)
}

// tuiRun is tui.Run; a var so tests can stand in for it.
var tuiRun = tui.Run

// syncAgentsOnChange sets what a catalog change (a provider or group
// saved, a vendor's list fetched anew) does: the model lists agents keep
// in files of their own are brought up to date behind the save, which
// answers without waiting for every agent's files to be written (on
// Windows, half a second for each one another program holds open).
func syncAgentsOnChange() { catalog.Changed = nil }

// startUp brings the files magpie and its agents keep up to what this
// magpie reads, before a command that may read them; a var so a test can
// tell it ran.
var startUp = func() {
	// started by an update, the magpie it replaces goes first: the moves
	// below write the files it may still be writing
	update.AwaitPredecessor()
	makeDirs()
	settings.Migrate()
	// every built-in plugin starts on (plugins.yml), on a new install too
	if _, err := pluginrt.Settle(); err != nil {
		fmt.Fprintf(os.Stderr, "magpie: plugins.yml: %v\n", err)
	}
	if _, err := pluginrt.Refresh(); err != nil {
		fmt.Fprintf(os.Stderr, "magpie: plugins.yml: %v\n", err)
	}
	// a vendor's keys saved a provider each ("NVIDIA #001"…) are one
	// provider with those keys, before anything lists them
	if _, err := provider.FoldCopies(); err != nil {
		fmt.Fprintf(os.Stderr, "magpie: providers.json: %v\n", err)
	}

}

func run(args []string) error {
	// a command that saved a provider or a group ends with the agents'
	// files written, which its save no longer waits for; and a test's run
	// leaves catalog.Changed as it found it
	was := catalog.Changed
	defer func() {
		agent.SyncCatalogSettled()
		catalog.Changed = was
	}()
	// before anything is read, moved or started: --help on any command
	// prints its usage and does nothing else
	if helpAsked(args) {
		fmt.Println(strings.TrimRight(helpFor(args), "\n"))
		return nil
	}
	if len(args) > 0 && args[0] == "healthcheck" {
		return healthcheck() // every few seconds in a container: nothing else
	}
	if len(args) > 0 && args[0] == "heap" {
		// the running magpie's heap profile: asks it, and moves nothing here
		return heapCmd(args[1:])
	}
	if len(args) > 0 && args[0] == "cpu" {
		// the running magpie's CPU profile: asks it, and moves nothing here
		return cpuCmd(args[1:])
	}
	if len(args) > 0 && args[0] == "goroutines" {
		// the running magpie's goroutines: asks it, and moves nothing here
		return goroutinesCmd(args[1:])
	}
	if len(args) > 0 && (args[0] == "log" || args[0] == "logs") {
		// where magpie's logs are, and the app log's last lines: reads
		// files, and moves nothing
		return logCmd(args[1:])
	}
	if len(args) > 1 && (args[0] == "accounts" || args[0] == "account") && args[1] == "discovery" {
		return accountsDiscoveryCmd(args[2:])
	}
	// what an agent starts for each session or message talks to the
	// running gateway and nothing else: the moves below list every
	// provider and account, which starts the plugin host and asks claude,
	// cursor-agent and devin who is signed in — five processes and a
	// Claude /usage for each Claude Code session's image MCP (#3318)
	if len(args) > 0 {
		switch args[0] {
		case "mcp":
			if len(args) > 1 && args[1] == "browser" {
				return browsermcp.Run(args[2:])
			}
			if len(args) > 1 && args[1] == "search" {
				return browsermcp.RunSearch(args[2:])
			}
			if len(args) > 1 && args[1] == "vision" {
				return browsermcp.RunVision(args[2:])
			}
			return imagemcp.Run(args[1:])
		case "claude-mcp-helper": // internal: stdio MCP subprocess spawned by Claude Code
			return claudebridge.RunMCP(args[1:])
		case claudetui.Command: // internal: Claude Code's interactive mode for the bridge
			// it ends its Claude Code and removes its files on SIGINT,
			// SIGTERM and SIGHUP itself: magpie dying of the signal first
			// left them behind
			leaveSignals()
			return claudetui.Run(args[1:])
		case "statusline":
			return statuslineCmd(args)
		case "tools":
			return toolsCmd(args[1:])
		case "tool":
			return toolCmd(args[1:])
		case codexrelay.Command: // internal: what a quitting magpie leaves for Codex's open threads
			return codexrelay.Main(args[1:], gateway.CodexRelay())
		case "uninstall":
			// what the app starts as it quits from Settings › Uninstall
			// magpie: nothing is made again before its folders go
			if len(args) > 1 && args[1] == "--data-after" {
				return uninstallData(args[2:])
			}
		}
	}
	startUp()
	// a provider added, edited or removed, or a list fetched anew, reaches
	// the model lists agents keep in files of their own, behind the save
	// that changed it (agent.SyncCatalogSoon)
	syncAgentsOnChange()
	// the setup kept the same on every computer, by whichever serves
	gateway.WhileServing = append(gateway.WhileServing, davsync.Run)
	// and the Claude Code magpie downloaded for a machine with none (a
	// server's, a container's), at Claude Code's newest release: its runs
	// turn its own updater off
	gateway.WhileServing = append(gateway.WhileServing, claudecode.KeepCurrent)
	// and the request archive, when it is on, goes to the bucket sync is to
	gateway.ArchiveBucket = func() (gateway.Putter, bool) {
		if b, ok := davsync.S3Bucket(); ok {
			return b, true
		}
		return nil, false
	}
	if len(args) == 0 {
		if hasGUI {
			return runGUI(true, "")
		}
		gateway.Serves = "tui"
		return runTUI()
	}
	// a magpie:// link the system handed over (Windows, Linux): the app
	// opens it for the user to confirm
	if strings.HasPrefix(strings.ToLower(args[0]), "magpie:") {
		if !hasGUI {
			return importCmd(args)
		}
		return runGUI(false, args[0])
	}
	// how this process was started, for the Routing page of another magpie on
	// this computer to name the one that serves the gateway (#2606)
	switch args[0] {
	case "tui", "web", "serve", "tray":
		gateway.Serves = args[0]
	case "app", "gui", "panel", "-Embedding":
		gateway.Serves = "app"
	}
	switch args[0] {
	case "tui":
		return runTUI()
	case "web":
		return webCmd(args[1:])
	case "app", "gui":
		// `magpie gui settings`: the window on that tab, as a restart to
		// update from it comes back (update.RelaunchArgs)
		if len(args) > 1 {
			return runWindow(args[1])
		}
		return runGUI(true, "")
	case "tray":
		return runGUI(false, "")
	case "-Embedding":
		// Windows starting magpie for a click on one of its notifications
		// (a usage alert, #368) left in the Action Center after it quit:
		// the window, on the Usage page
		return runWindow("usage")
	case "panel":
		return runPanel()
	case "autostart":
		return autostartCmd(args[1:])
	case "daemon":
		return daemonCmd(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	case "-v", "--version", "version":
		fmt.Println("magpie", version)
		return nil
	case "licenses":
		fmt.Print(notices.Text)
		return nil
	case "sync":
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := catalog.Sync(ctx); err != nil {
			return err
		}
		fmt.Println(green.Render("✓"), "catalog saved to", catalog.CachePath())
		refreshLive(ctx)
		return nil
	case "import":
		return importCmd(args[1:])
	case "providers":
		return providers()
	case "presets":
		return presets()
	case "provider":
		return providerCmd(args)
	case "models":
		return models(args[1:])
	case "model":
		return modelCmd(args[1:])
	case "search":
		return searchCmd(args[1:])
	case "groups":
		return groups()
	case "group":
		return groupCmd(args)
	case "serve":
		return serveCmd(args[1:])
	case "gateway-key":
		return gatewayKeys(args)
	case "accounts", "account":
		return accountsCmd(args)
	case "usage":
		return usageCmd(args)
	case "sessions":
		return sessionsCmd(args)
	case "library":
		return libraryCmd(args)
	case "quota", "quotas":
		return quotaCmd(args)
	case "update":
		return updateCmd(args)
	case "backup":
		return backupCmd(args[1:])
	case "restore":
		return restoreCmd(args[1:])
	case "uninstall":
		return uninstallCmd(args[1:])
	case "webdav", "dav":
		return webdavCmd(args[1:])
	case "plugin", "plugins":
		return pluginCmd(args)
	case "s3":
		return s3Cmd(args[1:])
	case "github":
		return githubCmd(args[1:])
	case "claude-code":
		return claudeCodeCmd(args[1:])
	case "settings":
		return settingsCmd(args[1:])
	}

	return fmt.Errorf("unknown command %q; run magpie help for gateway commands", args[0])
}

func set(a *agent.Agent, key, value string) error {
	f := a.Field(key)
	if f == nil {
		var keys []string
		for _, f := range a.Fields {
			keys = append(keys, f.Key)
		}
		return fmt.Errorf("%s has no field %q (fields: %s)", a.Name, key, strings.Join(keys, ", "))
	}
	// a field's name is no value: magpie codex subagent effort ran Codex's
	// subagents on a model called effort. Claude Code's opus is a model's
	// name as well as a tier's
	if g := a.Field(value); value != "" && g != nil && !modelAlias(a, value) {
		return fieldAsValue(a, f, g, value)
	}
	value, err := a.Spell(f.Key, value)
	if err != nil {
		return err
	}
	before := f.Get()
	wired := a.Wired()
	// a magpie model picked on an agent not connected yet connects it
	// first, as the Agents page's picker does (#1900): the agent is wired
	// in, and Disconnect puts back what it had before
	if err := a.Pick(f.Key, value); err != nil {
		return err
	}
	// what the config reads now, not what was asked: an agent may name the
	// model under a provider of its own (OpenCode's magpie-relay/…), and a
	// value it had already is said to be so
	now := f.Get()
	shown := now
	if value == "" || now == "" {
		shown = muted.Render("default")
	}
	if now == before {
		shown += " " + muted.Render("(unchanged)")
	}
	fmt.Println(green.Render("✓"), bold.Render(a.Name), muted.Render(f.Label), shown)
	// a value that is the agent's own, not one of magpie's, takes magpie
	// out: said, as Disconnect is, with the way back (#2000: magpie claude
	// sonnet put Claude Code back on its DeepSeek endpoint, which refused
	// Sonnet, and only "✓ Claude Code model sonnet" was printed)
	if wired && now != "" && !a.Wired() {
		fmt.Println(muted.Render(fmt.Sprintf("  disconnected from magpie: %s is %s's own, sent to the endpoint it had before magpie; magpie %s <model> with one of magpie's models connects it again", now, a.Name, a.ID)))
	}
	if a.Notice != nil && (!f.Live || !wired && a.Wired()) {
		if n := a.Notice(); n != "" {
			fmt.Println(muted.Render("  ↻ " + n))
		}
	}
	return nil
}

// modelAlias says whether a word names the agent's model, in any case,
// though a field goes by it too: Claude Code's opus, a tier's name as well
func modelAlias(a *agent.Agent, w string) bool {
	return slices.ContainsFunc(a.ModelAliases, func(m string) bool { return strings.EqualFold(m, w) })
}

// fieldAsValue refuses field g's name given as field f's value, for what
// was meant: two words that are one field's label (magpie codex subagent
// effort), a switch, which is turned on (Claude Code's ultracode), or the
// field, which its name alone shows.
func fieldAsValue(a *agent.Agent, f, g *agent.Field, value string) error {
	for _, name := range []string{f.Key + " " + value, f.Key + "_" + value} {
		if h := a.Field(name); h != nil {
			return fmt.Errorf("%q is one field of %s's: magpie %s %s shows it", h.Label, a.Name, a.ID, h.Key)
		}
	}
	if g.Options != nil {
		if o := g.Options(a.Values()); len(o) == 1 && o[0].Value == "on" {
			return fmt.Errorf("%q is a switch of %s's, not a value for its %s: magpie %s %s on", value, a.Name, f.Label, a.ID, g.Key)
		}
	}
	return fmt.Errorf("%q is a field of %s's, not a value for its %s: magpie %s %s shows it", value, a.Name, f.Label, a.ID, g.Key)
}

func fieldForValue(a *agent.Agent, v string) *agent.Field {
	vals := a.Values()
	// the agent's suffix after a model (omp's ":max") aside: a role on the
	// same model at that level offers it as typed, and would take it. A list
	// of models no picker offers: it is for the model
	if a.SplitSuffix != nil {
		m, _, one := a.SplitSuffix(v)
		if !one {
			return nil
		}
		v = m
	}
	// a model stays with the model, even where other fields offer it too
	// (Claude Code's opus/sonnet/haiku/fable)
	for _, o := range a.Fields[0].Options(vals) {
		if o.Value == v {
			return nil
		}
	}
	for i := 1; i < len(a.Fields); i++ {
		for _, o := range a.Fields[i].Options(vals) {
			if o.Value == v {
				return &a.Fields[i]
			}
		}
	}
	return nil
}

func tilde(p string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

func pad(s string, w int) string {
	if n := w - lipgloss.Width(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}
