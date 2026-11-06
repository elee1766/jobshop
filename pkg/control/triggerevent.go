package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Event triggers (JOB-62, docs/notes/JOB-62-event-triggers.md). A
// trigger with on: fires when an admitted webhook delivery matches,
// without waiting for the next tick; its when conditions and dedup are
// the same as a poll's. fallback: spaces its polls, so a lost webhook only
// costs latency.

// triggerEvent is the fields of one delivery a trigger can match and
// template on, keyed as config.EventFields.
type triggerEvent map[string]string

// eventOf reads the matchable fields of an inbox item. Labels are
// joined with newlines in "label".
func eventOf(it store.InboxItem) triggerEvent {
	ev := triggerEvent{"source": it.Source, "event": it.Kind, "repo": it.RepoOrTeam, "delivery": it.DeliveryID}
	if it.Number > 0 {
		ev["pr"] = fmt.Sprint(it.Number)
	}
	payload := it.Admitted
	if len(payload) == 0 || string(payload) == "null" {
		payload = it.Payload
	}
	switch it.Source {
	case ingest.SourceGitHub:
		githubEvent(ev, payload)
	case ingest.SourceLinear:
		var le ingest.LinearEvent
		if json.Unmarshal(payload, &le) == nil {
			ev["event"], ev["action"] = le.Type, le.Action
		}
	}
	return ev
}

type ghEventRun struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	Conclusion string `json:"conclusion"`
}

type ghLabel struct {
	Name string `json:"name"`
}

func githubEvent(ev triggerEvent, payload json.RawMessage) {
	var p struct {
		Action      string      `json:"action"`
		Ref         string      `json:"ref"`
		After       string      `json:"after"`
		WorkflowRun *ghEventRun `json:"workflow_run"`
		CheckSuite  *ghEventRun `json:"check_suite"`
		CheckRun    *struct {
			ghEventRun
			CheckSuite *ghEventRun `json:"check_suite"`
		} `json:"check_run"`
		PullRequest *struct {
			Head struct {
				Ref string `json:"ref"`
				SHA string `json:"sha"`
			} `json:"head"`
			Labels []ghLabel `json:"labels"`
		} `json:"pull_request"`
		Issue *struct {
			Labels []ghLabel `json:"labels"`
		} `json:"issue"`
		Label *ghLabel `json:"label"`
	}
	if json.Unmarshal(payload, &p) != nil {
		return
	}
	ev["action"] = p.Action
	set := func(k, v string) {
		if v != "" && ev[k] == "" {
			ev[k] = v
		}
	}
	if r := p.WorkflowRun; r != nil {
		set("workflow", r.Name)
		if r.Path != "" {
			ev["workflow_file"] = strings.TrimSuffix(path.Base(r.Path), path.Ext(r.Path))
		}
		set("branch", r.HeadBranch)
		set("head_sha", r.HeadSHA)
		set("conclusion", r.Conclusion)
	}
	if r := p.CheckRun; r != nil {
		set("workflow", r.Name)
		set("head_sha", r.HeadSHA)
		set("conclusion", r.Conclusion)
		if r.CheckSuite != nil {
			set("branch", r.CheckSuite.HeadBranch)
		}
	}
	if r := p.CheckSuite; r != nil {
		set("branch", r.HeadBranch)
		set("head_sha", r.HeadSHA)
		set("conclusion", r.Conclusion)
	}
	if pr := p.PullRequest; pr != nil {
		set("branch", pr.Head.Ref)
		set("head_sha", pr.Head.SHA)
	}
	if b, ok := strings.CutPrefix(p.Ref, "refs/heads/"); ok {
		set("branch", b)
	}
	set("head_sha", p.After)
	var labels []string
	if p.Label != nil {
		labels = append(labels, p.Label.Name)
	}
	if p.PullRequest != nil {
		for _, l := range p.PullRequest.Labels {
			labels = append(labels, l.Name)
		}
	}
	if p.Issue != nil {
		for _, l := range p.Issue.Labels {
			labels = append(labels, l.Name)
		}
	}
	ev["label"] = strings.Join(labels, "\n")
}

