package config

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
)

// Identity kinds.
const (
	IdentityGitHubApp = "github-app"
	IdentityLinear    = "linear"
)

// Identity is a credential jobshop acts as, and what it may act on (#200).
// Secrets come only from files; the yaml holds paths.
type Identity struct {
	Kind string `yaml:"kind"`
	// GitHub App: AppID, PrivateKeyFile (PEM), Installations maps an
	// owner (user or org) to its installation id, Repos are owner/repo
	// globs (path.Match) the identity covers. APIURL defaults to
	// https://api.github.com (tests point it at a fake).
	AppID          int64            `yaml:"app_id"`
	PrivateKeyFile string           `yaml:"private_key_file"`
	Installations  map[string]int64 `yaml:"installations"`
	Repos          []string         `yaml:"repos"`
	APIURL         string           `yaml:"api_url"`
	// Linear: TokenFile holds the API key, Workspace names the workspace,
	// Teams are the team keys covered ("*" for all).
	TokenFile string   `yaml:"token_file"`
	Workspace string   `yaml:"workspace"`
	Teams     []string `yaml:"teams"`
}

// CoversRepo reports whether a GitHub App identity covers owner/repo and
// has an installation for its owner.
func (i Identity) CoversRepo(repo string) bool {
	if i.Kind != IdentityGitHubApp || !RepoPattern.MatchString(repo) {
		return false
	}
	owner, _, _ := strings.Cut(repo, "/")
	if _, ok := i.Installations[owner]; !ok {
		return false
	}
	for _, g := range i.Repos {
		if ok, _ := path.Match(g, repo); ok {
			return true
		}
	}
	return false
}

// CoversTeam reports whether a Linear identity covers the team key.
func (i Identity) CoversTeam(team string) bool {
	if i.Kind != IdentityLinear || team == "" {
		return false
	}
	return slices.Contains(i.Teams, team) || slices.Contains(i.Teams, "*")
}

func checkIdentity(i Identity) error {
	switch i.Kind {
	case IdentityGitHubApp:
		if i.AppID <= 0 {
			return errors.New("app_id is required")
		}
		if i.PrivateKeyFile == "" {
			return errors.New("private_key_file is required")
		}
		if len(i.Installations) == 0 || len(i.Repos) == 0 {
			return errors.New("installations and repos are required")
		}
		for _, g := range i.Repos {
			if _, err := path.Match(g, ""); err != nil {
				return fmt.Errorf("repos: bad glob %q", g)
			}
		}
		if i.TokenFile != "" || i.Workspace != "" || len(i.Teams) > 0 {
			return errors.New("token_file, workspace and teams are for kind linear")
		}
	case IdentityLinear:
		if i.TokenFile == "" || i.Workspace == "" || len(i.Teams) == 0 {
			return errors.New("token_file, workspace and teams are required")
		}
		if i.AppID != 0 || i.PrivateKeyFile != "" || len(i.Installations) > 0 || len(i.Repos) > 0 || i.APIURL != "" {
			return errors.New("app_id, private_key_file, installations, repos and api_url are for kind github-app")
		}
	default:
		return fmt.Errorf("kind %q must be %q or %q", i.Kind, IdentityGitHubApp, IdentityLinear)
	}
	return nil
}

// IdentityFor returns the first identity, in sorted name order, among
// names that covers target: an owner/repo for GitHub, a team key for
// Linear. No match is "", false.
func (c *Config) IdentityFor(names []string, target string) (string, bool) {
	names = slices.Clone(names)
	sort.Strings(names)
	for _, n := range names {
		i, ok := c.Identities[n]
		if !ok {
			continue
		}
		if i.CoversRepo(target) || i.CoversTeam(target) {
			return n, true
		}
	}
	return "", false
}

// IdentityNames lists every configured identity, sorted.
func (c *Config) IdentityNames() []string {
	out := make([]string, 0, len(c.Identities))
	for n := range c.Identities {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
