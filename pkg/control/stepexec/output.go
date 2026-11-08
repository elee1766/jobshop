package stepexec

import (
	"bytes"
	"encoding/json"
	"strings"
)

// OutputMax bounds the stored output of one attempt.
const OutputMax = 64 << 10

// OutcomeOf is the outcome word an attempt set: a top-level JSON
// "outcome" string, or the last stdout line "outcome: <word>". For
// conditional steps (#385).
func OutcomeOf(out string) string {
	var m map[string]any
	if json.Unmarshal([]byte(out), &m) == nil {
		if s, ok := m["outcome"].(string); ok {
			return strings.TrimSpace(s)
		}
		return ""
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if w, ok := strings.CutPrefix(lines[len(lines)-1], "outcome:"); ok {
		return strings.TrimSpace(w)
	}
	return ""
}

// LastLines is the last n lines of s joined with " | ".
func LastLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, " | ")
}

// LimitedBuffer keeps the first OutputMax bytes and drops the rest,
// so a chatty command cannot grow the step ticket.
type LimitedBuffer struct {
	b         bytes.Buffer
	truncated bool
}

func (l *LimitedBuffer) Write(p []byte) (int, error) {
	if room := OutputMax - l.b.Len(); room < len(p) {
		l.truncated = true
		if room > 0 {
			l.b.Write(p[:room])
		}
		return len(p), nil
	}
	return l.b.Write(p)
}

func (l *LimitedBuffer) String() string {
	if l.truncated {
		return l.b.String() + "\n[truncated]"
	}
	return l.b.String()
}