// matches reports whether ev has every field m sets.
func (ev triggerEvent) matches(source string, m config.EventMatch) bool {
	if ev["source"] != source {
		return false
	}
	eq := func(want, got string) bool { return want == "" || strings.EqualFold(want, got) }
	if !eq(m.Event, ev["event"]) || !eq(m.Action, ev["action"]) || !eq(m.Repo, ev["repo"]) ||
		!eq(m.Branch, ev["branch"]) || !eq(m.Conclusion, ev["conclusion"]) {
		return false
	}
	if m.Workflow != "" && !strings.EqualFold(m.Workflow, ev["workflow"]) && !strings.EqualFold(m.Workflow, ev["workflow_file"]) {
		return false
	}
	if m.Label != "" && !slices.ContainsFunc(strings.Split(ev["label"], "\n"), func(l string) bool { return strings.EqualFold(l, m.Label) }) {
		return false
	}
	return true
}

// triggerOn reports whether ev matches one of tr.On.
func triggerOn(tr config.Trigger, ev triggerEvent) bool {
	for _, o := range tr.On {
		if o.GitHub != nil && ev.matches(ingest.SourceGitHub, *o.GitHub) {
			return true
		}
		if o.Linear != nil && ev.matches(ingest.SourceLinear, *o.Linear) {
			return true
		}
	}
	return false
}

// vars is the event template var: every config.EventFields key, empty
// when the delivery has none.
func (ev triggerEvent) vars() map[string]any {
	m := make(map[string]any, len(config.EventFields))
	for _, f := range config.EventFields {
		m[f] = ev[f]
	}
	return m
}

// shellSafe matches event values a stall check may get unquoted: repo,
// pr, sha, action and such. Branch names and labels come from outside
// and may hold shell syntax.
var shellSafe = regexp.MustCompile(`^[A-Za-z0-9._/@:+=-]*$`)

// shellVars is vars with every value shellSafe does not match left
// empty, as on a poll, so a check can paste them into sh -c (JOB-72).
func (ev triggerEvent) shellVars() map[string]any {
	m := ev.vars()
	for k, v := range m {
		if s, _ := v.(string); !shellSafe.MatchString(s) {
			m[k] = ""
		}
	}
	return m
}

// key is the dedup key of an event fire with no check: the commit it is
// about, else the delivery.
func (ev triggerEvent) key() string {
	if ev["head_sha"] != "" {
		return ev["head_sha"]
	}
	return ev["delivery"]
}

// triggerEvents fires the triggers whose on: matches one of the
// dispatched webhook rows. It runs in the background under the tick
// lock: dispatch may run inside a pass.
func (c *Controller) triggerEvents(items []store.InboxItem) {
	type fire struct {
		name string
		ev   triggerEvent
	}
	var fires []fire
	trs := c.config().Triggers
	for _, it := range items {
		ev := eventOf(it)
		for _, name := range slices.Sorted(maps.Keys(trs)) {
			if len(trs[name].On) > 0 && triggerOn(trs[name], ev) {
				fires = append(fires, fire{name, ev})
			}
		}
	}
	if len(fires) == 0 {
		return
	}
	names := make([]string, len(fires))
	for i, f := range fires {
		names[i] = f.name
	}
	label := strings.Join(slices.Compact(slices.Clone(names)), ",")
	c.wg.Go(func() {
		// A root span per event job (it runs in its own goroutine, on a ctx
		// with no span): its wait for tickMu, the store calls and the
		// checks nest under it, and a waiter behind it links it (JOB-109).
		ctx, span := c.rootSpan(c.ctx, "trigger event "+label, attribute.StringSlice("jobshop.triggers", names))
		defer span.End()
		defer c.lockTickFor(ctx, "event", label)()
		if c.draining() {
			return
		}
		// One config for the whole event pass, like a tick (JOB-104).
		cfg := c.config()
		for _, f := range fires {
			c.runTriggerChecks(ctx, cfg, f.name, true, func(ctx context.Context, cfg *config.Config) {
				tr, ok := cfg.Triggers[f.name]
				if !ok {
					return // reloaded away
				}
				c.onEvent(ctx, cfg, f.name, tr, f.ev)
			})
		}
	})
}

