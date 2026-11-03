package agent

import (
	"reflect"
	"runtime"
	"testing"
	"unicode/utf16"
)

func utf16le(s string, bom bool) []byte {
	var b []byte
	if bom {
		b = append(b, 0xff, 0xfe)
	}
	for _, u := range utf16.Encode([]rune(s)) {
		b = append(b, byte(u), byte(u>>8))
	}
	return b
}

// wsl.exe -l -q speaks UTF-16LE, a BOM or not, or UTF-8 under WSL_UTF8.
func TestParseDistros(t *testing.T) {
	list := "Ubuntu-24.04\r\ndocker-desktop\r\nDebian\r\n\r\n"
	want := []string{"Ubuntu-24.04", "Debian"}
	for name, b := range map[string][]byte{
		"utf16":     utf16le(list, false),
		"utf16 bom": utf16le(list, true),
		"utf8":      []byte(list),
	} {
		if got := parseDistros(b); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %q", name, got)
		}
	}
	if got := parseDistros(nil); len(got) != 0 {
		t.Errorf("none: %q", got)
	}
}

func TestWSLMirrored(t *testing.T) {
	for cfg, want := range map[string]bool{
		"":                                  false,
		"[wsl2]\nnetworkingMode=mirrored\n": true,
		"\ufeff[WSL2]\r\nnetworkingmode = Mirrored # yes\r\n":           true,
		"[wsl2]\nnetworkingMode=\"mirrored\"\n":                         true,
		"[wsl2]\nnetworkingMode=nat\n":                                  false,
		"[wsl2]\n#networkingMode=mirrored\n":                            false,
		"[experimental]\nnetworkingMode=mirrored\n":                     false,
		"[wsl2]\nmemory=8GB\n[experimental]\nnetworkingMode=mirrored\n": false,
	} {
		if got := wslMirrored(cfg); got != want {
			t.Errorf("%q: %v", cfg, got)
		}
	}
}

func TestParseProbe(t *testing.T) {
	d := parseProbe("Ubuntu", "home:/home/me/\ndir:.codex\nroute:default via 172.20.0.1 dev eth0 proto kernel\nns:nameserver 10.255.255.254\n")
	if d == nil || d.Home != "/home/me" || !d.Has["dir:.codex"] || d.Has["bin:codex"] || d.Gateway != "172.20.0.1" {
		t.Fatalf("%+v", d)
	}
	if d := parseProbe("U", "home:/root\nbin:codex\nns:nameserver 172.30.0.1\n"); d == nil || d.Gateway != "172.30.0.1" || !d.Has["bin:codex"] {
		t.Fatalf("resolv.conf: %+v", d)
	}
	if d := parseProbe("U", "sh: not found\n"); d != nil {
		t.Fatalf("no home: %+v", d)
	}
}

// Off Windows there are none, and nothing is run.
func TestWSLAgentsElsewhere(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("on Windows")
	}
	if len(wslAgents()) != 0 {
		t.Fatal("wsl agents off windows")
	}
}
