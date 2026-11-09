package domain

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Learning kinds, sources and statuses (#264).
const (
	LearningCorrection = "correction"
	LearningFailure    = "failure"
	LearningWorkaround = "workaround"

	LearningSourceAgent      = "agent"
	LearningSourceController = "controller"

	LearningOpen     = "open"
	LearningApplied  = "applied"
	LearningRejected = "rejected"
)

// LearningKinds are the default learning kinds, used when the config
// sets no learnings.kinds (#524). The controller's own learnings use
// correction and failure and are always allowed.
var LearningKinds = []string{LearningCorrection, LearningFailure, LearningWorkaround}

// Learning is a correction, failure or workaround recorded for the curator.
// Scope is "global", "repo:<owner/name>" or "pool:<name>".
type Learning struct {
	ID      int64  `json:"id"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Fix     string `json:"fix,omitempty"`
	Scope   string `json:"scope"`
	Ticket  string `json:"ticket,omitempty"`
	Session string `json:"session,omitempty"`
	Source  string `json:"source"`
	Status  string `json:"status"`
	// PR is the curator's PR that addresses it (#273): pending while
	// open, kept when applied or rejected.
	PR string `json:"pr,omitempty"`
	// Gate is the curator gate's result on the edit that cited it
	// (#274): "helped: ..." or "not_helped: ...".
	Gate      string    `json:"gate,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Learning scopes (#264).
const (
	ScopeGlobal     = "global"
	scopeRepoPrefix = "repo:"
	scopePoolPrefix = "pool:"
)

var (
	repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	poolPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

// ParseScope validates and normalizes a learning scope: "global",
// "repo:<owner/name>" or "pool:<name>". Empty means global.
func ParseScope(s string) (string, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "" || s == ScopeGlobal:
		return ScopeGlobal, nil
	case strings.HasPrefix(s, scopeRepoPrefix):
		repo := strings.TrimPrefix(s, scopeRepoPrefix)
		if !repoPattern.MatchString(repo) {
			return "", fmt.Errorf("%w: scope %q: repo must be owner/name", ErrInvalid, s)
		}
		return scopeRepoPrefix + repo, nil
	case strings.HasPrefix(s, scopePoolPrefix):
		pool := strings.TrimPrefix(s, scopePoolPrefix)
		if !poolPattern.MatchString(pool) {
			return "", fmt.Errorf("%w: scope %q: bad pool name", ErrInvalid, s)
		}
		return s, nil
	}
	return "", fmt.Errorf("%w: scope %q is not global, repo:<owner/name> or pool:<name>", ErrInvalid, s)
}

// RepoScope is the scope of one repo (owner/name).
func RepoScope(repo string) string { return scopeRepoPrefix + repo }

// PoolScope is the scope of one pool.
func PoolScope(pool string) string { return scopePoolPrefix + pool }

// ValidLearning checks a learning before it is stored and fills defaults.
// The kind must be a valid name; whether the config allows it is
// CheckLearningKind, run by the controller on agent learnings (#524).
func ValidLearning(l *Learning) error {
	l.Summary = strings.TrimSpace(l.Summary)
	l.Fix = strings.TrimSpace(l.Fix)
	l.Kind = strings.TrimSpace(l.Kind)
	if !kindName.MatchString(l.Kind) {
		return fmt.Errorf("%w: learning kind %q: want a name like %s", ErrInvalid, l.Kind, kindName)
	}
	if l.Summary == "" {
		return fmt.Errorf("%w: learning summary is required", ErrInvalid)
	}
	if l.Source != LearningSourceAgent && l.Source != LearningSourceController {
		return fmt.Errorf("%w: learning source %q", ErrInvalid, l.Source)
	}
	if l.Status == "" {
		l.Status = LearningOpen
	}
	if l.Status != LearningOpen && l.Status != LearningApplied && l.Status != LearningRejected {
		return fmt.Errorf("%w: learning status %q", ErrInvalid, l.Status)
	}
	sc, err := ParseScope(l.Scope)
	if err != nil {
		return err
	}
	l.Scope = sc
	return nil
}

// kindName is what a learning kind looks like (#524).
var kindName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// kindPatternChars are name chars and the glob chars that can stand for
// them; a pattern with anything else can match no kind.
var kindPatternChars = regexp.MustCompile(`^[a-z*?\[][a-z0-9*?\[\]^!-]*$`)

// ValidKindPattern checks a learnings.kinds entry: a kind name or a glob
// that can match one.
func ValidKindPattern(p string) error {
	if _, err := path.Match(p, ""); err != nil {
		return fmt.Errorf("learning kind pattern %q: %w", p, err)
	}
	if !kindPatternChars.MatchString(p) {
		return fmt.Errorf("learning kind pattern %q can match no kind (names are %s)", p, kindName)
	}
	return nil
}

// CheckLearningKind reports whether kind matches one of patterns (empty:
// the default LearningKinds). correction and failure are always allowed:
// the controller records them itself.
func CheckLearningKind(kind string, patterns []string) error {
	if kind == LearningCorrection || kind == LearningFailure {
		return nil
	}
	if len(patterns) == 0 {
		patterns = LearningKinds
	}
	if MatchAny(patterns, kind) {
		return nil
	}
	return fmt.Errorf("%w: learning kind %q matches none of learnings.kinds: %s", ErrInvalid, kind, strings.Join(patterns, ", "))
}

// Glob matches s against a path.Match pattern where * and ? also match
// "/", so repo:* matches repo:owner/name.
func Glob(pattern, s string) bool {
	const sep = "\x1f"
	ok, err := path.Match(strings.ReplaceAll(pattern, "/", sep), strings.ReplaceAll(s, "/", sep))
	return err == nil && ok
}

// MatchAny reports whether s matches one of patterns.
func MatchAny(patterns []string, s string) bool {
	return slices.ContainsFunc(patterns, func(p string) bool { return Glob(p, s) })
}

// LearningCount selects learnings to count (#524). Empty fields match
// everything; set fields AND.
type LearningCount struct {
	Status string
	Kinds  []string // globs
	Scopes []string // globs over "global", "repo:<owner/name>", "pool:<name>"
	Source string   // agent or controller
}

// Matches reports whether a learning of kind, scope and source counts.
func (f LearningCount) Matches(kind, scope, source string) bool {
	if len(f.Kinds) > 0 && !MatchAny(f.Kinds, kind) {
		return false
	}
	if len(f.Scopes) > 0 && !MatchAny(f.Scopes, scope) {
		return false
	}
	return f.Source == "" || f.Source == source
}
