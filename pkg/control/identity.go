package control

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Identities (#200). The controller holds the App keys and Linear token
// files and hands sessions short-lived tokens for the repo or team they
// ask about, only through identities their project or pool names.

// identitiesFor lists the identities the caller may act as: all for the
// admin socket, else the session's project plus pool identities.
func (c *Controller) identitiesFor(cfg *config.Config, sc scope) []string {
	if sc.admin {
		return cfg.IdentityNames()
	}
	var out []string
	if p, ok := cfg.Projects[sc.project]; ok {
		out = append(out, p.Identities...)
	}
	if p, ok := cfg.Pools[sc.pool]; ok {
		out = append(out, p.Identities...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// resolveIdentity picks the identity covering target. No match is a
// forbidden error naming the target, so an agent escalates instead of
// retrying.
func (c *Controller) resolveIdentity(cfg *config.Config, sc scope, target string) (string, config.Identity, error) {
	if target == "" {
		return "", config.Identity{}, forbidden("identity: no target given")
	}
	names := c.identitiesFor(cfg, sc)
	name, ok := cfg.IdentityFor(names, target)
	if !ok {
		who := "this session"
		if sc.admin {
			who = "jobshop"
		}
		return "", config.Identity{}, forbidden("no identity covers %s for %s (allowed: %v); add it to an identity's repos or teams and the project's identities", target, who, names)
	}
	return name, cfg.Identities[name], nil
}

func (c *Controller) identityList(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := requireAdmin(sc, "identity list"); err != nil {
		return nil, err
	}
	cfg := c.config()
	out := []rpc.IdentityInfo{}
	for _, n := range cfg.IdentityNames() {
		i := cfg.Identities[n]
		info := rpc.IdentityInfo{Name: n, Kind: i.Kind, AppID: i.AppID, Repos: i.Repos, Workspace: i.Workspace, Teams: i.Teams}
		for o := range i.Installations {
			info.Owners = append(info.Owners, o)
		}
		sort.Strings(info.Owners)
		for pn, p := range cfg.Projects {
			if slices.Contains(p.Identities, n) {
				info.Projects = append(info.Projects, pn)
			}
		}
		sort.Strings(info.Projects)
		for pn, p := range cfg.Pools {
			if slices.Contains(p.Identities, n) {
				info.Pools = append(info.Pools, pn)
			}
		}
		sort.Strings(info.Pools)
		out = append(out, info)
	}
	return out, nil
}

func (c *Controller) identityCheck(ctx context.Context, sc scope, p rpc.IdentityParams) (any, error) {
	if err := requireAdmin(sc, "identity check"); err != nil {
		return nil, err
	}
	cfg := c.config()
	name, id, err := c.resolveIdentity(cfg, sc, p.Target)
	if err != nil {
		return nil, err
	}
	t, err := c.minter.Token(ctx, name, id, p.Target)
	if err != nil {
		return nil, err
	}
	res := rpc.IdentityCheckResult{Target: p.Target, Identity: name, Kind: id.Kind, Expires: t.Expires}
	if config.IsLinear(id.Kind) {
		// A minted token can still be refused by the API; reading viewer
		// proves it works (#278).
		v, err := linear.New(c.linearURL, t.Token).Viewer(ctx)
		if err != nil {
			return nil, fmt.Errorf("identity %s: %w", name, err)
		}
		res.Viewer = v.Name
	}
	if id.Kind == config.IdentityGitHubToken {
		l, err := c.minter.Login(ctx, name, id)
		if err != nil {
			return nil, err
		}
		res.Viewer = l
	}
	return res, nil
}

func (c *Controller) identityToken(ctx context.Context, sc scope, p rpc.IdentityParams) (any, error) {
	cfg := c.config()
	name, id, err := c.resolveIdentity(cfg, sc, p.Target)
	if err != nil {
		return nil, err
	}
	t, err := c.minter.Token(ctx, name, id, p.Target)
	if err != nil {
		return nil, fmt.Errorf("identity token for %s: %w", p.Target, err)
	}
	c.log.InfoContext(ctx, "identity token", "identity", name, "target", p.Target, "actor", sc.actor)
	return rpc.IdentityTokenResult{Identity: name, Kind: id.Kind, Token: t.Token, Expires: t.Expires}, nil
}
