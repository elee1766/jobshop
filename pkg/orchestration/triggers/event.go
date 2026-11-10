package triggers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/ingest"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Event triggers (JOB-62, docs/notes/JOB-62-event-triggers.md). A
// trigger with on: fires when an admitted webhook delivery matches,
// without waiting for the next tick; its when conditions and dedup are
// the same as a poll's. fallback: spaces its polls, so a lost webhook only
// costs latency.

// Event is the fields of one delivery a trigger can match and template
// on, keyed as config.EventFields.
type Event map[string]string

// EventOf reads the matchable fields of an inbox item. Labels are joined
// with newlines in "label".
func EventOf(it domain.InboxItem) Event {
	ev := Event{"source": it.Source, "event": it.Kind, "repo": it.RepoOrTeam, "delivery": it.DeliveryID}
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

func githubEvent(ev Event, payload json.RawMessage) {
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

// Matches reports whether ev has every field m sets.
func (ev Event) Matches(source string, m config.EventMatch) bool {
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

// On reports whether ev matches one of tr.On.
func On(tr config.Trigger, ev Event) bool {
	for _, o := range tr.On {
		if o.Source != "" && strings.EqualFold(o.Source, ev["source"]) && (o.Event == "" || strings.EqualFold(o.Event, ev["event"])) {
			return true
		}
		if o.GitHub != nil && ev.Matches(ingest.SourceGitHub, *o.GitHub) {
			return true
		}
		if o.Linear != nil && ev.Matches(ingest.SourceLinear, *o.Linear) {
			return true
		}
	}
	return false
}

// Vars is the event template var: every config.EventFields key, empty
// when the delivery has none.
func (ev Event) Vars() map[string]any {
	m := make(map[string]any, len(config.EventFields))
	for _, f := range config.EventFields {
		m[f] = ev[f]
	}
	return m
}

// shellSafe matches event values a signal check may get unquoted: repo,
// pr, sha, action and such. Branch names and labels come from outside
// and may hold shell syntax.
var shellSafe = regexp.MustCompile(`^[A-Za-z0-9._/@:+=-]*$`)

// ShellVars is Vars with every value shellSafe does not match left
// empty, as on a poll, so a check can paste them into sh -c (JOB-72).
func (ev Event) ShellVars() map[string]any {
	m := ev.Vars()
	for k, v := range m {
		if s, _ := v.(string); !shellSafe.MatchString(s) {
			m[k] = ""
		}
	}
	return m
}

// Key is the dedup key of an event fire with no check: the commit it is
// about, else the delivery.
func (ev Event) Key() string {
	if ev["head_sha"] != "" {
		return ev["head_sha"]
	}
	return ev["delivery"]
}

// Events fires the triggers whose on: matches one of the dispatched
// webhook rows. It runs in the background under the tick lock: dispatch
// may run inside a pass.
func (e *Engine) Events(items []domain.InboxItem) {
	type fire struct {
		name string
		ev   Event
	}
	var fires []fire
	trs := e.host.Config().Triggers
	for _, it := range items {
		ev := EventOf(it)
		for _, name := range slices.Sorted(maps.Keys(trs)) {
			if trs[name].Handler == "" && len(trs[name].On) > 0 && On(trs[name], ev) {
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
	e.host.Go(func() {
		// A root span per event job (it runs in its own goroutine, on a ctx
		// with no span): its wait for the tick lock, the store calls and the
		// checks nest under it, and a waiter behind it links it (JOB-109).
		ctx, span := e.host.RootSpan(e.host.Context(), "trigger event "+label, attribute.StringSlice("jobshop.triggers", names))
		defer span.End()
		defer e.host.LockTick(ctx, "event", label)()
		if e.host.Draining() {
			return
		}
		// One config for the whole event pass, like a tick (JOB-104).
		cfg := e.host.Config()
		for _, f := range fires {
			e.RunChecked(ctx, cfg, f.name, true, func(ctx context.Context, cfg *config.Config) {
				tr, ok := cfg.Triggers[f.name]
				if !ok {
					return // reloaded away
				}
				e.OnEvent(ctx, cfg, f.name, tr, f.ev)
			})
		}
	})
}

// OnEvent runs one matched trigger for an event: a scan trigger
// evaluates its signals (JOB-72), any other evaluates and fires. Holds
// the tick lock.
func (e *Engine) OnEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev Event) {
	if len(tr.When.Scan) > 0 {
		e.runScan(ctx, cfg, name, tr, ev)
		return
	}
	e.fireOnEvent(ctx, cfg, name, tr, ev)
}

// fireOnEvent evaluates and fires one trigger for an event. Holds the
// tick lock.
func (e *Engine) fireOnEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev Event) {
	st, err := e.Eval(ctx, cfg, name, tr)
	if errors.Is(err, ErrCheckPending) {
		return // its check runs outside the tick (JOB-109)
	}
	if err != nil {
		e.log().ErrorContext(ctx, "trigger", "trigger", name, "err", err)
		return
	}
	if !st.Holds {
		e.log().DebugContext(ctx, "trigger event: does not hold", "trigger", name, "why", st.Why)
		return
	}
	key := st.Key
	if tr.When.Check == "" {
		key = ev.Key()
		seen, err := e.keySeen(ctx, name, key)
		if err != nil || seen != "" {
			return
		}
	}
	t, err := e.host.Fire(ctx, cfg, name, tr, key, ev)
	if errors.Is(err, ErrCheckPending) {
		return // the run's intake check runs outside the tick (JOB-103)
	}
	if err != nil {
		e.log().ErrorContext(ctx, "trigger fire", "trigger", name, "err", err)
		return
	}
	e.log().InfoContext(ctx, "trigger fired", "trigger", name, "ticket", t.ID, "event", ev["event"], "delivery", ev["delivery"])
}

// keySeen is the ticket a trigger filed under key, "" if none.
func (e *Engine) keySeen(ctx context.Context, name, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	seen, err := e.host.Store().List(ctx, store.Filter{MetaPack: Pack, MetaKey: "key", MetaValue: name + "/" + key, Limit: 1})
	if err != nil || len(seen) == 0 {
		return "", err
	}
	return seen[0].ID, nil
}
