package control

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Intake hooks (#288): Linear issue text is untrusted. Before it reaches
// a prompt (a run started for a Linear issue, a Linear comment delivered
// to a subscriber) the on: intake hooks read it. Anything but allow keeps
// it out: the run is not started or the comment not delivered, the Linear
// issue gets the needs-human label and a needs-human ticket lands on the
// attention list. Closing that ticket done approves the issue's text.

// Intake event sources.
const (
	intakeRunStart = "run_start"
	intakeComment  = "comment"
)

// intakeExtra is hooks.GuardEvent.Extra for on: intake.
type intakeExtra struct {
	Source string `json:"source"`
	Issue  string `json:"issue"`
	Author string `json:"author,omitempty"`
	Text   string `json:"text"`
}

func hasIntakeHooks(cfg *config.Config) bool {
	return slices.ContainsFunc(cfg.Hooks, func(h config.Hook) bool { return h.On == config.HookIntake })
}

// intake runs the intake hooks on x and handles a refusal. key names the
// decision so a human's answer on the filed ticket is found again. It
// returns the answer; allow means the text may reach a prompt.
func (c *Controller) intake(ctx context.Context, cfg *config.Config, project, key string, x intakeExtra) hooks.Answer {
	if r, ok := c.askedHuman(ctx, key); ok {
		return hooks.Answer{Decision: r.Decision, Reason: r.Reason}
	}
	raw, err := json.Marshal(x)
	if err != nil {
		return hooks.Answer{Decision: rpc.GuardDeny, Reason: "intake event: " + err.Error()}
	}
	a := c.runHooks(ctx, cfg, hooks.GuardEvent{On: config.HookIntake, Target: x.Issue, Project: project, Extra: raw})
	if a.Decision != rpc.GuardAllow {
		c.refuseIntake(ctx, cfg, project, key, x, a)
	}
	return a
}

// intakeIssue checks a Linear issue's title, description and comments
// before a run starts for it. A read failure is deny (fail closed).
func (c *Controller) intakeIssue(ctx context.Context, cfg *config.Config, project, key string) hooks.Answer {
	x := intakeExtra{Source: intakeRunStart, Issue: key}
	is, err := c.linearIssue(ctx, cfg, key)
	if err != nil {
		a := hooks.Answer{Decision: rpc.GuardDeny, Reason: "intake: read " + key + ": " + err.Error()}
		c.refuseIntake(ctx, cfg, project, "intake/"+key, x, a)
		return a
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n%s\n", is.Title, is.Description)
	for _, cm := range is.Comments {
		if cm.Self {
			continue
		}
		fmt.Fprintf(&b, "\n[%s] %s:\n%s\n", cm.CreatedAt.UTC().Format(time.RFC3339), cm.Author, cm.Body)
	}
	x.Text = b.String()
	return c.intake(ctx, cfg, project, "intake/"+key, x)
}

// intakeRun gates a run start for a Linear workflow. Started from a pass
// (a trigger), the Linear read and the hooks run outside the tick like a
// check (JOB-103): the start returns triggers.ErrCheckPending and the replay has
// the answer.
func (c *Controller) intakeRun(ctx context.Context, cfg *config.Config, plan runPlan) error {
	if plan.w.Tracker != config.TrackerLinear || !hasIntakeHooks(cfg) {
		return nil
	}
	key := metaString(plan.meta, "linear", "issue")
	res, err := triggers.OffTick(ctx, triggers.Need{
		Key: "intake\x00" + plan.proj + "\x00" + key, Poller: "intake", Timeout: intakeTimeout,
		Run: func(ctx context.Context) triggers.Result {
			return triggers.Result{Val: c.intakeIssue(ctx, cfg, plan.proj, key)}
		},
	})
	if err != nil {
		return err
	}
	a, _ := res.Val.(hooks.Answer)
	if a.Decision == rpc.GuardAllow {
		return nil
	}
	return forbidden("intake hook %s for %s: %s", a.Decision, key, a.Reason)
}

// intakeTimeout bounds an intake check: the Linear read and the hooks,
// each with its own timeout.
const intakeTimeout = 5 * time.Minute

// linearClient returns the shared client of the identity covering key's team.
func (c *Controller) linearClient(key string) (*linear.Client, error) {
	return c.linearClientIn(c.config(), key)
}

// linearClientIn is linearClient under cfg.
func (c *Controller) linearClientIn(cfg *config.Config, key string) (*linear.Client, error) {
	team, ok := linear.Team(key)
	if !ok {
		return nil, fmt.Errorf("bad Linear issue key %q", key)
	}
	return c.linearForTeam(cfg, team)
}

func (c *Controller) linearIssue(ctx context.Context, cfg *config.Config, key string) (linear.Issue, error) {
	cl, err := c.linearClientIn(cfg, key)
	if err != nil {
		return linear.Issue{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return cl.GetIssue(ctx, key)
}

// refuseIntake labels the Linear issue needs-human and files one
// needs-human ticket per key; ask_human also labels it a question.
func (c *Controller) refuseIntake(ctx context.Context, cfg *config.Config, project, key string, x intakeExtra, a hooks.Answer) {
	c.log.InfoContext(ctx, "intake refused", "issue", x.Issue, "source", x.Source, "decision", a.Decision, "reason", a.Reason)
	if cl, err := c.linearClientIn(cfg, x.Issue); err != nil {
		c.log.WarnContext(ctx, "intake: label", "issue", x.Issue, "err", err)
	} else {
		lctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		err := cl.AddLabel(lctx, x.Issue, needsHumanLabel)
		cancel()
		if err != nil {
			c.log.WarnContext(ctx, "intake: label", "issue", x.Issue, "err", err)
		}
	}
	open, err := c.store.List(ctx, store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: key, Order: store.OrderNewest, Limit: 1})
	if err == nil && len(open) > 0 && open[0].Status != store.StatusClosed {
		return
	}
	meta, err := json.Marshal(map[string]map[string]string{hookMetaPack: {"key": key}, "linear": {"issue": x.Issue}})
	if err != nil {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "An intake hook refused Linear issue %s text (%s): %s\n\nReason: %s\n", x.Issue, x.Source, a.Decision, a.Reason)
	if x.Author != "" {
		fmt.Fprintf(&b, "Author: %s\n", x.Author)
	}
	labels := []string{needsHumanLabel}
	if a.Decision == rpc.GuardAskHuman {
		labels = append(labels, questionLabel)
		b.WriteString("\nQuestion: may this text reach an agent prompt?")
	}
	b.WriteString("\nClose done to allow it, close with any other resolution to keep it out.")
	if project == "" {
		project = cfg.DefaultProject
	}
	prio := 1
	if _, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: store.KindTask, Title: "Intake refused for " + x.Issue,
		Body: b.String(), Priority: &prio, Labels: labels, Meta: meta,
	}); err != nil {
		c.log.WarnContext(ctx, "intake: file ticket", "issue", x.Issue, "err", err)
	}
}
