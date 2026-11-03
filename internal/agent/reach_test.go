package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The probe takes any HTTP answer for one (a gateway shared on the network
// answers 401 without a key) and a refused connection for none.
func TestReachProbe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("asked %s", r.URL.Path)
		}
		http.Error(w, "no key", http.StatusUnauthorized)
	}))
	if !realReachProbe(srv.URL) {
		t.Fatal("a 401 taken for nothing listening")
	}
	srv.Close()
	if realReachProbe(srv.URL) {
		t.Fatal("a closed port taken for an answer")
	}
}
