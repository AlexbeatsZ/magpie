package agent

// Zoo Code (Zoo-Code-Org/Zoo-Code, the VS Code extension
// ZooCodeOrganization.zoo-code; a fork of Roo Code) keeps its provider
// profiles in VS Code's secret storage, encrypted and out of reach. What it
// does read from a file is the one its zoo-code.autoImportSettingsPath
// setting names: on every start of the extension (autoImportSettings,
// src/utils/autoImportSettings.ts, Zoo Code 3.90.0) it imports that file as
// its own Settings → Import does (importSettingsFromPath,
// src/core/config/importExport.ts):
//
//	{"providerProfiles":{"currentApiConfigName":"magpie",
//	  "apiConfigs":{"magpie":{"id":"magpie","apiProvider":"openai",
//	    "openAiBaseUrl":"http://127.0.0.1:3425/v1","openAiApiKey":"magpie-zoo-code",
//	    "openAiModelId":"<model>","openAiCustomModelInfo":{…}}}}}
//
// Imported profiles are merged over the user's own, which stay, and the
// imported currentApiConfigName becomes the profile Zoo Code uses. magpie
// writes that file in Zoo Code's own folder of VS Code's
// (zooCodeImportPath) and points the setting at it in VS Code's settings.json, the default profile's and
// each other profile's (vscodeProfiles), a value of the user's own kept
// aside and put back by Disconnect. Its OpenAI Compatible provider asks
// the gateway's Chat Completions under Bearer openAiApiKey, its requests
// saying "ZooCode/<version>" as their User-Agent.
//
// What this route can't do: Zoo Code reads the file only when it starts
// (the Notice asks for a window reload), and imports it again on each
// start, so while connected each start puts Zoo Code back on magpie's
// profile. A profile once imported lives in Zoo Code's secret storage,
// which magpie can't write, so Disconnect stops the import and leaves the
// profile "magpie" there for the user to switch from or delete. Which
// profile the user picks inside Zoo Code magpie can't see either.
//
// Zoo Code for JetBrains runs the same extension in its own extension
// host, which is handed an empty configuration on every start
// (RPCManager.startInitialize), so the setting can't reach it there, and
// keeps its secrets in ~/.zoo-code/secrets.json, a store of its own with
// no lock and no documented format. magpie doesn't write either.

import (
	"os"
	"path/filepath"
	"strings"
)

// zooExtension is Zoo Code's extension id, as the folder it is installed
// in begins: zoocodeorganization.zoo-code-<version>.
const zooExtension = "zoocodeorganization.zoo-code"

func zooCode(home, cfg string) *Agent {
	// VS Code's own User and extensions folders: a portable build's are
	// both in its data folder
	return zooCodeAt(vscodeDirs(vscodeStable, home, cfg))
}

// zooCodeAt is Zoo Code in the VS Code whose User folder is dir and whose
// extensions are in exts.
func zooCodeAt(dir, exts string) *Agent {
	path := filepath.Join(dir, "settings.json")

	return &Agent{
		ID: "zoo-code", Name: "Zoo Code", Icon: "zoo-code",
		// its requests say "ZooCode/<version>" (DEFAULT_HEADERS,
		// src/api/providers/constants.ts): usage names them by the word
		// before the slash
		Aliases: []string{"zoocode", "zoo"},
		Dir:     dir, Path: path,
		detect: func() bool { return !Taken(dir) && zooCodeInstalled(exts) },
	}
}

// zooCodeInstalled reports whether Zoo Code is in an extensions folder.
func zooCodeInstalled(exts string) bool {
	xs, err := os.ReadDir(exts)
	if err != nil {
		return false
	}
	for _, x := range xs {
		if x.IsDir() && strings.HasPrefix(strings.ToLower(x.Name()), zooExtension+"-") {
			return true
		}
	}
	return false
}
