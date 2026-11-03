package agent

import (
	"strings"

	"github.com/yetone/magpie/internal/gateway"
)

// kept is what kept.away knows of this magpie, read before a sync's writes:
// the gateway's address (gateway.URL) and the LAN sharing key.
type kept struct{ gateway, lan string }

// ours says v, with any "Bearer " before it, is a key magpie gives an
// agent: gateway.Token, one that names the agent (gateway.TokenFor), or the
// LAN sharing key (ourKey's, as read before).
func (k kept) ours(v string) bool {
	v = strings.TrimPrefix(v, "Bearer ")
	return v == gateway.Token || v != "" && v == k.lan || strings.HasPrefix(v, gateway.Token+"-")
}
