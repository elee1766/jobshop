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
	// IdentityLinearApp is a Linear OAuth app that mints its own tokens
	// with client_credentials (#278).
	IdentityLinearApp = "linear-app"
	// IdentityGitHubToken is a personal access token, for running jobshop
	// as a user instead of an App (#340).
	IdentityGitHubToken = "github-token"
)

// IsGitHub reports whether kind is a GitHub identity (App or token).
func IsGitHub(kind string) bool { return kind == IdentityGitHubApp || kind == IdentityGitHubToken }

// DefaultLinearAppScopes are requested when a linear-app sets none.
var DefaultLinearAppScopes = []string{"read", "write", "issues:create", "comments:create"}

// IsLinear reports whether kind is a Linear identity (key or app).
func IsLinear(kind string) bool { return kind == IdentityLinear || kind == IdentityLinearApp }

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
	// GitHub token (#340): TokenFile holds the PAT; Repos and APIURL as
	// for an App, no installations.
	// Linear: TokenFile holds the API key, Workspace names the workspace,
	// Teams are the team keys covered ("*" for all).
	TokenFile string   `yaml:"token_file"`
	Workspace string   `yaml:"workspace"`
	Teams     []string `yaml:"teams"`
	// Linear App (#278): ClientIDFile and ClientSecretFile hold the OAuth
	// app credentials; Scopes default to DefaultLinearAppScopes. TokenURL
	// defaults to Linear's token endpoint (tests point it at a fake).
	ClientIDFile     string   `yaml:"client_id_file"`
	ClientSecretFile string   `yaml:"client_secret_file"`
	Scopes           []string `yaml:"scopes"`
	TokenURL         string   `yaml:"token_url"`
	// WebhookSecretFile holds the Linear webhook signing secret (#300):
	// the App's, or a workspace webhook's for kind linear (#340).
	WebhookSecretFile string `yaml:"webhook_secret_file"`
}

// LinearScopes is Scopes or the default set.
func (i Identity) LinearScopes() []string {
	if len(i.Scopes) == 0 {
		return DefaultLinearAppScopes
	}
	return i.Scopes
}

// CoversRepo reports whether a GitHub identity covers owner/repo and, for
// an App, has an installation for its owner.
func (i Identity) CoversRepo(repo string) bool {
	if !IsGitHub(i.Kind) || !RepoPattern.MatchString(repo) {
		return false
	}
	owner, _, _ := strings.Cut(repo, "/")
	if _, ok := i.Installations[owner]; !ok && i.Kind == IdentityGitHubApp {
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
	if !IsLinear(i.Kind) || team == "" {
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
		if err := checkRepoGlobs(i.Repos); err != nil {
			return err
		}
		if i.TokenFile != "" || i.Workspace != "" || len(i.Teams) > 0 || i.linearAppSet() || i.WebhookSecretFile != "" {
			return errors.New("token_file, workspace, teams, webhook_secret_file and client_* are not for kind github-app")
		}
	case IdentityGitHubToken:
		if i.TokenFile == "" || len(i.Repos) == 0 {
			return errors.New("token_file and repos are required")
		}
		if err := checkRepoGlobs(i.Repos); err != nil {
			return err
		}
		if i.AppID != 0 || i.PrivateKeyFile != "" || len(i.Installations) > 0 {
			return errors.New("app_id, private_key_file and installations are for kind github-app")
		}
		if i.Workspace != "" || len(i.Teams) > 0 || i.linearAppSet() || i.WebhookSecretFile != "" {
			return errors.New("workspace, teams, webhook_secret_file and client_* are for Linear kinds")
		}
	case IdentityLinear:
		if i.TokenFile == "" || i.Workspace == "" || len(i.Teams) == 0 {
			return errors.New("token_file, workspace and teams are required")
		}
		if i.githubSet() {
			return errors.New("app_id, private_key_file, installations, repos and api_url are for GitHub kinds")
		}
		if i.linearAppSet() {
			return errors.New("client_id_file, client_secret_file, scopes and token_url are for kind linear-app")
		}
	case IdentityLinearApp:
		if i.ClientIDFile == "" || i.ClientSecretFile == "" || i.Workspace == "" || len(i.Teams) == 0 {
			return errors.New("client_id_file, client_secret_file, workspace and teams are required")
		}
		if i.TokenFile != "" {
			return errors.New("token_file is for kind linear; linear-app mints its own token")
		}
		if i.githubSet() {
			return errors.New("app_id, private_key_file, installations, repos and api_url are for kind github-app")
		}
		for _, s := range i.Scopes {
			if s == "" || strings.ContainsAny(s, ", ") {
				return fmt.Errorf("scopes: bad scope %q (one per list item)", s)
			}
		}
	default:
		return fmt.Errorf("kind %q must be %q, %q, %q or %q", i.Kind, IdentityGitHubApp, IdentityGitHubToken, IdentityLinear, IdentityLinearApp)
	}
	return nil
}

func checkRepoGlobs(repos []string) error {
	for _, g := range repos {
		if _, err := path.Match(g, ""); err != nil {
			return fmt.Errorf("repos: bad glob %q", g)
		}
	}
	return nil
}

func (i Identity) githubSet() bool {
	return i.AppID != 0 || i.PrivateKeyFile != "" || len(i.Installations) > 0 || len(i.Repos) > 0 || i.APIURL != ""
}

func (i Identity) linearAppSet() bool {
	return i.ClientIDFile != "" || i.ClientSecretFile != "" || len(i.Scopes) > 0 || i.TokenURL != ""
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
