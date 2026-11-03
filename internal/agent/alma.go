package agent

// Alma (a desktop AI chat app) keeps its providers and settings in the app,
// not in a file magpie can edit: they are reached through its local REST API
// (http://localhost:23001, while Alma runs).
//
//	GET  /api/providers              [{"id","name","type","baseURL","enabled","models":["<id>",…],"availableModels":[{"id","name",…}]},…]
//	POST /api/providers              {"name","type","apiKey","baseURL","enabled"} → the provider
//	PUT  /api/providers/:id          the fields to change
//	PUT  /api/providers/:id/models   {"models":["<id>",…],"availableModels":[{"id","name","capabilityOverrides"}]}:
//	                                 models, kept as given, are the ones Alma offers; a model's
//	                                 capabilityOverrides, when sent, take the place of the ones it had
//	GET  /api/settings, PUT it back  the whole settings; chat.defaultModel is "<providerId>:<model>",
//	                                 imageGen.model the same ("" for Auto)
//
// magpie is one provider there, named magpie, of type openai at the
// gateway's /v1; choosing one of its models makes it Alma's default model,
// and the model magpie draws with Alma's image generation model when Alma
// has none picked (almaSetImage).
// Alma not running is no error: there is nothing to read or keep current.

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yetone/magpie/internal/appdir"
)

// almaAPI is where Alma's API answers. Under test it is nothing, so no test
// reaches a real Alma; a test points it at a fake one.
var almaAPI = func() string {
	if testing.Testing() {
		return ""
	}
	return "http://localhost:23001"
}()

// almaDir is where Alma keeps its data, there once Alma was installed and
// opened: ~/Library/Application Support/alma on a Mac, the user config
// folder elsewhere. On Linux alma-server, Alma without a desktop, keeps it
// in ALMA_DATA_DIR, else $XDG_DATA_HOME/alma (~/.local/share/alma), and
// answers on the same port: with only the desktop's folder looked for it
// was never found, so its magpie provider never got Alma's key and its
// requests read as the AI SDK's (Lutra.x on Discord). The desktop's folder
// comes first; "" when neither is there.
func almaDir(goos string) string {
	var dirs []string
	if d, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(d, "alma"))
	}
	if goos != "darwin" && goos != "windows" {
		if d := appdir.Getenv("ALMA_DATA_DIR"); d != "" {
			dirs = append(dirs, d)
		}
		if d := appdir.Getenv("XDG_DATA_HOME"); d != "" {
			dirs = append(dirs, filepath.Join(d, "alma"))
		} else if h, err := os.UserHomeDir(); err == nil {
			dirs = append(dirs, filepath.Join(h, ".local", "share", "alma"))
		}
	}
	for _, d := range dirs {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d
		}
	}
	if len(dirs) > 0 {
		return dirs[0]
	}
	return ""
}

// almaOS is the system almaDir looks for Alma's data on; a test sets it.
var almaOS = runtime.GOOS

func alma() *Agent {
	dir := almaDir(almaOS)
	return &Agent{
		ID: "alma", Name: "Alma", Icon: "alma",
		UA:  []string{"alma"},
		Dir: dir, Path: almaAPI,
	}
}
