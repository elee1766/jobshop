package control

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elee1766/acphost/pkg/proc"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Ingest admission (#348): a stage that lease-claims pending inbox items,
// runs the mutating hooks in config order and then the validating hooks
// on the final event, and moves the item to admitted, denied or held.

const (
	admissionHolder = "admission"
	// admissionIdle is how long the pump sleeps when nothing is pending.
	admissionIdle = time.Second
	// admissionOn is the "on" field of an admission event; on: intake
	// hooks run as validating hooks get "intake".
	admissionOn = "admission"
)

// admissionEvent is what an admission hook reads. DeliveryID lets a hook
// dedupe: an item whose lease expired is claimed and admitted again.
type admissionEvent struct {
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

// admissionResult is one hook's entry in an item's hook_results.
type admissionResult struct {
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
	admitMutated   = "mutated"
	admitUnchanged = "unchanged"
	admitFailed    = "failed"
	admitIgnored   = "ignored"
)

// admissionHooks is ingest.admission plus every on: intake hook as a
// validating hook for Linear events, so #288 setups keep their guard.
func admissionHooks(cfg *config.Config) []config.AdmissionHook {
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

// pumpAdmission admits pending items and dispatches admitted ones (#442)
// until the controller stops. It waits on the inbox notifier (#679), so
// a new item is admitted as it lands; admissionIdle stays as the
// backstop for state changes the notifier does not see (a lease
// expiring, a hold answered).
func (c *Controller) pumpAdmission() {
	for c.ctx.Err() == nil {
		changed := c.inboxNotifier().Changed()
		c.ingestPass(c.ctx)
		select {
		case <-c.ctx.Done():
		case <-changed:
		case <-time.After(admissionIdle):
		}
	}
}

// releaseAnswered moves held items whose needs-human ticket is closed
// back to pending (#353). Admission only claims pending items; on the
// next claim admit reads the answer through askedHuman: done admits,
// any other resolution denies. Held is not pruned, so without this an
// answered item would stay held forever.
func (c *Controller) releaseAnswered(ctx context.Context) error {
	held, err := ingest.InboxList(ctx, c.store, store.InboxFilter{State: store.InboxHeld, Limit: 200})
	if err != nil {
		return err
	}
	for _, it := range held {
		a, ok := c.askedHuman(ctx, "admission/"+it.DeliveryID)
		if !ok || a.Decision == rpc.GuardAskHuman {
			continue
		}
		if _, err := ingest.InboxRequeue(ctx, c.store, it.ID, store.InboxHeld, store.InboxPending); err != nil {
			if errors.Is(err, store.ErrConflict) {
				continue // moved or leased meanwhile
			}
			return err
		}
		c.log.InfoContext(ctx, "inbox admission released", "delivery", it.DeliveryID, "ticket", a.Ticket, "answer", a.Decision)
	}
	return nil
}

// admissionLease is how long an admitter holds an item: a minute plus
// every hook's timeout.
func admissionLease(hooks []config.AdmissionHook) time.Duration {
	lease := time.Minute
	for _, h := range hooks {
		lease += config.Duration(h.Timeout, 30*time.Second)
	}
	return lease
}

// admitNext claims one pending item and admits it. false: none pending.
func (c *Controller) admitNext(ctx context.Context) (bool, error) {
	cfg := c.config()
	hooks := admissionHooks(cfg)
	it, err := ingest.InboxClaim(ctx, c.store, []string{store.InboxPending}, admissionHolder, admissionLease(hooks))
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, answered := c.humanAnswer(ctx, it.DeliveryID)
	run := c.admissionRunner().run(ctx, cfg.DefaultProject, hooks, it, answered)
	return true, c.finishAdmission(ctx, it, admissionHolder, run)
}

// humanAnswer is a human's closed answer to an earlier hold of delivery.
func (c *Controller) humanAnswer(ctx context.Context, delivery string) (rpc.GuardResult, bool) {
	a, ok := c.askedHuman(ctx, "admission/"+delivery)
	if !ok || a.Decision == rpc.GuardAskHuman {
		return rpc.GuardResult{}, false
	}
	return a, true
}

// admissionRun is what running the hooks on one item produced. Event
// is the event after the mutating hooks; Asks are the ask_human
// reasons. The state is decided by finishAdmission, which alone sees
// the store, so an external worker (#443) and the in-process loop
// decide the same way.
type admissionRun struct {
	Event   admissionEvent    `json:"event"`
	Results []admissionResult `json:"results"`
	Denied  bool              `json:"denied,omitempty"`
	Asks    []string          `json:"asks,omitempty"`
}

// admissionRunner runs admission hooks. It needs no store, so the
// external `ingest work` process uses it too.
type admissionRunner struct {
	dir string // working dir for commands, root for secret files
}

func (c *Controller) admissionRunner() admissionRunner { return admissionRunner{dir: c.root.Dir} }

// run runs the mutating hooks, then, unless a human already answered,
// the validating hooks.
func (ar admissionRunner) run(ctx context.Context, project string, hooks []config.AdmissionHook, it store.InboxItem, answered bool) admissionRun {
	ev := admissionEvent{
		On: admissionOn, DeliveryID: it.DeliveryID, Source: it.Source, Kind: it.Kind, RepoOrTeam: it.RepoOrTeam,
		Number: it.Number, CommentID: it.CommentID, Actor: it.Actor, Project: project,
		ReceivedAt: it.ReceivedAt, Payload: it.Payload,
	}
	subj := config.AdmissionSubject{
		Source: it.Source, Kind: it.Kind, RepoOrTeam: it.RepoOrTeam, Actor: it.Actor,
		Project: ev.Project, Bot: strings.HasSuffix(it.Actor, "[bot]"),
	}
	out := admissionRun{Results: []admissionResult{}}
	for _, h := range hooks {
		if h.Type == config.AdmissionMutating && h.Match.Matches(subj) {
			var r admissionResult
			ev, r = ar.mutate(ctx, h, ev)
			out.Results = append(out.Results, r)
		}
	}
	out.Event = ev
	if answered {
		return out // a human's answer stands; validating hooks do not run
	}
	for _, h := range hooks {
		if h.Type != config.AdmissionValidating || !h.Match.Matches(subj) {
			continue
		}
		r := ar.validate(ctx, h, ev)
		out.Results = append(out.Results, r)
		switch r.Decision {
		case rpc.GuardDeny:
			out.Denied = true
		case rpc.GuardAskHuman:
			out.Asks = append(out.Asks, h.Name+": "+r.Reason)
		}
	}
	return out
}

// settle decides an item's state from a run and a human's answer.
// answered is false when no human has closed a hold for it.
func settle(run admissionRun, a rpc.GuardResult, answered bool) (string, json.RawMessage, []admissionResult) {
	results := run.Results
	if answered {
		results = append(results, admissionResult{Name: "human", Type: config.AdmissionValidating, Decision: a.Decision, Reason: a.Reason})
		if a.Decision == rpc.GuardAllow {
			return store.InboxAdmitted, run.Event.Payload, results
		}
		return store.InboxDenied, nil, results
	}
	switch {
	case run.Denied:
		return store.InboxDenied, nil, results
	case len(run.Asks) > 0:
		return store.InboxHeld, nil, results
	}
	return store.InboxAdmitted, run.Event.Payload, results
}

// finishAdmission stores a run's outcome for an item holder leases,
// filing the needs-human ticket for a hold. A human's closed answer to
// an earlier hold overrides the validating hooks.
func (c *Controller) finishAdmission(ctx context.Context, it store.InboxItem, holder string, run admissionRun) error {
	a, answered := c.humanAnswer(ctx, it.DeliveryID)
	state, admitted, results := settle(run, a, answered)
	if state == store.InboxHeld {
		c.holdForHuman(ctx, "admission/"+it.DeliveryID, run.Event, strings.Join(run.Asks, "\n"))
	}
	raw, err := json.Marshal(results)
	if err != nil {
		return err
	}
	u := store.InboxUpdate{State: state, HookResults: raw}
	if state == store.InboxAdmitted {
		u.Admitted = admitted
	}
	if _, err := ingest.InboxSetState(ctx, c.store, it.ID, holder, u); err != nil {
		return err
	}
	if state == store.InboxAdmitted {
		// Dispatch-only workers long-poll for admitted items (#679).
		c.inboxNotifier().Advance()
	}
	c.log.InfoContext(ctx, "inbox admission", "delivery", it.DeliveryID, "state", state, "by", holder)
	return nil
}

// mutate runs a mutating hook. Any failure leaves ev unchanged.
func (ar admissionRunner) mutate(ctx context.Context, h config.AdmissionHook, ev admissionEvent) (admissionEvent, admissionResult) {
	r := admissionResult{Name: h.Name, Type: h.Type}
	start := time.Now()
	out, err := ar.callAdmission(ctx, h, ev)
	r.DurationMS = time.Since(start).Milliseconds()
	fail := func(err error) (admissionEvent, admissionResult) {
		r.Decision, r.Error = admitFailed, err.Error()
		if h.Policy() == config.FailurePolicyIgnore {
			r.Decision = admitIgnored
		}
		return ev, r
	}
	if err != nil {
		return fail(err)
	}
	r.Decision = admitUnchanged
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
	r.Decision = admitMutated
	return next, r
}

// validate runs a validating hook. Under failure_policy fail an error,
// timeout or bad answer denies; under ignore it is skipped.
func (ar admissionRunner) validate(ctx context.Context, h config.AdmissionHook, ev admissionEvent) admissionResult {
	r := admissionResult{Name: h.Name, Type: h.Type}
	if h.Intake {
		ev.On = config.HookIntake
	}
	start := time.Now()
	out, err := ar.callAdmission(ctx, h, ev)
	r.DurationMS = time.Since(start).Milliseconds()
	var a hookAnswer
	if err == nil && len(out) > 0 {
		if jerr := json.Unmarshal(out, &a); jerr != nil {
			err = fmt.Errorf("answer is not JSON: %w", jerr)
		} else if a.Decision != rpc.GuardAllow && a.Decision != rpc.GuardDeny && a.Decision != rpc.GuardAskHuman {
			err = fmt.Errorf("decision %q is not allow, deny or ask_human", a.Decision)
		}
	}
	if err != nil {
		r.Error = err.Error()
		if h.Policy() == config.FailurePolicyIgnore {
			r.Decision = admitIgnored
		} else {
			r.Decision, r.Reason = rpc.GuardDeny, "hook failed: "+err.Error()
		}
		return r
	}
	if len(out) == 0 {
		a.Decision = rpc.GuardAllow
	}
	r.Decision, r.Reason = a.Decision, a.Reason
	return r
}

// callAdmission sends ev to the hook and returns its trimmed answer.
func (ar admissionRunner) callAdmission(ctx context.Context, h config.AdmissionHook, ev admissionEvent) ([]byte, error) {
	body, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, config.Duration(h.Timeout, 30*time.Second))
	defer cancel()
	var out []byte
	if h.URL != "" {
		out, err = ar.postAdmission(ctx, h, ev.DeliveryID, body)
	} else {
		out, err = ar.execAdmission(ctx, h, body)
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("timed out after %s", config.Duration(h.Timeout, 30*time.Second))
	}
	return bytes.TrimSpace(out), err
}

func (ar admissionRunner) execAdmission(ctx context.Context, h config.AdmissionHook, body []byte) ([]byte, error) {
	// proc.Start runs it in its own process group and kills the group on
	// timeout, so a child of sh (sleep, curl) cannot hold stdout open.
	// Admission has no pool: no sandbox (#810).
	spec := proc.Spec{Command: "sh", Args: []string{"-c", h.Command}, Dir: ar.dir}
	var stdout bytes.Buffer
	var stderr strings.Builder
	p, err := proc.Start(ctx, spec, proc.Stdin(bytes.NewReader(body)), proc.Stdout(&stdout), proc.Stderr(&stderr))
	if err != nil {
		return nil, err
	}
	ex, err := p.Wait()
	if err != nil {
		return nil, err
	}
	if ex.ExitCode != 0 || ex.Signal != "" {
		return nil, fmt.Errorf("exited %s: %s", exitText(ex), strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// admissionSignature is the X-Jobshop-Signature-256 value for body.
func admissionSignature(secret, body []byte) string {
	m := hmac.New(sha256.New, secret)
	m.Write(body)
	return "sha256=" + hex.EncodeToString(m.Sum(nil))
}

func (ar admissionRunner) postAdmission(ctx context.Context, h config.AdmissionHook, delivery string, body []byte) ([]byte, error) {
	secret, err := config.ReadSecret(ctx, h.SecretFile.In(ar.dir))
	if err != nil {
		return nil, fmt.Errorf("secret: %w", err)
	}
	secret = bytes.TrimSpace(secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	reqID := make([]byte, 16)
	_, _ = rand.Read(reqID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jobshop-Request-Id", hex.EncodeToString(reqID))
	req.Header.Set("X-Jobshop-Delivery", delivery)
	req.Header.Set("X-Jobshop-Signature-256", admissionSignature(secret, body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(out[:min(len(out), 200)])))
	}
	return out, nil
}

// holdForHuman files one needs-human ticket per held delivery, like a
// refused intake (#288). Closing it done admits the item when it is
// next claimed; any other resolution denies it.
func (c *Controller) holdForHuman(ctx context.Context, key string, ev admissionEvent, reason string) {
	open, err := c.store.List(ctx, store.Filter{MetaPack: hookMetaPack, MetaKey: "key", MetaValue: key, Order: store.OrderNewest, Limit: 1})
	if err == nil && len(open) > 0 && open[0].Status != store.StatusClosed {
		return
	}
	meta, err := json.Marshal(map[string]map[string]string{hookMetaPack: {"key": key}})
	if err != nil {
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "An admission hook asks a human about %s %s delivery %s", ev.Source, ev.Kind, ev.DeliveryID)
	if ev.RepoOrTeam != "" {
		fmt.Fprintf(&b, " in %s", ev.RepoOrTeam)
		if ev.Number != 0 {
			fmt.Fprintf(&b, "#%d", ev.Number)
		}
	}
	if ev.Actor != "" {
		fmt.Fprintf(&b, " by %s", ev.Actor)
	}
	fmt.Fprintf(&b, ".\n\nReason:\n%s\n\nQuestion: may this event reach an agent?\n", reason)
	b.WriteString("Close done to allow it, close with any other resolution to keep it out.")
	project := ev.Project
	if project == "" {
		project = c.config().DefaultProject
	}
	prio := 1
	if _, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: store.KindTask, Title: "Admission held " + ev.Source + " delivery " + ev.DeliveryID,
		Body: b.String(), Priority: &prio, Labels: []string{needsHumanLabel, questionLabel}, Meta: meta,
	}); err != nil {
		c.log.WarnContext(ctx, "admission: file ticket", "delivery", ev.DeliveryID, "err", err)
	}
}
