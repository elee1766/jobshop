package config

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// PoolRepo is one keyed repo of a pool (JOB-207): the checkout its
// sessions take worktrees of, and the commit-ish they start from
// (default: the pool's ref).
type PoolRepo struct {
	Checkout string `yaml:"checkout"`
	Ref      string `yaml:"ref"`
}

// checkRepos validates Repos: owner/name keys, nonblank checkouts, and a
// default Repo for tickets with no keyed repo.
func (p Pool) checkRepos() error {
	if len(p.Repos) == 0 {
		return nil
	}
	if p.Repo == "" {
		return errors.New("repos needs repo, the checkout for tickets of any other repo")
	}
	if p.Chat() {
		return errors.New("repos does not apply to chat pools")
	}
	seen := map[string]string{}
	for _, k := range p.RepoKeys() {
		if !RepoPattern.MatchString(k) {
			return fmt.Errorf("repos: key %q must be owner/name", k)
		}
		if strings.TrimSpace(p.Repos[k].Checkout) == "" {
			return fmt.Errorf("repos: %s: checkout is required", k)
		}
		low := strings.ToLower(k)
		if other, dup := seen[low]; dup {
			return fmt.Errorf("repos: %s and %s differ only in case", other, k)
		}
		seen[low] = k
	}
	return nil
}

// RepoKeys are the pool's keyed repos, sorted.
func (p Pool) RepoKeys() []string { return slices.Sorted(maps.Keys(p.Repos)) }

// RepoKey is the Repos key matching repo (GitHub names are not case
// sensitive), or "" when the pool has none: such a ticket goes to the
// pool's default sessions.
func (p Pool) RepoKey(repo string) string {
	if repo == "" {
		return ""
	}
	if _, ok := p.Repos[repo]; ok {
		return repo
	}
	for k := range p.Repos {
		if strings.EqualFold(k, repo) {
			return k
		}
	}
	return ""
}

// Checkout is the git checkout and ref a session keyed by key works
// from: its Repos entry, or Repo and Ref for the default key "" and for
// a key the pool no longer has.
func (p Pool) Checkout(key string) (repo, ref string) {
	if r, ok := p.Repos[key]; ok && key != "" {
		return r.Checkout, cmp.Or(r.Ref, p.Ref)
	}
	return p.Repo, p.Ref
}
