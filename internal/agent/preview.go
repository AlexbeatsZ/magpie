package agent

import (
	"regexp"
	"strings"
)

// dryProviders is, in the process DryRun runs, the ids of the providers the
// magpie that asked for the preview has on.
var dryProviders map[string]bool

// FileChange is how disconnecting changes one of the agent's files.
type FileChange struct {
	Path    string     `json:"path"`              // as the user knows it, ~ for home
	Removed bool       `json:"removed,omitempty"` // the file goes
	Created bool       `json:"created,omitempty"` // the file is made
	Lines   []LineDiff `json:"lines,omitempty"`
}

// LineDiff is a line of a file disconnecting changes: Op is "-" for a line
// taken out, "+" for one put in, "~" for a line whose value goes back to
// another (Was is the line now).
type LineDiff struct {
	Op   string `json:"op"`
	Text string `json:"text"`
	Was  string `json:"was,omitempty"`
}

// diffLines is the lines that differ between two texts, in the order of the
// new one: a line taken out, one put in, and a line whose key stays with its
// value changed (model = "x" → model = "y") as one going back. Values that
// read as secrets are masked. Between the lines both start and end with,
// texts too far apart to compare line by line (OpenCode's table of
// thousands of magpie models taken out) keep the lines of the new one
// found in order in the old.
func diffLines(a, b string) []LineDiff {
	x, y := splitLines(a), splitLines(b)
	pre := 0
	for pre < len(x) && pre < len(y) && x[pre] == y[pre] {
		pre++
	}
	suf := 0
	for suf < len(x)-pre && suf < len(y)-pre && x[len(x)-1-suf] == y[len(y)-1-suf] {
		suf++
	}
	keepX, keepY := make([]bool, len(x)), make([]bool, len(y))
	for i := 0; i < pre; i++ {
		keepX[i], keepY[i] = true, true
	}
	for i := 1; i <= suf; i++ {
		keepX[len(x)-i], keepY[len(y)-i] = true, true
	}
	mx, my := x[pre:len(x)-suf], y[pre:len(y)-suf]
	if kx, ky, ok := sameLines(mx, my, 4000); ok {
		copy(keepX[pre:], kx)
		copy(keepY[pre:], ky)
	} else if len(mx)*len(my) <= 50_000_000 {
		// each line of the one, in order, kept where the other has it
		// next: what is left of a file once a block is taken out of it
		// (past that, all of it taken out and put in)
		i := 0
		for j := range my {
			for k := i; k < len(mx); k++ {
				if mx[k] == my[j] {
					keepX[pre+k], keepY[pre+j] = true, true
					i = k + 1
					break
				}
			}
		}
	}
	var out []LineDiff
	var gone, came []string
	flush := func() {
		// a line taken out and one put in under the same key, as deep in
		// the file, are one going back (Claude's own "model" with
		// modelPicker's taken out, not one of the picker's)
		for _, c := range came {
			k := lineKey(c)
			paired := false
			for i, g := range gone {
				if k != "" && lineKey(g) == k && indent(g) == indent(c) {
					out = append(out, LineDiff{Op: "~", Text: mask(c), Was: mask(g)})
					gone = append(gone[:i], gone[i+1:]...)
					paired = true
					break
				}
			}
			if !paired {
				out = append(out, LineDiff{Op: "+", Text: mask(c)})
			}
		}
		for _, g := range gone {
			out = append(out, LineDiff{Op: "-", Text: mask(g)})
		}
		gone, came = nil, nil
	}
	i, j := 0, 0
	for i < len(x) || j < len(y) {
		switch {
		case i < len(x) && j < len(y) && keepX[i] && keepY[j]:
			flush()
			i++
			j++
		case i < len(x) && !keepX[i]:
			if strings.TrimSpace(x[i]) != "" {
				gone = append(gone, x[i])
			}
			i++
		default:
			if strings.TrimSpace(y[j]) != "" {
				came = append(came, y[j])
			}
			j++
		}
	}
	flush()
	// a bracket or brace left alone on its line says nothing
	var kept []LineDiff
	for _, l := range out {
		if t := strings.Trim(strings.TrimSpace(l.Text), "{}[](),"); t == "" {
			continue
		}
		kept = append(kept, l)
	}
	return kept
}

// sameLines is which lines of x and of y a shortest edit keeps (Myers'
// diff: time and memory grow with the edits, not the files — an agent's
// config can be 30,000 lines, and magpie changes a few). Not ok past most
// edits.
func sameLines(x, y []string, most int) (keepX, keepY []bool, ok bool) {
	n, m := len(x), len(y)
	if most > n+m {
		most = n + m
	}
	off := most + 1
	v := make([]int, 2*off+1)
	// trace[d] is v as step d found it, for k in -d-1..d+1
	var trace [][]int
	get := func(d, k int) int { return trace[d][k+d+1] }
	for d := 0; d <= most; d++ {
		trace = append(trace, append([]int(nil), v[off-d-1:off+d+2]...))
		for k := -d; k <= d; k += 2 {
			var xi int
			if k == -d || (k != d && v[off+k-1] < v[off+k+1]) {
				xi = v[off+k+1]
			} else {
				xi = v[off+k-1] + 1
			}
			yi := xi - k
			for xi < n && yi < m && x[xi] == y[yi] {
				xi++
				yi++
			}
			v[off+k] = xi
			if xi < n || yi < m {
				continue
			}
			// back from the end, marking the diagonals walked
			keepX, keepY = make([]bool, n), make([]bool, m)
			xi, yi = n, m
			for e := d; e >= 0; e-- {
				k := xi - yi
				var pk int
				if k == -e || (k != e && get(e, k-1) < get(e, k+1)) {
					pk = k + 1
				} else {
					pk = k - 1
				}
				px := get(e, pk)
				py := px - pk
				for xi > px && yi > py && xi > 0 && yi > 0 {
					keepX[xi-1], keepY[yi-1] = true, true
					xi--
					yi--
				}
				xi, yi = px, py
			}
			return keepX, keepY, true
		}
	}
	return nil, nil, false
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(strings.ReplaceAll(s, "\r\n", "\n"), "\n"), "\n")
}

// lineKey is what a line sets: TOML's and .env's key before =, JSON's and
// YAML's before :, "" for a line that sets nothing.
var keyRe = regexp.MustCompile(`^\s*(?:export\s+)?"?([A-Za-z0-9_.\-]+)"?\s*[:=]`)

func lineKey(l string) string {
	if m := keyRe.FindStringSubmatch(l); m != nil {
		return m[1]
	}
	if t := strings.TrimSpace(l); strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
		return t
	}
	return ""
}

func indent(l string) string { return l[:len(l)-len(strings.TrimLeft(l, " \t"))] }

var secretKey = regexp.MustCompile(`(?i)(key|token|secret|password|auth)`)

// mask hides the value of a line whose key reads as a secret's; magpie's
// own token, which is no secret, stays.
func mask(l string) string {
	k := lineKey(l)
	// …_TOKENS is a count of them (CLAUDE_CODE_MAX_CONTEXT_TOKENS)
	if k == "" || !secretKey.MatchString(k) || strings.HasSuffix(strings.ToUpper(k), "TOKENS") {
		return l
	}
	m := keyRe.FindStringIndex(l)
	if m == nil { // a table header or compact array, not a key/value line
		return l
	}
	val := strings.TrimSpace(l[m[1]:])
	bare := strings.Trim(val, `"', `)
	if bare == "" || bare == "magpie" || strings.HasPrefix(bare, "{") || strings.HasPrefix(bare, "$") {
		return l
	}
	return l[:m[1]] + " ••••"
}
