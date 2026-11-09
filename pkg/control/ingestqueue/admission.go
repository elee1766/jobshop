package ingestqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/acphost/pkg/proc"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Ingest admission (#348): the hook half. A Runner takes one claimed
// inbox item through the mutating hooks in config order and then the
// validating hooks on the final event. Deciding the item's state from
// the run (Settle) is pure; storing it, the hold ticket and the
// human's answer stay with the controller, which alone sees the store.

const (
	// Holder is the in-process admitter's inbox lease holder.
	Holder = "admission"
	// WorkerPrefix makes an external worker's inbox holder distinct from
	// the in-process admitter's.
	WorkerPrefix = "worker:"
	// On is the "on" field of an admission event; on: intake hooks run as
	// validating hooks get "intake".
	On = "admission"
)

// Event is what an admission hook reads. DeliveryID lets a hook dedupe:
// an item whose lease expired is claimed and admitted again.
type Event struct {
	On         string          `json:"on"`
	DeliveryID string          `json:"delivery_id"`
	Source     string          `json:"source"`
	Kind       string          `json:"kind,omitempty"`
	RepoOrTeam string          `json:"repo_or_team,omitempty"`
	Number     int64           `json:"number,omitempty"`
	CommentID  string          `json:"comment_id,omitempty"`
	Actor      string          `json:"actor,omitempty"`
	Project    string          `json:"project,omitempty"`
	ReceivedAt time.Time       `json:"received_at"`
	Payload    json.RawMessage `json:"payload"`
}

// mutateAnswer is a mutating hook's answer: a full event or an RFC 6902
// patch against the event. Neither is no change.
type mutateAnswer struct {
	Event  json.RawMessage `json:"event"`
	Patch  json.RawMessage `json:"patch"`
	Reason string          `json:"reason"`
}

// Result is one hook's entry in an item's hook_results.
type Result struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Decision   string          `json:"decision"` // allow, deny, ask_human, mutated, unchanged, failed, ignored
	Reason     string          `json:"reason,omitempty"`
	Diff       json.RawMessage `json:"diff,omitempty"`
	DurationMS int64           `json:"duration_ms"`
	Error      string          `json:"error,omitempty"`
}

// Result decisions beyond the validating ones.
const (
	Mutated   = "mutated"
	Unchanged = "unchanged"
	Failed    = "failed"
	Ignored   = "ignored"
)

// Hooks is ingest.admission plus every on: intake hook as a validating
// hook for Linear events, so #288 setups keep their guard.
func Hooks(cfg *config.Config) []config.AdmissionHook {
	out := append([]config.AdmissionHook(nil), cfg.Ingest.Admission...)
	n := 0
	for _, h := range cfg.Hooks {
		if h.On != config.HookIntake {
			continue
		}
		n++
		out = append(out, config.AdmissionHook{
			Name: fmt.Sprintf("on-intake-%d", n), Type: config.AdmissionValidating, Command: h.Command, Intake: true,
			Timeout: h.Timeout, Match: config.AdmissionMatch{Sources: []string{config.IngestLinear}},
		})
	}
	return out
}

// Lease is how long an admitter holds an item: a minute plus every
// hook's timeout.
func Lease(hs []config.AdmissionHook) time.Duration {
	lease := time.Minute
	for _, h := range hs {
		lease += h.Retry.Budget(config.Duration(h.Timeout, 30*time.Second))
	}
	return lease
}

// Run is what running the hooks on one item produced. Event is the
// event after the mutating hooks; Asks are the ask_human reasons. The
// state is decided by Settle and stored by the controller, so an
// external worker (#443) and the in-process loop decide the same way.
type Run struct {
	Event   Event    `json:"event"`
	Results []Result `json:"results"`
	Denied  bool     `json:"denied,omitempty"`
	Asks    []string `json:"asks,omitempty"`
}

// Runner runs admission hooks. It needs no store, so the external
// `ingest work` process uses it too.
type Runner struct {
	Dir string // working dir for commands, root for secret files
	// After waits between url hook retries; nil is time.After.
	After func(time.Duration) <-chan time.Time
	// Executor runs command hooks; nil runs them as `sh -c` in Dir with
	// the process group killed on timeout.
	Executor hooks.Executor
}

// Run runs the mutating hooks, then, unless a human already answered,
// the validating hooks.
func (ar Runner) Run(ctx context.Context, project string, hs []config.AdmissionHook, it domain.InboxItem, answered bool) Run {
	ev := Event{
		On: On, DeliveryID: it.DeliveryID, Source: it.Source, Kind: it.Kind, RepoOrTeam: it.RepoOrTeam,
		Number: it.Number, CommentID: it.CommentID, Actor: it.Actor, Project: project,
		ReceivedAt: it.ReceivedAt, Payload: it.Payload,
	}
	subj := config.AdmissionSubject{
		Source: it.Source, Kind: it.Kind, RepoOrTeam: it.RepoOrTeam, Actor: it.Actor,
		Project: ev.Project, Bot: strings.HasSuffix(it.Actor, "[bot]"),
	}
	out := Run{Results: []Result{}}
	for _, h := range hs {
		if h.Type == config.AdmissionMutating && h.Match.Matches(subj) {
			var r Result
			ev, r = ar.mutate(ctx, h, ev)
			out.Results = append(out.Results, r)
		}
	}
	out.Event = ev
	if answered {
		return out // a human's answer stands; validating hooks do not run
	}
	for _, h := range hs {
		if h.Type != config.AdmissionValidating || !h.Match.Matches(subj) {
			continue
		}
		r := ar.validate(ctx, h, ev)
		out.Results = append(out.Results, r)
		switch r.Decision {
		case domain.GuardDeny:
			out.Denied = true
		case domain.GuardAskHuman:
			out.Asks = append(out.Asks, h.Name+": "+r.Reason)
		}
	}
	return out
}

