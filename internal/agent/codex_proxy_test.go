package agent

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A proxy's own 502 isn't magpie; a proxy that passes the request on is.
func TestRealProxyProbe(t *testing.T) {
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"name":"magpie"}`)
	}))
	defer gw.Close()
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream", http.StatusBadGateway)
	}))
	defer dead.Close()
	pass := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := url.Parse(r.RequestURI)
		res, err := http.Get(u.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer res.Body.Close()
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	}))
	defer pass.Close()
	if realProxyProbe(gw.URL+"/v1", dead.URL) {
		t.Fatal("a proxy answering 502 counted as reaching magpie")
	}
	if !realProxyProbe(gw.URL+"/v1", pass.URL) {
		t.Fatal("a proxy passing the request on didn't count")
	}
	// Codex's reqwest has no SOCKS: a socks5:// proxy is asked as an HTTP
	// one at that address. Clash's mixed port passes it on; a SOCKS-only
	// port, which Go would speak SOCKS to, doesn't, and Codex 502s
	if !realProxyProbe(gw.URL+"/v1", "socks5://"+strings.TrimPrefix(pass.URL, "http://")) {
		t.Fatal("an HTTP proxy named socks5:// didn't count, though Codex gets through it")
	}
	if realProxyProbe(gw.URL+"/v1", "socks5://"+socksOnly(t)) {
		t.Fatal("a SOCKS-only proxy counted, though Codex can't speak SOCKS")
	}
}

// socksOnly is the address of a SOCKS5 proxy that speaks nothing else: it
// passes a CONNECT on and drops anything that isn't SOCKS5.
func socksOnly(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				c.SetDeadline(time.Now().Add(5 * time.Second))
				b := make([]byte, 262)
				if _, err := io.ReadFull(c, b[:2]); err != nil || b[0] != 5 {
					return
				}
				if _, err := io.ReadFull(c, b[:b[1]]); err != nil {
					return
				}
				c.Write([]byte{5, 0})
				if _, err := io.ReadFull(c, b[:4]); err != nil || b[3] != 1 {
					return
				}
				if _, err := io.ReadFull(c, b[:6]); err != nil {
					return
				}
				to := net.JoinHostPort(net.IP(b[:4]).String(), strconv.Itoa(int(b[4])<<8|int(b[5])))
				up, err := net.Dial("tcp", to)
				if err != nil {
					return
				}
				defer up.Close()
				c.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
				c.SetDeadline(time.Time{})
				go io.Copy(up, c)
				io.Copy(c, up)
			}()
		}
	}()
	return ln.Addr().String()
}