// onEvent runs one matched trigger for an event: a stall trigger
// evaluates its signals (JOB-72), any other evaluates and fires. Holds
// the tick lock.
func (c *Controller) onEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) {
	if len(tr.When.Stall) > 0 {
		c.runStallTrigger(ctx, cfg, name, tr, ev)
		return
	}
	c.fireOnEvent(ctx, cfg, name, tr, ev)
}

// fireOnEvent evaluates and fires one trigger for an event. Holds the
// tick lock.
func (c *Controller) fireOnEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) {
	st, err := c.evalTrigger(ctx, cfg, name, tr)
	if errors.Is(err, errCheckPending) {
		return // its check runs outside the tick (JOB-109)
	}
	if err != nil {
		c.log.ErrorContext(ctx, "trigger", "trigger", name, "err", err)
		return
	}
	if !st.holds {
		c.log.DebugContext(ctx, "trigger event: does not hold", "trigger", name, "why", st.why)
		return
	}
	key := st.key
	if tr.When.Check == "" {
		key = ev.key()
		seen, err := c.triggerKeySeen(ctx, name, key)
		if err != nil || seen != "" {
			return
		}
	}
	t, err := c.fireTriggerEvent(ctx, cfg, name, tr, key, ev)
	if errors.Is(err, errCheckPending) {
		return // the run's intake check runs outside the tick (JOB-103)
	}
	if err != nil {
		c.log.ErrorContext(ctx, "trigger fire", "trigger", name, "err", err)
		return
	}
	c.log.InfoContext(ctx, "trigger fired", "trigger", name, "ticket", t.ID, "event", ev["event"], "delivery", ev["delivery"])
}

// triggerKeySeen is the ticket a trigger filed under key, "" if none.
func (c *Controller) triggerKeySeen(ctx context.Context, name, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	seen, err := c.store.List(ctx, store.Filter{MetaPack: triggerPack, MetaKey: "key", MetaValue: name + "/" + key, Limit: 1})
	if err != nil || len(seen) == 0 {
		return "", err
	}
	return seen[0].ID, nil
}

// triggerPollDue reports whether a tick should poll tr, and records the
// poll. A trigger with on: polls every fallback (default
// config.DefaultTriggerFallback), and never without when conditions;
// without on:, every tick unless fallback is set. Stall triggers too
// (JOB-72). Tick-only state.
func (c *Controller) triggerPollDue(name string, tr config.Trigger) bool {
	every := time.Duration(0)
	if len(tr.On) > 0 {
		if !tr.When.HasConditions() {
			return false
		}
		every = config.Duration(tr.Fallback, config.DefaultTriggerFallback)
	} else if tr.Fallback != "" {
		every = config.Duration(tr.Fallback, 0)
	}
	if every == 0 {
		return true
	}
	now := c.now()
	if last, ok := c.triggerPolled[name]; ok && now.Before(last.Add(every)) {
		return false
	}
	if c.triggerPolled == nil {
		c.triggerPolled = map[string]time.Time{}
	}
	c.triggerPolled[name] = now
	return true
}

// commandLabel marks the record of a trigger command run.
const commandLabel = "trigger-command"

