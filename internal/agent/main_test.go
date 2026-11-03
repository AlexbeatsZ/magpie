package agent

import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func TestMain(m *testing.M) {
	// Discovery tests never inspect installed apps or packages on this machine.

	windowsStorePackages = func() []string { return nil }
	windowsRegisteredApp = func(string) (string, string) { return "", "" }
	spotlightApps = func(string) []string { return nil }
	os.Exit(testenv.Run(m))
}
