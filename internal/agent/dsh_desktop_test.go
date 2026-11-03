package agent

import (
	"testing"
)

// #2325: DeepSeek Harness Desktop installed to D:\tools\dsh was not found.
// Its installer registers the folder under HKCU for the current user and
// under HKLM for all users, and its Uninstall entry names the exe too.
func TestDshDesktopRegisteredAnyHive(t *testing.T) {
	const guid = "1bf39983-50d0-5fe0-9ef4-cece76f67c5e"
	own := `Software\` + guid
	uninstall := `Software\Microsoft\Windows\CurrentVersion\Uninstall\` + guid
	for _, c := range []struct {
		name     string
		reg      map[string]string // hive|key|name
		dir, ver string
	}{
		{"nothing registered", nil, "", ""},
		{"per user", map[string]string{"HKCU|" + own + "|InstallLocation": `D:\tools\dsh`, "HKCU|" + uninstall + "|DisplayVersion": "0.2.0"}, `D:\tools\dsh`, "0.2.0"},
		{"for all users", map[string]string{"HKLM|" + own + "|InstallLocation": `D:\tools\dsh`, "HKLM|" + uninstall + "|DisplayVersion": "0.2.1"}, `D:\tools\dsh`, "0.2.1"},
		{"the Uninstall entry's own folder", map[string]string{"HKLM|" + uninstall + "|InstallLocation": `"D:\tools\dsh"`}, `D:\tools\dsh`, ""},
		{"the Uninstall entry's icon", map[string]string{"HKCU|" + uninstall + "|DisplayIcon": `D:\tools\dsh\DeepSeek Harness.exe,0`}, `D:\tools\dsh`, ""},
		{"the Uninstall entry's uninstaller", map[string]string{"HKCU|" + uninstall + "|UninstallString": `"D:\tools\dsh\Uninstall DeepSeek Harness.exe" /currentuser`}, `D:\tools\dsh`, ""},
		{"an icon that is no exe", map[string]string{"HKCU|" + uninstall + "|DisplayIcon": `D:\tools\dsh\icon.ico`}, "", ""},
	} {
		read := func(hive, key, name string) string { return c.reg[hive+"|"+key+"|"+name] }
		dir, ver := registeredIn(guid, read)
		if dir != c.dir || ver != c.ver {
			t.Errorf("%s: got %q %q, want %q %q", c.name, dir, ver, c.dir, c.ver)
		}
	}
}
