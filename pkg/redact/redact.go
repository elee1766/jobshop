// Package redact removes secrets from text and JSON before it leaves the
// controller: session.journal pages, journal shipping (#530) and command
// step output. It was conversation export's redaction (#45), moved here
// when that export was removed.
package redact

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/zricethezav/gitleaks/v8/detect"
)

// Mask replaces each redacted value.
const Mask = "[REDACTED]"

// Redactor holds the secret values of one env. Build it once per session
// and reuse it: the env scan is per call otherwise.
type Redactor struct {
	env []string // secret values, longest first
}

// New returns a redactor for env (KEY=VALUE entries): values of
// secret-named keys are replaced wherever they appear.
func New(env []string) *Redactor { return &Redactor{env: SecretEnvValues(env)} }

// Text replaces env secret values and gitleaks findings in s.
func (r *Redactor) Text(s string) string {
	for _, v := range r.env {
		s = strings.ReplaceAll(s, v, Mask)
	}
	d := leaks()
	if d == nil || s == "" {
		return s
	}
	for _, f := range d.DetectString(s) {
		if f.Secret != "" {
			s = strings.ReplaceAll(s, f.Secret, Mask)
		}
	}
	return s
}

// JSON redacts values under secret-named keys, then env secret values and
// gitleaks findings in the encoded text. The result is valid JSON when
// raw is: replacements only touch string contents.
func (r *Redactor) JSON(raw json.RawMessage) json.RawMessage {
	s := r.Text(keys(raw))
	if !json.Valid([]byte(s)) {
		// A finding spanning an escape broke the encoding; fall back to
		// a string so nothing unredacted leaks.
		b, _ := json.Marshal(s)
		return b
	}
	return json.RawMessage(s)
}

// Text is New(env).Text(s).
func Text(s string, env []string) string { return New(env).Text(s) }

// JSON is New(env).JSON(raw).
func JSON(raw json.RawMessage, env []string) json.RawMessage { return New(env).JSON(raw) }

var (
	leakOnce sync.Once
	leakDet  *detect.Detector
)

// leaks returns the shared gitleaks detector with the default rules. Its
// config load uses viper's global state, so it is built once. DetectString
// keeps no findings on the detector, so sharing it is safe.
func leaks() *detect.Detector {
	leakOnce.Do(func() {
		if d, err := detect.NewDetectorDefaultConfig(); err == nil {
			leakDet = d
		}
	})
	return leakDet
}

var secretKey = regexp.MustCompile(`(?i)(secret|token|passw(or)?d|api[_-]?key|auth|credential|private[_-]?key|session[_-]?key|cookie)`)

// SecretEnvValues returns values of secret-named env vars, longest first
// so a value containing another is replaced whole. Short values are
// skipped: replacing "1" everywhere would wreck the text.
func SecretEnvValues(env []string) []string {
	var out []string
	for _, kv := range env {
		k, v, ok := strings.Cut(kv, "=")
		if ok && len(v) >= 8 && secretKey.MatchString(k) {
			out = append(out, v)
		}
	}
	slices.SortFunc(out, func(a, b string) int { return len(b) - len(a) })
	return out
}

// keys re-encodes raw JSON with values under secret-named keys replaced.
// Non-JSON input is returned as is.
func keys(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return string(raw)
	}
	b, err := json.Marshal(value(v))
	if err != nil {
		return string(raw)
	}
	return string(b)
}

func value(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			if secretKey.MatchString(k) {
				t[k] = Mask
			} else {
				t[k] = value(x)
			}
		}
	case []any:
		for i, x := range t {
			t[i] = value(x)
		}
	}
	return v
}
