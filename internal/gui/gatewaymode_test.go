package gui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

// Resource management stays available without model wiring. Only an explicit
// web --gateway hides those pages; old preferences never change the boundary.
func TestGatewayMode(t *testing.T) {
	sandboxHome(t)
	t.Cleanup(func() { WebGateway.Store(false); webPage.Store(false) })
	for _, saved := range []string{"", "on", "off"} {
		s := settings.Load()
		s.GatewayMode = saved
		if err := settings.Save(s); err != nil {
			t.Fatal(err)
		}
		for _, flag := range []bool{false, true} {
			WebGateway.Store(flag)
			for _, web := range []bool{false, true} {
				on, why := gatewayMode(web)
				if on != (web && flag) {
					t.Fatalf("saved=%q flag=%v web=%v: %v %q", saved, flag, web, on, why)
				}
			}
		}
	}
	WebGateway.Store(false)
	w := httptest.NewRecorder()
	Handler(webHost{}, nil).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/settings/gateway-mode", strings.NewReader(`{"mode":"on"}`)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("removed preference route: %d %s", w.Code, w.Body)
	}
}