// fireCommand runs tr.Command. The run is recorded as a message ticket
// carrying the trigger meta, so it holds the trigger while the command
// runs and dedups its key. Success closes it; failure keeps it open as
// needs-human with the output, so a broken command is reported once
// and the trigger waits for a human to close it.
func (c *Controller) fireCommand(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, meta json.RawMessage, tv map[string]any) (store.Ticket, error) {
	cm := tr.Command
	script, err := config.Fill(cm.Run, tv)
	if err != nil {
		return store.Ticket{}, fmt.Errorf("command: %w", err)
	}
	fields, err := json.Marshal(store.MessageFields{To: questionTo, From: "controller"})
	if err != nil {
		return store.Ticket{}, err
	}
	// A root with several projects has no default; the record only needs
	// one to live in.
	project := cfg.DefaultProject
	if projects := slices.Sorted(maps.Keys(cfg.Projects)); project == "" && len(projects) > 0 {
		project = projects[0]
	}
	if project == "" {
		return store.Ticket{}, errors.New("command: no project to record the run in")
	}
	t, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: store.KindMessage, Title: "Trigger " + name + ": command",
		Body: "```\n" + script + "\n```\n", Labels: []string{commandLabel}, Meta: meta, Fields: fields,
	})
	if err != nil {
		return store.Ticket{}, err
	}
	// The open record holds the trigger, so the command runs outside the
	// tick (JOB-103) and closes or flags the record when it ends. Its
	// timeout is the command's own, logged as a failure on the record.
	timeout := config.Duration(cm.Timeout, time.Minute)
	c.offGo("command\x00"+t.ID, "trigger_command", timeout+time.Minute, false, func(ctx context.Context) any {
		c.finishCommand(ctx, name, t, script, timeout)
		return nil
	})
	return t, nil
}

// finishCommand runs a trigger command and closes its record, or marks
// it needs-human with the output when the command fails.
func (c *Controller) finishCommand(ctx context.Context, name string, t store.Ticket, script string, timeout time.Duration) {
	out, runErr := c.runTriggerCommand(ctx, script, timeout)
	if runErr == nil {
		if out != "" {
			_, _ = store.AddComment(ctx, c.store, "controller", t.ID, "Output:\n```\n"+out+"\n```")
		}
		c.log.InfoContext(ctx, "trigger command ran", "trigger", name, "ticket", t.ID)
		if _, err := store.Close(ctx, c.store, "controller", t.ID, nil, "done"); err != nil && c.ctx.Err() == nil {
			c.log.ErrorContext(ctx, "trigger command close", "trigger", name, "ticket", t.ID, "err", err)
		}
		return
	}
	c.log.ErrorContext(ctx, "trigger command failed", "trigger", name, "ticket", t.ID, "err", runErr)
	body := fmt.Sprintf("Command failed: %v\n\nOutput:\n```\n%s\n```\nThe trigger fires again once this ticket is closed.", runErr, out)
	_, _ = store.AddComment(ctx, c.store, "controller", t.ID, body)
	if _, err := c.store.Update(ctx, "controller", t.ID, store.Update{AddLabels: []string{needsHumanLabel}}); err != nil && c.ctx.Err() == nil {
		c.log.ErrorContext(ctx, "trigger command flag", "trigger", name, "ticket", t.ID, "err", err)
	}
}

// triggerOutputMax bounds the command output kept on the ticket.
const triggerOutputMax = 8 << 10

// runTriggerCommand runs script with sh in the root dir and returns its
// combined output, trimmed to its tail.
func (c *Controller) runTriggerCommand(ctx context.Context, script string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd, done := telemetry.ExecContext(ctx, "trigger-command", "sh", "-c", script)
	cmd.Dir = c.root.Dir
	b, err := cmd.CombinedOutput()
	done(err)
	out := strings.TrimSpace(string(b))
	if len(out) > triggerOutputMax {
		out = "..." + out[len(out)-triggerOutputMax:]
	}
	if ctx.Err() != nil {
		c.offTimedOut(ctx, "trigger_command", "command", timeout)
		return out, fmt.Errorf("timed out after %s", timeout)
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return out, fmt.Errorf("exit %d", exit.ExitCode())
	}
	return out, err
}
