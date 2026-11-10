package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Linear read commands (#279). Each asks the controller for the team's
// token (identity.token), so team scoping is the same as issue|comment|state.

type linearContextCmd struct {
	Key         string `arg:"" help:"Issue key, e.g. ENG-123."`
	AllComments bool   `name:"all-comments" help:"Every comment (up to 250), not the 10 newest."`
}

type linearFilterFlags struct {
	Team     string `required:"" help:"Team key, e.g. ENG."`
	State    string `help:"State name, e.g. Todo."`
	Label    string `help:"Label name."`
	Assignee string `help:"Assignee name, or me."`
	Limit    int    `default:"50" help:"Page size (max 250)."`
	After    string `help:"Cursor: end_cursor of the previous page."`
}

func (f linearFilterFlags) filter() linear.Filter {
	return linear.Filter{Team: f.Team, State: f.State, Label: f.Label, Assignee: f.Assignee, Limit: f.Limit, After: f.After}
}

type linearListCmd struct {
	linearFilterFlags
}

type linearSearchCmd struct {
	Query string `arg:"" help:"Search text."`
	linearFilterFlags
}

type linearCatalogCmd struct {
	Kind string `arg:"" enum:"projects,labels,states,users" help:"projects, labels, states or users."`
	Team string `required:"" help:"Team key, e.g. ENG."`
}

type linearEligibleCmd struct {
	Key        string   `arg:"" help:"Issue key."`
	SkipLabels []string `name:"skip-labels" sep:"," help:"Labels that make an issue ineligible (comma-separated; the pack's linear_skip_labels)."`
}

func printOut(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(b))
	return err
}

func (c linearContextCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, c.Key)
	if err != nil {
		return err
	}
	defer cancel()
	is, err := cl.GetContext(ctx, c.Key, c.AllComments)
	if err != nil {
		return err
	}
	return printOut(is)
}

// teamKey returns a key-shaped name for the team so client() can parse it.
func teamKey(team string) string { return strings.ToUpper(strings.TrimSpace(team)) + "-1" }

func (c linearListCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, teamKey(c.Team))
	if err != nil {
		return err
	}
	defer cancel()
	p, err := cl.List(ctx, c.filter())
	if err != nil {
		return err
	}
	return printOut(p)
}

func (c linearSearchCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, teamKey(c.Team))
	if err != nil {
		return err
	}
	defer cancel()
	p, err := cl.Search(ctx, c.Query, c.filter())
	if err != nil {
		return err
	}
	return printOut(p)
}

func (c linearCatalogCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, teamKey(c.Team))
	if err != nil {
		return err
	}
	defer cancel()
	es, err := cl.Catalog(ctx, strings.ToUpper(c.Team), c.Kind)
	if err != nil {
		return err
	}
	return printOut(es)
}

// errNotEligible exits 1 after the JSON answer is printed.
var errNotEligible = errors.New("not eligible")

func (c linearEligibleCmd) Run(g *Globals, l *linearCmd) error {
	cl, ctx, cancel, err := l.client(g, c.Key)
	if err != nil {
		return err
	}
	defer cancel()
	is, err := cl.GetContext(ctx, c.Key, false)
	if err != nil {
		return err
	}
	var runs []domain.Ticket
	for _, st := range []string{domain.RunRunning, domain.RunNeedsHuman} {
		var rs []domain.Ticket
		if err := g.call(rpc.MethodRunList, rpc.RunListParams{State: st}, &rs); err != nil {
			return err
		}
		runs = append(runs, rs...)
	}
	e := linear.Eligible(is, c.SkipLabels, openLinearRun(runs, is.Identifier))
	if e.Eligible {
		// Stated GitHub dependencies (JOB-68) cost a gh call each, so
		// only check them when nothing else already blocks.
		owner, _, _ := strings.Cut(originRepo(), "/")
		if err := is.CheckDeps(ctx, owner, ghIssueOpen); err != nil {
			return err
		}
		e = linear.Eligible(is, c.SkipLabels, openLinearRun(runs, is.Identifier))
	}
	if err := printOut(e); err != nil {
		return err
	}
	if !e.Eligible {
		return fmt.Errorf("%w: %s", errNotEligible, e.Reason)
	}
	return nil
}

// ghIssueOpen reads a GitHub issue's state with gh.
func ghIssueOpen(ctx context.Context, d linear.GitHubDep) (bool, error) {
	gh, err := realGH()
	if err != nil {
		return false, err
	}
	out, err := exec.CommandContext(ctx, gh, "api", fmt.Sprintf("repos/%s/issues/%d", d.Repo, d.Number), "--jq", ".state").Output()
	if err != nil {
		return false, fmt.Errorf("gh api: %w", err)
	}
	return strings.TrimSpace(string(out)) == "open", nil
}

// openLinearRun is the id of an open run whose meta.linear.issue is key.
func openLinearRun(runs []domain.Ticket, key string) string {
	for _, r := range runs {
		if r.Status == domain.StatusClosed {
			continue
		}
		var m struct {
			Linear struct {
				Issue string `json:"issue"`
			} `json:"linear"`
		}
		if json.Unmarshal(r.Meta, &m) == nil && strings.EqualFold(m.Linear.Issue, key) {
			return r.ID
		}
	}
	return ""
}
