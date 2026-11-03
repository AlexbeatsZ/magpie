package agent

import (
	"github.com/yetone/magpie/internal/access"
	"github.com/yetone/magpie/internal/gateway"
)

// keyAt is the key an agent that reaches the gateway at gw sends it. On
// this machine's loopback that is gateway.Token, which the gateway takes
// from this computer only; from another machine — a WSL distro under NAT,
// which asks Windows' address — the gateway turns a request without a
// named key away, so it is the LAN sharing key, while sharing is on
// (whqtian on Discord: omp, then OpenCode, in WSL were refused). At the
// address Windows has on a WSL network, which the gateway listens on while
// it is on loopback, it is the WSL key (gateway.WSLKey), sharing or not
// (zaidu259434 on X: Codex in WSL failed with sharing off).
func keyAt(gw string) string {
	if onAnotherMachine(gw) {
		if gateway.OnWSLNet(gw) {
			if k := gateway.WSLKey(); k != "" {
				return k
			}
		}
		if k := access.LANSecret(); k != "" {
			return k
		}
	}
	return gateway.Token
}

// KeyAt is keyAt, for what else magpie gives an agent that reaches the
// gateway at gw (the library's MCP servers relayed through it).
func KeyAt(gw string) string { return keyAt(gw) }

// gwKey is keyAt for the agent at p.
func (p place) gwKey() string { return keyAt(p.gw()) }

// agentKeyAt is keyAt for an agent whose key on this machine names it
// (gateway.TokenFor), for requests that carry nothing else of it.
func agentKeyAt(agent, gw string) string {
	k := keyAt(gw)
	if wk := gateway.WSLKeyMade(); k == wk && wk != "" {
		return wk + "-" + agent // the gateway reads it as TokenFor(agent)
	}
	if k != gateway.Token {
		return k
	}
	return gateway.TokenFor(agent)
}

// agentKey is agentKeyAt for the gateway's own address, as hereKey.
func agentKey(agent string) string { return agentKeyAt(agent, gateway.URL()) }

// ourKey says k is a key magpie gives an agent: gateway.Token, the LAN
// sharing key one that reaches the gateway from elsewhere is given, or the
// WSL key one in a WSL distro under NAT is.
func ourKey(k string) bool {
	if k == gateway.Token {
		return true
	}
	return k != "" && (k == access.LANSecret() || k == gateway.WSLKeyMade())
}
