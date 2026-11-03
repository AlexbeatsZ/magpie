package agent

import (
	"testing"

	"github.com/yetone/magpie/internal/usage"
)

// its requests come with GitHubCopilotChat/<version>
func TestVSCodeUA(t *testing.T) {
	if got := usage.AgentOf("GitHubCopilotChat/0.69.0"); got != "vscode" {
		t.Fatalf("%q", got)
	}
}

// VS Code's Agents window runs Codex under the name vscode_agent_host,
// its User-Agent and originator: its calls are VS Code's, not a row of
// their own named vscode_agent_host (justruman on Discord). VS Code's
// Copilot-proxied Codex and Claude names are not it.
func TestVSCodeAgentHostUA(t *testing.T) {
	for _, ua := range []string{"vscode_agent_host/0.163.0 (Mac OS 26.2.0; arm64) unknown (vscode_agent_host; 0.1.0)", "vscode_agent_host"} {
		if got := usage.AgentOf(ua); got != "vscode" {
			t.Errorf("%q: %q", ua, got)
		}
	}
}
