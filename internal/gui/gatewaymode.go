package gui

import (
	"sync/atomic"
)

// Resource management is available in the app and on the web. The explicit
// web --gateway flag can still leave out local resource management.

// WebGateway is `magpie web --gateway`'s.
var WebGateway atomic.Bool

// webPage is set once the page is served by `magpie web`: settingsState,
// which has no Windows, tells the page its gateway mode by it.
var webPage atomic.Bool

func gatewayMode(web bool) (bool, string) {
	if web && WebGateway.Load() {
		return true, "flag"
	}
	return false, "manual-models"
}
