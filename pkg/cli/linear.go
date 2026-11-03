package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// linearCmd reads and updates Linear issues as the identity that covers
// the issue's team (#201). The key never leaves the controller except as
// the token for this call.
type linearCmd struct {
	APIURL  string           `name:"api-url" hidden:"" env:"JOBSHOP_LINEAR_URL" help:"Linear GraphQL endpoint (tests)."`
	Issue   linearIssueCmd   `cmd:"" help:"Print a Linear issue with its comments as JSON."`
	Comment linearCommentCmd `cmd:"" help:"Comment on a Linear issue (e.g. the PR link)."`
	State   linearStateCmd   `cmd:"" help:"Move a Linear issue to a state of its team by name (e.g. Done on merge)."`
}

type linearIssueCmd struct {
	Key string `arg:"" help:"Issue key, e.g. ENG-123."`
}

type linearCommentCmd struct {
	Key  string `arg:"" help:"Issue key."`
	Body string `arg:"" help:"Comment text. Use - to read stdin."`
}

type linearStateCmd struct {
	Key   string `arg:"" help:"Issue key."`
	State string `arg:"" help:"State name, e.g. \"In Review\" or Done."`
}

func (l *linearCmd) client(g *Globals, key string) (*linear.Client, context.Context, context.CancelFunc, error) {
	team, ok := linear.Team(key)
	if !ok {
		return nil, nil, nil, fmt.Errorf("%q is not a Linear issue key like ENG-123", key)
	}
	var r rpc.IdentityTokenResult
	if err := g.call(rpc.MethodIdentityToken, rpc.IdentityParams{Target: team}, &r); err != nil {
		return nil, nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	return linear.New(l.APIURL, r.Token), ctx, cancel, nil
}

func (c linearIssueCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, c.Key)
	if err != nil {
		return err
	}
	defer cancel()
	is, err := cl.GetIssue(ctx, c.Key)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(is, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(b))
	return err
}

func (c linearCommentCmd) Run(g *Globals, l *linearCmd) error {
	body := c.Body
	if body == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		body = string(b)
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("empty comment")
	}
	cl, ctx, cancel, err := l.client(g, c.Key)
	if err != nil {
		return err
	}
	defer cancel()
	return cl.AddComment(ctx, c.Key, body)
}

func (c linearStateCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, c.Key)
	if err != nil {
		return err
	}
	defer cancel()
	st, err := cl.SetState(ctx, c.Key, c.State)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s: %s\n", c.Key, st.Name)
	return err
}
