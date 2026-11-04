package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
	var runs []store.Ticket
	for _, st := range []string{store.RunRunning, store.RunNeedsHuman} {
		var rs []store.Ticket
		if err := g.call(rpc.MethodRunList, rpc.RunListParams{State: st}, &rs); err != nil {
			return err
		}
		runs = append(runs, rs...)
	}
	e := linear.Eligible(is, c.SkipLabels, openLinearRun(runs, is.Identifier))
	if err := printOut(e); err != nil {
		return err
	}
	if !e.Eligible {
		return fmt.Errorf("%w: %s", errNotEligible, e.Reason)
	}
	return nil
}

// openLinearRun is the id of an open run whose meta.linear.issue is key.
func openLinearRun(runs []store.Ticket, key string) string {
	for _, r := range runs {
		if r.Status == store.StatusClosed {
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
