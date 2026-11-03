package agent

import (
	"testing"
)

// Which Windows processes are each app's: Claude Desktop's claude.exe, not
// Claude Code's; WorkBuddy's WorkBuddy.exe, not WorkBuddy AI's.
func TestRestartAppsOnWindows(t *testing.T) {
	cases := []struct {
		id, image string
		want      bool
	}{
		{"claude-desktop", `C:\Users\u\AppData\Local\AnthropicClaude\app-2.9939.2\claude.exe`, true},
		{"claude-desktop", `C:\Users\u\AppData\Local\AnthropicClaude\claude.exe`, true},
		{"claude-desktop", `C:\Program Files\WindowsApps\Claude_1.0.0.0_x64__pzs8sxrjxfjjc\app\claude.exe`, true},
		{"claude-desktop", `C:\Program Files\WindowsApps\AnthropicPBC.Claude_1.0.0.0_x64__fnn82j28hfe8t\app\Claude.exe`, true},
		{"claude-desktop", `C:\Users\u\.local\bin\claude.exe`, false},
		{"claude-desktop", `C:\Users\u\AppData\Roaming\npm\node_modules\@anthropic-ai\claude-code\bin\claude.exe`, false},
		{"claude-desktop", `C:\Users\u\AppData\Local\AnthropicClaude\Update.exe`, false},
		{"workbuddy", `C:\Program Files\WorkBuddy\WorkBuddy.exe`, true},
		{"workbuddy", `C:\Program Files\WorkBuddy AI\WorkBuddy.exe`, false},
		{"workbuddy", `C:\Program Files\WorkBuddy\resources\cli\codebuddy.exe`, false},
	}
	for _, c := range cases {
		if got := restartApps[c.id].win(c.image); got != c.want {
			t.Errorf("%s %s: %v, want %v", c.id, c.image, got, c.want)
		}
	}
	if restartApps["workbuddy-ai"].win != nil {
		t.Error("WorkBuddy AI's Windows build isn't known, yet it has a matcher")
	}
	if _, ok := restartApps["codex"]; ok {
		t.Error("the Codex app has its own Restart (codex_app.go)")
	}
}
