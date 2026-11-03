package sessions

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// References follow agent-history's device:harness:native-id convention. The
// native id and Magpie's agent:id storage keys remain unchanged.
var referenceDevice = sync.OnceValue(func() string {
	host, _ := os.Hostname()
	return historyDevice(runtime.GOOS, host)
})

func historyDevice(goos, host string) string {
	if goos == "darwin" {
		return "mac"
	}
	if goos == "windows" {
		if strings.Contains(strings.ToLower(host), "rog") {
			return "rog"
		}
		return "omen"
	}
	host = strings.ToLower(strings.Split(host, ".")[0])
	if referenceName.MatchString(host) {
		return host
	}
	return "local"
}

var (
	referenceName         = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	referenceID           = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{3,127}$`)
	ErrReferenceMissing   = errors.New("no such local session")
	ErrReferenceAmbiguous = errors.New("session prefix matches more than one session; use a longer prefix or the full reference")
)

func historyHarness(agent string) string {
	if agent == "agy" {
		return "antigravity"
	}
	return agent
}

// Reference identifies a native session on this computer. A gateway-only
// conversation has no confirmed source device; a WSL session belongs to a
// separate environment. Neither is labeled as this computer's history.
func Reference(s Session) string {
	if s.Gateway || s.WSL != "" || !referenceName.MatchString(s.Agent) || !referenceID.MatchString(s.ID) {
		return ""
	}
	return referenceDevice() + ":" + historyHarness(s.Agent) + ":" + s.ID
}

// ResolveReference accepts a native id or unique prefix, agent:id, or a
// device:harness:id reference (also wrapped in [session ...]). Reads keep the
// existing per-agent consent gate. Foreign references never resolve locally.
func ResolveReference(input string) (Session, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "[session ") && strings.HasSuffix(input, "]") {
		input = strings.TrimSpace(input[len("[session ") : len(input)-1])
	}
	parts := strings.Split(input, ":")
	var agent, id string
	switch len(parts) {
	case 1:
		id = parts[0]
	case 2:
		agent, id = parts[0], parts[1]
	case 3:
		if !referenceName.MatchString(parts[0]) {
			return Session{}, errors.New("invalid session device")
		}
		if parts[0] != referenceDevice() {
			return Session{}, fmt.Errorf("session belongs to %s; use agent-history-%s on its owning computer", parts[0], parts[0])
		}
		agent, id = parts[1], parts[2]
	default:
		return Session{}, errors.New("use a native id, agent:id or device:harness:id session reference")
	}
	if !referenceID.MatchString(id) || agent != "" && !referenceName.MatchString(agent) {
		return Session{}, errors.New("session id must contain at least 4 letters, digits, dots, underscores or hyphens")
	}
	if agent == "antigravity" {
		agent = "agy"
	}
	var candidates []Session
	if agent != "" {
		for _, s := range ListAgent(agent) {
			candidates = append(candidates, s.Session)
		}
	} else {
		for _, a := range Agents() {
			for _, s := range ListAgent(a.Agent) {
				candidates = append(candidates, s.Session)
			}
		}
	}
	return matchReference(candidates, id)
}

func matchReference(candidates []Session, id string) (Session, error) {
	var exact, prefix []Session
	for _, s := range candidates {
		if Reference(s) == "" {
			continue
		}
		if s.ID == id {
			exact = append(exact, s)
		} else if strings.HasPrefix(s.ID, id) {
			prefix = append(prefix, s)
		}
	}
	if len(exact) > 0 {
		prefix = exact
	}
	switch len(prefix) {
	case 0:
		return Session{}, ErrReferenceMissing
	case 1:
		return prefix[0], nil
	default:
		return Session{}, ErrReferenceAmbiguous
	}
}
