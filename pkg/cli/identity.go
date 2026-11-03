package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// identityCmd exposes the controller's identities (#200).
type identityCmd struct {
	List          identityListCmd          `cmd:"" help:"List identities and who may act as them (admin)."`
	Check         identityCheckCmd         `cmd:"" help:"Show which identity covers a repo or team and obtain a credential to prove it works (admin)."`
	Token         identityTokenCmd         `cmd:"" help:"Print a token for a repo (owner/name) or Linear team key, as the identity that covers it."`
	GitCredential identityGitCredentialCmd `cmd:"" name:"git-credential" help:"git credential helper: get returns an installation token for the repo in the request. Set credential.useHttpPath=true."`
}

type identityListCmd struct{}

func (identityListCmd) Run(g *Globals) error { return g.call(rpc.MethodIdentityList, nil, nil) }

type identityCheckCmd struct {
	Target string `arg:"" help:"owner/repo or Linear team key."`
}

func (c identityCheckCmd) Run(g *Globals) error {
	return g.call(rpc.MethodIdentityCheck, rpc.IdentityParams{Target: c.Target}, nil)
}

type identityTokenCmd struct {
	Target string `arg:"" help:"owner/repo or Linear team key."`
}

func (c identityTokenCmd) Run(g *Globals) error {
	var r rpc.IdentityTokenResult
	if err := g.call(rpc.MethodIdentityToken, rpc.IdentityParams{Target: c.Target}, &r); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, r.Token)
	return err
}

type identityGitCredentialCmd struct {
	Op string `arg:"" help:"get, store or erase (git passes it)."`
}

func (c identityGitCredentialCmd) Run(g *Globals) error {
	if c.Op != "get" {
		// Tokens are minted per request; nothing to store or erase.
		_, _ = io.Copy(io.Discard, stdin)
		return nil
	}
	req := parseCredential(stdin)
	repo, ok := credentialRepo(req)
	if !ok {
		// Not ours: let git try the next helper.
		return nil
	}
	var r rpc.IdentityTokenResult
	if err := g.call(rpc.MethodIdentityToken, rpc.IdentityParams{Target: repo}, &r); err != nil {
		return err
	}
	_, err := fmt.Fprintf(stdout, "username=x-access-token\npassword=%s\n", r.Token)
	return err
}

func parseCredential(r io.Reader) map[string]string {
	m := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			break
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			m[k] = v
		}
	}
	return m
}

// credentialRepo returns owner/repo from a github.com https request. git
// sends path only with credential.useHttpPath; without it every repo
// would share one token, which cannot work across identities.
func credentialRepo(m map[string]string) (string, bool) {
	if m["protocol"] != "https" || m["host"] != "github.com" {
		return "", false
	}
	p := strings.TrimSuffix(strings.Trim(m["path"], "/"), ".git")
	parts := strings.Split(p, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	return parts[0] + "/" + strings.TrimSuffix(parts[1], ".git"), true
}
