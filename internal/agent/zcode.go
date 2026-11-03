package agent

import (
	"net"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/yetone/magpie/internal/appdir"
)

// ZCode (Zhipu's desktop app) keeps its model providers in
// ~/.zcode/v2/config.json (under the folder the user moved its data to, when
// they did: appdir.ZCodeDataIn), OpenCode's provider shape with a kind of its own:
//
//	{"provider":{"magpie":{"name":"magpie","kind":"anthropic",
//	  "options":{"apiKey":"magpie","baseURL":"http://127.0.0.1:3425"},
//	  "enabled":true,"source":"custom",
//	  "models":{"<id>":{"name":…,"limit":{"context":…,"output":…},"modalities":{…},
//	    "reasoning":{"enabled":true,"levels":[…],"defaultLevel":…,
//	      "providerOptionsByLevel":{"<level>":{"thinking":{"type":"enabled","budget_tokens":…}}}}}}}}}
//
// An anthropic provider is asked at baseURL + /v1/messages. The model is
// picked per task in ZCode's own picker and kept in its window, not in a
// file, so what magpie sets is whether its models are in that picker.
//
// ZCode 3.14 moved its providers to ~/.zcode/v2/provider_config.json and
// reads config.json's only once, to import them, so a provider added there
// later never reached its picker. There a provider is a rule, and a model's
// context window and inputs are rules of their own:
//
//	{"schemaVersion":1,"config":{
//	  "providerConfigRules":{"providerRules":[{"providerId":"magpie",
//	    "providerName":"magpie","enabled":true,"config":{
//	      "group":"standard-personal",
//	      "access":{"type":"api-key","apiKey":"magpie"},
//	      "api":{"type":"anthropic-messages","baseUrl":"http://127.0.0.1:3425"},
//	      "personalModelIds":[…],"modelOrder":[…]}}]},
//	  "modelConfigRules":{"providerModelRules":[{"providerId":"magpie",
//	    "modelId":…,"config":{"properties":{"contextWindow":…,
//	      "inputFormat":{"supportsImage":…},"supportsNativeWebSearch":true},
//	      "optionSpecs":{"maxOutputTokens":{"max":…},"reasoningLevel":{"values":[…]}}}}],
//	    "manualProviderModelRules":[…]}}}
//
// magpie writes both files, so an older ZCode sees its models too. A model
// the user set by hand in ZCode (a manual rule) keeps what they set.
//
// magpie's gateway answers all three of ZCode's API formats at the same
// address (/v1/messages, /chat/completions, /responses), so the format the
// user picked in ZCode's provider settings is theirs: a sync keeps it in both
// files (zcodeFormat), and only a provider magpie adds new starts on
// Anthropic Messages.

// zcodeOS is the OS ZCode's home is found for, so a test can ask for
// Windows' (appdir.ZCodeHomeOn).
var zcodeOS = runtime.GOOS

func zcode(home string) *Agent {
	// where ZCode keeps its data, moved there in its settings or not, under
	// the home ZCode reads (HOME first, on Windows too)
	dir := filepath.Join(appdir.ZCodeDataIn(appdir.ZCodeHomeOn(zcodeOS, home)), ".zcode")
	path := filepath.Join(dir, "v2", "config.json")

	return &Agent{
		ID: "zcode", Name: "ZCode", Icon: "zcode", Aliases: []string{"z-code"},
		UA:  []string{"zcode"},
		Dir: dir, Path: path,
	}
}

// onAnotherMachine says base is an http(s) address whose host isn't this
// machine's loopback.
func onAnotherMachine(base string) bool {
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return false
	}
	h := u.Hostname()
	if h == "localhost" {
		return false
	}
	ip := net.ParseIP(h)
	return ip == nil || !ip.IsLoopback() && !ip.IsUnspecified()
}
