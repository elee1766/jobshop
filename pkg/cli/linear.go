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
	APIURL string `name:"api-url" hidden:"" env:"JOBSHOP_LINEAR_URL" help:"Linear GraphQL endpoint (tests)."`
	// AllowSecretFindings skips the outbound secret check (#574); admin only.
	AllowSecretFindings bool              `name:"allow-secret-findings" help:"Send writes that hold a secret finding (admin callers only)."`
	Issue               linearIssueCmd    `cmd:"" help:"Print a Linear issue with its comments as JSON."`
	Comment             linearCommentCmd  `cmd:"" help:"Comment on a Linear issue (e.g. the PR link)."`
	State               linearStateCmd    `cmd:"" help:"Move a Linear issue to a state of its team by name (e.g. Done on merge)."`
	Context             linearContextCmd  `cmd:"" help:"Print an issue with labels, blockers, relations, children, assignee, state and recent comments as JSON."`
	List                linearListCmd     `cmd:"" help:"List a team's issues as JSON, one page at a time."`
	Search              linearSearchCmd   `cmd:"" help:"Search a team's issues as JSON, one page at a time."`
	Catalog             linearCatalogCmd  `cmd:"" help:"List a team's projects, labels, states or users as JSON."`
	Eligible            linearEligibleCmd `cmd:"" help:"Intake gate: print whether an issue may start a run and why not. Exits 1 when not eligible."`
	Create              linearCreateCmd   `cmd:"" help:"Create an issue. --key makes a retry return the first result."`
	Update              linearUpdateCmd   `cmd:"" help:"Change title, state, assignee or labels; reads the issue back to verify."`
	Assign              linearAssignCmd   `cmd:"" help:"Take an issue: assign it to a user, or to yourself with me (an app token becomes the issue's delegate). linear assign <KEY> me."`
	Label               linearLabelCmd    `cmd:"" help:"Add or remove labels: linear label <KEY> add|remove <label>... (same as update --add-label/--remove-label)."`
	Ask                 linearAskCmd      `cmd:"" help:"Comment a question, mentioning users so they are notified."`
	Relate              linearRelateCmd   `cmd:"" help:"Mark an issue as blocking or related to another."`
	API                 linearAPICmd      `cmd:"" name:"api" help:"Raw GraphQL call: LINEAR_API_KEY if set, else the team's identity. Mutations ask the guard (linear-api-mutation)."`
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
	if err := checkOut(g, "linear comment", l.AllowSecretFindings, outText{"body", body}); err != nil {
		return err
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
