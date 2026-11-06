package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Linear write commands (#280). Each takes --key: the controller stores
// key -> result, so a retry after an unknown outcome prints the stored
// record instead of posting twice. Created entities get an id derived
// from the key, so a retry whose first attempt never reported back finds
// what it created. --dry-run prints the mutation and sends nothing.

type linearWriteFlags struct {
	Key    string `help:"Idempotency key: a repeat with the same key returns the first result instead of writing again."`
	DryRun bool   `name:"dry-run" help:"Print the mutation and send nothing."`
}

// id is the entity id for a keyed create, empty without a key.
func (f linearWriteFlags) id() string {
	if f.Key == "" {
		return ""
	}
	return linear.KeyID(f.Key)
}

func (f linearWriteFlags) run(ctx context.Context, g *Globals, cl *linear.Client, m linear.Mutation) error {
	out, err := f.do(ctx, g, cl, m)
	if err != nil {
		return err
	}
	return printOut(out)
}

// do sends m under the key; out is a linear.Result, or m on --dry-run.
func (f linearWriteFlags) do(ctx context.Context, g *Globals, cl *linear.Client, m linear.Mutation) (any, error) {
	var ks linear.KeyStore
	if f.Key != "" {
		ks = rpcKeys{g}
	}
	return cl.Do(ctx, ks, f.Key, f.DryRun, m)
}

// rpcKeys stores idempotency keys on the controller.
type rpcKeys struct{ g *Globals }

func (k rpcKeys) Claim(_ context.Context, key, op string) (string, bool, bool, error) {
	var w store.WriteKey
	if err := k.g.call(rpc.MethodWriteKeyClaim, rpc.WriteKeyParams{Key: key, Op: op}, &w); err != nil {
		return "", false, false, err
	}
	return w.Result, w.Done, !w.Done && !w.New, nil
}

func (k rpcKeys) Finish(_ context.Context, key, result string) (string, error) {
	var w store.WriteKey
	if err := k.g.call(rpc.MethodWriteKeyFinish, rpc.WriteKeyParams{Key: key, Result: result}, &w); err != nil {
		return "", err
	}
	return w.Result, nil
}

type linearCreateCmd struct {
	Team        string   `required:"" help:"Team key, e.g. ENG."`
	Title       string   `required:"" help:"Issue title."`
	Description string   `help:"Issue description (markdown). Use - to read stdin."`
	State       string   `help:"State name. Without it the issue lands in the team's default state (often Backlog), and Linear intake (linear_issue_no_run) only starts runs for issues in linear_intake_states (default Todo); a note on stderr says when it lands outside them."`
	Label       []string `help:"Label name (repeatable)."`
	Assignee    string   `help:"Assignee name, display name or email; me is the token's user."`
	Parent      string   `help:"Parent issue key."`
	linearWriteFlags
}

func (c linearCreateCmd) Run(g *Globals, l *linearCmd) error {
	desc := c.Description
	if desc == "-" {
		b, err := readBody("-")
		if err != nil {
			return err
		}
		desc = b
	}
	if err := checkOut(g, "linear create", l.AllowSecretFindings, outText{"title", c.Title}, outText{"description", desc}); err != nil {
		return err
	}
	cl, ctx, cancel, err := l.client(g, teamKey(c.Team))
	if err != nil {
		return err
	}
	defer cancel()
	m, err := cl.PlanCreate(ctx, linear.CreateParams{
		Team: c.Team, Title: c.Title, Description: desc, State: c.State, Labels: c.Label, Assignee: c.Assignee, Parent: c.Parent,
	}, c.id())
	if err != nil {
		return err
	}
	out, err := c.do(ctx, g, cl, m)
	if err != nil {
		return err
	}
	if err := printOut(out); err != nil {
		return err
	}
	if res, ok := out.(linear.Result); ok && c.State == "" && !res.Replayed {
		intakeNote(g, c.Team, res)
	}
	return nil
}

// intakeNote warns on stderr when an issue created without --state landed
// outside the states Linear intake takes for its team (JOB-121): it sits
// there with no run until someone moves it. Quiet when no signal watches
// the team, when one takes any state, when the result is a replay (the
// first call noted it), or when the controller cannot say.
func intakeNote(g *Globals, team string, res linear.Result) {
	var in rpc.LinearIntakeResult
	if err := g.call(rpc.MethodLinearIntake, rpc.LinearIntakeParams{Team: team}, &in); err != nil || !in.Watched || len(in.States) == 0 {
		return
	}
	if res.State != "" && slices.ContainsFunc(in.States, func(s string) bool { return strings.EqualFold(s, res.State) }) {
		return
	}
	state := "the team's default state"
	if res.State != "" {
		state = res.State
	}
	key := res.Identifier
	if key == "" {
		key = "the issue"
	}
	warn(stderr, fmt.Sprintf("note: %s is in %s; Linear intake only takes %s (linear_intake_states), so no run starts until it moves there (jobshop linear state %s %q)",
		key, state, strings.Join(in.States, ", "), key, in.States[0]))
}

type linearUpdateCmd struct {
	IssueKey    string    `arg:"" name:"issue" help:"Issue key."`
	Title       string    `help:"New title."`
	State       string    `help:"State name."`
	Assignee    string    `help:"Assignee name, display name or email; me is the token's user."`
	AddLabel    []string  `name:"add-label" help:"Label to add (repeatable)."`
	RemoveLabel []string  `name:"remove-label" help:"Label to remove (repeatable)."`
	IfUpdatedAt time.Time `name:"if-updated-at" help:"Fail unless the issue's updated_at is this (RFC3339), so a concurrent change is not overwritten."`
	linearWriteFlags
}

func (c linearUpdateCmd) Run(g *Globals, l *linearCmd) error {
	if err := checkOut(g, "linear update", l.AllowSecretFindings, outText{"title", c.Title}); err != nil {
		return err
	}
	cl, ctx, cancel, err := l.client(g, c.IssueKey)
	if err != nil {
		return err
	}
	defer cancel()
	m, err := cl.PlanUpdate(ctx, c.IssueKey, linear.UpdateParams{
		Title: c.Title, State: c.State, Assignee: c.Assignee, AddLabels: c.AddLabel, RemoveLabels: c.RemoveLabel, IfUpdatedAt: c.IfUpdatedAt,
	})
	if err != nil {
		return err
	}
	return c.run(ctx, g, cl, m)
}

type linearAskCmd struct {
	IssueKey string   `arg:"" name:"issue" help:"Issue key."`
	Body     string   `arg:"" help:"Question text. Use - to read stdin."`
	Mention  []string `help:"User to notify: name, display name or email (repeatable). Their profile URL goes in the body, since a plain @name does not notify."`
	ReplyTo  string   `name:"reply-to" help:"Comment id to reply under."`
	linearWriteFlags
}

func (c linearAskCmd) Run(g *Globals, l *linearCmd) error {
	body, err := readBody(c.Body)
	if err != nil {
		return err
	}
	if err := checkOut(g, "linear ask", l.AllowSecretFindings, outText{"body", body}); err != nil {
		return err
	}
	cl, ctx, cancel, err := l.client(g, c.IssueKey)
	if err != nil {
		return err
	}
	defer cancel()
	m, err := cl.PlanAsk(ctx, c.IssueKey, linear.AskParams{Body: body, Mention: c.Mention, ReplyTo: c.ReplyTo}, c.id())
	if err != nil {
		return err
	}
	return c.run(ctx, g, cl, m)
}

// linearLabelCmd adds or removes labels (JOB-9), like `jobshop label add`
// for tickets: `update --add-label/--remove-label` under the name agents
// look for. An unknown label fails with the team's label names.
type linearLabelCmd struct {
	IssueKey string   `arg:"" name:"issue" help:"Issue key."`
	Op       string   `arg:"" enum:"add,remove" help:"add or remove."`
	Labels   []string `arg:"" name:"label" help:"Label names."`
	linearWriteFlags
}

func (c linearLabelCmd) Run(g *Globals, l *linearCmd) error {
	u := linearUpdateCmd{IssueKey: c.IssueKey, linearWriteFlags: c.linearWriteFlags}
	if c.Op == "add" {
		u.AddLabel = c.Labels
	} else {
		u.RemoveLabel = c.Labels
	}
	return u.Run(g, l)
}

// linearAssignCmd is update --assignee under the name agents look for.
type linearAssignCmd struct {
	IssueKey string `arg:"" name:"issue" help:"Issue key."`
	User     string `arg:"" name:"user" help:"me, or a team member's name, display name or email."`
	State    string `help:"Also move the issue to this state (e.g. In Progress)."`
	linearWriteFlags
}

func (c linearAssignCmd) Run(g *Globals, l *linearCmd) error {
	return linearUpdateCmd{IssueKey: c.IssueKey, Assignee: c.User, State: c.State, linearWriteFlags: c.linearWriteFlags}.Run(g, l)
}

type linearRelateCmd struct {
	IssueKey string `arg:"" name:"issue" help:"Issue key."`
	Blocks   string `xor:"rel" required:"" help:"Issue key this issue blocks."`
	Related  string `xor:"rel" required:"" help:"Issue key this issue is related to."`
	linearWriteFlags
}

func (c linearRelateCmd) Run(g *Globals, l *linearCmd) error {
	typ, other := linear.RelBlocks, c.Blocks
	if c.Related != "" {
		typ, other = linear.RelRelated, c.Related
	}
	cl, ctx, cancel, err := l.client(g, c.IssueKey)
	if err != nil {
		return err
	}
	defer cancel()
	m, err := cl.PlanRelate(ctx, c.IssueKey, typ, strings.ToUpper(other), c.id())
	if err != nil {
		return err
	}
	return c.run(ctx, g, cl, m)
}
