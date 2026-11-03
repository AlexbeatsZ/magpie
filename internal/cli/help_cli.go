package cli

import (
	"strings"
)

// helpAsked is whether args ask for help rather than for the command to
// run: -h or --help anywhere, or help as the word after the command. It
// is looked at before run does anything, since a command that took
// --help for a value acted on it: `magpie snow --help` set Snow's model
// to "--help", which took magpie out of it (#1457), and `magpie serve
// --help` started a gateway.
func helpAsked(args []string) bool {
	// what magpie is started with by others (Claude Code's MCP helper, the
	// Agents page's dry run, a link, Windows' -Embedding) never carries
	// an exact -h, --help or help
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return len(args) > 1 && args[1] == "help"
}

// helpFor is what `magpie <command> --help` prints: the command's own
// usage where it has one, its lines of magpie's usage otherwise, an
// agent's forms and fields for an agent, and the whole usage for
// anything else.
func helpFor(args []string) string {
	cmd := args[0]
	switch cmd {
	case "-h", "--help", "help":
		return usage
	case "model":
		if len(args) > 1 && (args[1] == "spec" || args[1] == "specs") {
			return specUsage
		}
		return modelUsage
	case "group", "groups":
		if len(args) > 1 && (args[1] == "rule" || args[1] == "rules") {
			return ruleUsage
		}
		return groupUsage
	case "plugin", "plugins":
		return pluginUsage
	case "library":
		return libraryUsage
	case "quota", "quotas":
		return quotaUsage
	case "statusline":
		return statuslineUsage
	case "heap":
		return heapUsage
	case "cpu":
		return cpuUsage
	case "goroutines":
		return goroutinesUsage
	case "log", "logs":
		return logUsage
	case "sessions":
		return sessionsUsage
	case "search":
		return searchUsage
	case "tool", "tools":
		return toolUsage
	case "webdav", "dav":
		return webdavUsage
	case "s3":
		return s3Usage
	case "github":
		return githubUsage
	case "claude-code":
		return claudeCodeUsage
	case "settings":
		return settingsUsage
	case "list":
		cmd = "ls"
	case "account":
		cmd = "accounts"
	case "app", "gui":
		cmd = ""
	}
	if cmd != "" && !strings.HasPrefix(cmd, "-") {
		if lines := usageLines(cmd); lines != "" {
			return lines
		}
	}
	return usage
}

// usageLines are the lines of magpie's usage for `magpie <cmd>`, with the
// lines that carry on from them.
func usageLines(cmd string) string {
	var b strings.Builder
	in := false
	for _, l := range strings.Split(usage, "\n") {
		if rest, ok := strings.CutPrefix(l, "  magpie "); ok {
			in = rest == cmd || strings.HasPrefix(rest, cmd+" ")
		} else if !strings.HasPrefix(l, "          ") {
			in = false
		}
		if in {
			b.WriteString(l + "\n")
		}
	}
	return b.String()
}
