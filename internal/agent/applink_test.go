package agent

import (
	"fmt"
	"testing"
	"time"
)

func TestCommandProgram(t *testing.T) {
	for in, want := range map[string]string{
		`"C:\Users\a b\AppData\Local\Programs\Cindy\Cindy.exe" "%1"`: `C:\Users\a b\AppData\Local\Programs\Cindy\Cindy.exe`,
		`"C:\Program Files\Lody\Lody.exe" -- "%1"`:                   `C:\Program Files\Lody\Lody.exe`,
		`C:\Tools\aionui.exe %1`:                                     `C:\Tools\aionui.exe`,
		`  "C:\x\y.exe"`:                                             `C:\x\y.exe`,
		`"C:\unclosed\y.exe`:                                         `C:\unclosed\y.exe`,
		``:                                                           ``,
		`rundll32.exe url.dll,FileProtocolHandler %1`:                `rundll32.exe`,
	} {
		if got := commandProgram(in); got != want {
			t.Errorf("commandProgram(%q) = %q, want %q", in, got, want)
		}
	}
}

// The system's own answer for a scheme nothing registers: none, not an
// error, where it can be read (the registry, Launch Services, xdg-mime).
func TestSystemLinkHandlerNoSuchScheme(t *testing.T) {
	scheme := fmt.Sprintf("magpie-no-such-%d", time.Now().UnixNano())
	exe, found, err := systemLinkHandler(scheme)
	if err != nil {
		t.Skipf("can't be read here: %v", err)
	}
	if found || exe != "" {
		t.Fatalf("%s:// opens %q", scheme, exe)
	}
}
