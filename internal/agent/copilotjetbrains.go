package agent

// GitHub Copilot in JetBrains IDEs (IntelliJ IDEA, PyCharm, GoLand, …, and
// Android Studio; the github-copilot-intellij plugin) takes models of other
// vendors through Bring Your Own Key: Copilot Chat's model picker, Manage
// Models…, Add Custom Endpoint Provider (KamToHung on Discord). The plugin
// keeps none of it itself: it asks the Copilot language server it bundles
// (copilot-language-server, the npm @github/copilot-language-server), which
// keeps every provider in one JSON object, byok.json in
// $XDG_CONFIG_HOME/github-copilot when that is set (and absolute),
// %USERPROFILE%\AppData\Local\github-copilot on Windows and
// ~/.config/github-copilot elsewhere, read again at every use. A custom
// endpoint provider <name> is three keys there, as the server's own
// copilot/byok/saveCustomProviderConfig and saveModel write them:
//
//	"<name>-provider-config": {"groupName": <name>, "apiType": "chatCompletions"}
//	"<name>-api-key": "<key, in plain text>"
//	"<name>-models-config": {"<model id>": {"deploymentUrl": …, "isRegistered": true,
//	  "isCustomModel": true, "modelCapabilities": {"name": …, "maxInputTokens": …,
//	  "maxOutputTokens": …, "toolCalling": true, "vision": …}}}
//
// So magpie adds the provider "magpie" there: its catalog, each model at the
// gateway's /v1/chat/completions, sent with the key as a Bearer token (the
// key is the file's, not the IDE's password safe). Tool calling is on, as
// Copilot's agent mode lists only models that call tools; a model's
// maxInputTokens is the window the server fills, its output kept within it.
// The user's other providers and keys stay as they are.
//
// The IDE lists the server's providers when Copilot signs in, so the models
// show after the IDE restarts. The model is picked in Copilot Chat, kept in
// the IDE's project files, so what magpie sets is whether its models are in
// that picker. BYOK needs a GitHub sign-in with Copilot: the server offers
// it to a personal plan (Free, Pro, Pro+) and to Business or Enterprise
// where the organisation's policy allows it. Its requests say
// "GithubCopilot/<server version>" as every client of that server does, so
// they are told apart by their key. Xcode's and Eclipse's Copilot run the
// same server and read the same file: they list magpie's models too.
//
// Checked with the language server 1.545.1 bundled in plugin 1.18.0, under
// a sandbox HOME signed in to a real Copilot account: the file magpie
// writes is listed by copilot/byok/listCustomProviderConfigs and
// listModels, and a Chat turn on one of its models reached the gateway with
// magpie's key. Not with a JetBrains IDE's own window.

import (
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
	"github.com/yetone/magpie/internal/edit"
)

// copilotJBID is the agent's id, and what its token names.
const copilotJBID = "copilot-jetbrains"

// The keys of magpie's provider in byok.json.
const (
	copilotJBConfig = magpieID + "-provider-config"
	copilotJBKey    = magpieID + "-api-key"
	copilotJBModels = magpieID + "-models-config"
)

// copilotJBPlugin is the folder the plugin is installed in, under an IDE's
// plugins folder.
const copilotJBPlugin = "github-copilot-intellij"

func copilotJetBrains(home string) *Agent {
	return copilotJetBrainsAt(copilotBYOKPath(home), func() bool { return copilotJBInstalled(home) })
}

// copilotBYOKPath is byok.json where the language server keeps it
// (getXdgConfigPath).
func copilotBYOKPath(home string) string {
	if x := appdir.Getenv("XDG_CONFIG_HOME"); x != "" && filepath.IsAbs(x) {
		return filepath.Join(x, "github-copilot", "byok.json")
	}
	if runtime.GOOS == "windows" {
		// the server's os.homedir(), USERPROFILE, which home is too
		return filepath.Join(home, "AppData", "Local", "github-copilot", "byok.json")
	}
	return filepath.Join(home, ".config", "github-copilot", "byok.json")
}

// copilotJBInstalled reports whether some JetBrains IDE, or Android Studio,
// has the Copilot plugin installed: in the plugins folder of an IDE's
// config folder on macOS and Windows, in the IDE's data folder itself on
// Linux.
func copilotJBInstalled(home string) bool {
	var base, sub string
	switch runtime.GOOS {
	case "darwin":
		base, sub = filepath.Join(home, "Library", "Application Support"), "plugins"
	case "windows":
		if base = appdir.Getenv("APPDATA"); base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
		sub = "plugins"
	default:
		if base = appdir.Getenv("XDG_DATA_HOME"); base == "" {
			base = filepath.Join(home, ".local", "share")
		}
	}
	for _, vendor := range []string{"JetBrains", "Google"} {
		ms, _ := filepath.Glob(filepath.Join(base, vendor, "*", sub, copilotJBPlugin))
		for _, m := range ms {
			if isDir(m) {
				return true
			}
		}
	}
	return false
}

func copilotJetBrainsAt(path string, installed func() bool) *Agent {
	wired := func() bool { _, ok := edit.GetJSON(path, copilotJBConfig); return ok }
	return &Agent{
		ID: copilotJBID, Name: "Copilot (JetBrains)", Icon: "githubcopilot",
		Aliases: []string{"jetbrains", "jetbrains-copilot", "copilot-idea", "copilot-intellij", "intellij", "idea"},
		Dir:     filepath.Dir(path), Path: path,
		// the folder is every Copilot's (VS Code's sign-in too): the
		// plugin installed in an IDE, or magpie's provider written, says
		// it is here
		detect: func() bool { return installed() || wired() },
	}
}
