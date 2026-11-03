package agent

// AstrBot (AstrBotDevs/AstrBot, a chatbot framework for QQ, Telegram,
// Discord, Lark and the rest) keeps every setting in one file,
// <root>/data/cmd_config.json, where the root is ASTRBOT_ROOT, else ~/.astrbot
// for the desktop app (AstrBot-desktop's runtime_paths.rs, and
// astrbot_path.get_astrbot_root for a packaged desktop runtime), else the
// folder it was started in: `astrbot init` marks one with a .astrbot file,
// and `astrbot run` there sets ASTRBOT_ROOT to it (cli/commands/cmd_run.py).
//
// Since 4.10 the file splits a chat provider into two lists (default.py,
// the dashboard's buildModelProviderConfig):
//
//	"provider_sources": [{"id":…,"provider":"openai",
//	  "type":"openai_chat_completion","provider_type":"chat_completion",
//	  "enable":true,"key":["…"],"api_base":"…/v1","timeout":120,
//	  "proxy":"","custom_headers":{}}],
//	"provider": [{"id":"<source>/<model>","enable":true,
//	  "provider_source_id":"<source>","model":"…",
//	  "modalities":["text","image","tool_use"],"custom_extra_body":{},
//	  "max_context_tokens":0}]
//
// — a source is an endpoint and its keys, each entry of provider one model
// of a source, and what AstrBot runs is the two merged, the model's keys
// winning (ProviderManager.get_merged_provider_config). The types are
// config_service.validate_config's, from default.py's schema: key a list of
// strings, timeout and max_context_tokens ints (0: AstrBot looks the window
// up itself), custom_headers and custom_extra_body objects, modalities a
// list of text/image/audio/tool_use.
//
// magpie adds a source "magpie", OpenAI Compatible at the gateway's /v1 with
// a key naming AstrBot (its requests carry no User-Agent of its own), and an
// entry "magpie/<provider>/<model>" for every model of magpie's, and makes
// the one picked AstrBot's default chat model: agent_runner.config.model.
// provider_id, which AstrBot reads when agent_runner.runner_type is "local"
// (4.28's ProviderManager._resolve_using_provider), or
// provider_settings.default_provider_id in a file from before agent_runner,
// which AstrBot's migration (migra_helper.py) carries over. The user's own
// sources and models are left where they are, byte for byte, and so is
// what they changed in magpie's own entries in AstrBot's WebUI (a source's
// timeout or proxy, a model switched off, its extra body). A model picked
// for one chat with /provider, which AstrBot keeps in its own database, is
// the user's and stands. The file is copied to cmd_config.json.before-magpie
// before magpie first writes it.
//
// AstrBot reads the file once, at start, and keeps it in memory: its WebUI
// writes the whole of it back on every save. So magpie's change needs
// AstrBot restarted, and a WebUI save made before that restart writes over
// it (Check then says so). AstrBot writes the file with a UTF-8 BOM
// (astrbot_config.py, encoding utf-8-sig), which magpie keeps.
//
// AstrBot in Docker is not one magpie wires: its data folder is a volume
// magpie can't know, and the gateway on this machine's loopback is not the
// container's.

import (
	"path/filepath"
	"strings"
)

// astrbotRoot is the folder AstrBot keeps data/ in: ASTRBOT_ROOT, else the
// desktop app's ~/.astrbot, else a folder `astrbot init` was run in that
// magpie can know of — the home itself (a .astrbot file there), or ~/astrbot
// or ~/AstrBot, where its README and a clone of its source put it. A root
// none of these is reaches magpie through ASTRBOT_ROOT. With none found it is
// the desktop app's, which then isn't there.
func astrbotRoot(at place) string {
	if d := strings.TrimSpace(at.getenv("ASTRBOT_ROOT")); d != "" {
		return d
	}
	desktop := at.join(".astrbot")
	if isFile(filepath.Join(desktop, "data", "cmd_config.json")) {
		return desktop
	}
	if isFile(desktop) {
		return at.home
	}
	for _, d := range []string{"astrbot", "AstrBot"} {
		root := at.join(d)
		if isFile(filepath.Join(root, "data", "cmd_config.json")) || isFile(filepath.Join(root, ".astrbot")) {
			return root
		}
	}
	return desktop
}

func astrbot(home string) *Agent { return astrbotAt(here(home)) }

func astrbotAt(at place) *Agent {
	root := astrbotRoot(at)
	data := filepath.Join(root, "data")
	path := filepath.Join(data, "cmd_config.json")

	return &Agent{
		ID: "astrbot", Name: "AstrBot", Icon: "astrbot",
		Bin: "astrbot", Dir: data, Path: path,
	}
}