// Settle decides an item's state from a run and a human's answer.
// answered is false when no human has closed a hold for it.
func Settle(run Run, a rpc.GuardResult, answered bool) (string, json.RawMessage, []Result) {
	results := run.Results
	if answered {
		results = append(results, Result{Name: "human", Type: config.AdmissionValidating, Decision: a.Decision, Reason: a.Reason})
		if a.Decision == domain.GuardAllow {
			return domain.InboxAdmitted, run.Event.Payload, results
		}
		return domain.InboxDenied, nil, results
	}
	switch {
	case run.Denied:
		return domain.InboxDenied, nil, results
	case len(run.Asks) > 0:
		return domain.InboxHeld, nil, results
	}
	return domain.InboxAdmitted, run.Event.Payload, results
}

// mutate runs a mutating hook. Any failure leaves ev unchanged.
func (ar Runner) mutate(ctx context.Context, h config.AdmissionHook, ev Event) (Event, Result) {
	r := Result{Name: h.Name, Type: h.Type}
	start := time.Now()
	out, err := ar.call(ctx, h, ev)
	r.DurationMS = time.Since(start).Milliseconds()
	fail := func(err error) (Event, Result) {
		r.Decision, r.Error = Failed, err.Error()
		if h.Policy() == config.FailurePolicyIgnore {
			r.Decision = Ignored
		}
		return ev, r
	}
	if err != nil {
		return fail(err)
	}
	r.Decision = Unchanged
	if len(out) == 0 {
		return ev, r
	}
	var a mutateAnswer
	if err := json.Unmarshal(out, &a); err != nil {
		return fail(fmt.Errorf("answer is not JSON: %w", err))
	}
	r.Reason = a.Reason
	before, err := json.Marshal(ev)
	if err != nil {
		return fail(err)
	}
	var after []byte
	switch {
	case len(a.Event) > 0 && len(a.Patch) > 0:
		return fail(errors.New("answer has both event and patch"))
	case len(a.Event) > 0:
		after = a.Event
	case len(a.Patch) > 0:
		after, err = applyJSONPatch(before, a.Patch)
		if err != nil {
			return fail(err)
		}
		r.Diff = a.Patch
	default:
		return ev, r
	}
	next := ev
	next.Payload = nil
	if err := json.Unmarshal(after, &next); err != nil {
		return fail(fmt.Errorf("mutated event: %w", err))
	}
	// Identity fields stay: a hook may not move an item to another delivery.
	next.On, next.DeliveryID, next.Source = ev.On, ev.DeliveryID, ev.Source
	if len(next.Payload) == 0 {
		next.Payload = json.RawMessage("null")
	}
	if r.Diff == nil {
		r.Diff = topLevelDiff(before, after)
	}
	r.Decision = Mutated
	return next, r
}

// validate runs a validating hook. Under failure_policy fail an error,
// timeout or bad answer denies; under ignore it is skipped.
func (ar Runner) validate(ctx context.Context, h config.AdmissionHook, ev Event) Result {
	r := Result{Name: h.Name, Type: h.Type}
	if h.Intake {
		ev.On = config.HookIntake
	}
	start := time.Now()
	out, err := ar.call(ctx, h, ev)
	r.DurationMS = time.Since(start).Milliseconds()
	var a hooks.Answer
	if err == nil && len(out) > 0 {
		if jerr := json.Unmarshal(out, &a); jerr != nil {
			err = fmt.Errorf("answer is not JSON: %w", jerr)
		} else if a.Decision != domain.GuardAllow && a.Decision != domain.GuardDeny && a.Decision != domain.GuardAskHuman {
			err = fmt.Errorf("decision %q is not allow, deny or ask_human", a.Decision)
		}
	}
	if err != nil {
		r.Error = err.Error()
		if h.Policy() == config.FailurePolicyIgnore {
			r.Decision = Ignored
		} else {
			r.Decision, r.Reason = domain.GuardDeny, "hook failed: "+err.Error()
		}
		return r
	}
	if len(out) == 0 {
		a.Decision = domain.GuardAllow
	}
	r.Decision, r.Reason = a.Decision, a.Reason
	return r
}

// call sends ev to the hook and returns its trimmed answer.
func (ar Runner) call(ctx context.Context, h config.AdmissionHook, ev Event) ([]byte, error) {
	body, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}
	ex := ar.Executor
	if ex == nil {
		ex = shell{dir: ar.Dir}
	}
	return (hooks.Runner{Dir: ar.Dir, After: ar.After, Executor: ex}).Admission(ctx, h, body, ev.DeliveryID)
}

// shell runs a hook command as `sh -c` in dir. Admission has no pool, so
// there is no sandbox; the process group still dies on timeout.
type shell struct{ dir string }

func (s shell) ExecuteHook(ctx context.Context, command, _ string, body []byte) (hooks.CommandResult, error) {
	var stdout bytes.Buffer
	var stderr strings.Builder
	spec := proc.Spec{Command: "sh", Args: []string{"-c", command}, Dir: s.dir}
	p, err := proc.Start(ctx, spec, proc.Stdin(bytes.NewReader(body)), proc.Stdout(&stdout), proc.Stderr(&stderr))
	if err != nil {
		return hooks.CommandResult{}, err
	}
	ex, err := p.Wait()
	return hooks.CommandResult{Output: stdout.Bytes(), Stderr: stderr.String(), ExitCode: ex.ExitCode, Signal: ex.Signal}, err
}
