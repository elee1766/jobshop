package store

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

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
func ValidLearning(l *Learning) error {
	l.Summary = strings.TrimSpace(l.Summary)
	l.Fix = strings.TrimSpace(l.Fix)
	if !slices.Contains(LearningKinds, l.Kind) {
		return fmt.Errorf("%w: learning kind %q is not one of %s", ErrInvalid, l.Kind, strings.Join(LearningKinds, ", "))
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
