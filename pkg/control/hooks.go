package control

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/automation"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/store"
)

var (
	hookInvocations = store.CollectionSpec{Name: "hook_invocations", Index: []store.Field{{Name: "run", Path: "run"}, {Name: "trigger", Path: "trigger"}, {Name: "event", Path: "event.delivery"}}}
	hookCalls       = store.CollectionSpec{Name: "hook_calls", Index: []store.Field{{Name: "invocation", Path: "invocation"}}}
	hookAttempts    = store.CollectionSpec{Name: "hook_attempts", Index: []store.Field{{Name: "invocation", Path: "invocation"}}}
)

// Hook inputs are immutable event-time snapshots, not reads of mutable tickets.
type automationEvent struct {
	Source     string         `json:"source"`
	Event      string         `json:"event"`
	Delivery   string         `json:"delivery"`
	Repository string         `json:"repository"`
	Payload    map[string]any `json:"payload"`
	Fields     map[string]any `json:"fields"`
}

func (c *Controller) lastHookFire(ctx context.Context, name string) (time.Time, error) {
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return time.Time{}, err
	}
	records, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"trigger": name}, States: []string{"done", "planned", "failed"}, Order: store.OrderCreatedDesc, Limit: 1})
	if err != nil || len(records) == 0 {
		return time.Time{}, err
	}
	var in hookInvocation
	if err := decodeHookJSON(records[0].Value, &in); err != nil {
		return time.Time{}, err
	}
	return in.AcceptedAt, nil
}

type hookStep struct {
	ID         string         `json:"id"`
	TicketID   string         `json:"ticket_id"`
	Pool       string         `json:"pool"`
	Status     string         `json:"status"`
	Resolution string         `json:"resolution"`
	Outcome    string         `json:"outcome"`
	Output     map[string]any `json:"output"`
	Inputs     map[string]any `json:"inputs"`
	Failures   int            `json:"failures"`
	Retries    int            `json:"retries"`
}
type hookRun struct {
	ID       string            `json:"id"`
	Workflow string            `json:"workflow"`
	Key      string            `json:"key"`
	Project  string            `json:"project"`
	Inputs   map[string]any    `json:"inputs"`
	Vars     map[string]string `json:"vars"`
	State    string            `json:"state"`
	Steps    []hookStep        `json:"steps"`
}
type hookInvocation struct {
	Public           json.RawMessage       `json:"public,omitempty"`
	AcceptedAt       time.Time             `json:"accepted_at"`
	ParentInvocation string                `json:"parent_invocation,omitempty"`
	Targets          []string              `json:"targets,omitempty"`
	Trigger          string                `json:"trigger,omitempty"`
	Run              string                `json:"run,omitempty"`
	Step             string                `json:"step,omitempty"`
	StepRevision     int64                 `json:"step_revision,omitempty"`
	Entry            string                `json:"entry"`
	Digest           string                `json:"digest"`
	ReadSet          map[string]int64      `json:"read_set,omitempty"`
	Event            *automationEvent      `json:"event,omitempty"`
	Args             map[string]any        `json:"args,omitempty"`
	RunInput         *hookRun              `json:"run_input,omitempty"`
	StepInput        *hookStep             `json:"step_input,omitempty"`
	Actions          []automation.Action   `json:"actions,omitempty"`
	Plan             *store.Graph          `json:"plan,omitempty"`
	Error            string                `json:"error,omitempty"`
	Output           string                `json:"output,omitempty"`
	Decisions        []automation.Decision `json:"decisions,omitempty"`
}

func hookKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (c *Controller) putHook(ctx context.Context, key string, in hookInvocation) error {
	if in.AcceptedAt.IsZero() {
		in.AcceptedAt = c.now()
	}
	col, err := c.store.Collection(ctx, hookInvocations)
	if err != nil {
		return err
	}
	in.Public = c.hookInvocationPublic(in)
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	rec, created, err := col.Put(ctx, store.Record{Key: hookKey(key), State: "pending", Value: data})
	if err == nil && !created && (rec.State == "superseded" || rec.State == "retry") {
		rec.State, rec.Value = "pending", data
		_, err = col.Update(ctx, rec)
	}
	return err
}

// enqueueHookEvents runs before acknowledging ingress. Re-delivery is harmless.
func (c *Controller) enqueueHookEvents(ctx context.Context, items []store.InboxItem) error {
	lc := c.live.Load()
	for _, it := range items {
		ev := eventOf(it)
		for _, name := range slices.Sorted(maps.Keys(lc.cfg.Triggers)) {
			tr := lc.cfg.Triggers[name]
			if tr.Handler == "" || !triggerOn(tr, ev) {
				continue
			}
			program := lc.hooks[tr.Handler]
			if program == nil {
				return fmt.Errorf("trigger %s has no loaded hook", name)
			}
			payload := it.Admitted
			if len(payload) == 0 || string(payload) == "null" {
				payload = it.Payload
			}
			var object map[string]any
			if err := decodeHookJSON(payload, &object); err != nil {
				return fmt.Errorf("trigger %s: payload must be a JSON object: %w", name, err)
			}
			event := &automationEvent{Source: it.Source, Event: it.Kind, Delivery: it.DeliveryID, Repository: it.RepoOrTeam, Payload: object, Fields: ev.vars()}
			if err := c.putHook(ctx, "event/"+name+"/"+it.Source+"/"+it.DeliveryID, hookInvocation{Trigger: name, Entry: tr.Handler, Digest: program.digest, Event: event, Args: tr.Args}); err != nil {
				return err
			}
		}
	}
	c.kickHooks()
	return nil
}

func stepHookInput(t store.Ticket) (hookStep, error) {
	f, err := store.DecodeFields[store.StepFields](t)
	return hookStep{ID: f.Step, TicketID: t.ID, Pool: t.Pool, Status: string(t.Status), Resolution: t.Resolution, Outcome: f.Outcome, Output: f.StructuredOutput, Inputs: f.Inputs, Failures: f.Failures, Retries: f.Retries}, err
}

func (c *Controller) enqueueRunHooks(ctx context.Context) error {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunRunning, Limit: store.NoLimit})
	if err != nil {
		return err
	}
	for _, run := range runs {
		f, err := store.DecodeFields[store.RunFields](run)
		if err != nil {
			return err
		}
		a := f.Automation
		if a == nil || run.Status == store.StatusClosed {
			continue
		}
		children, err := store.Children(ctx, c.store, run.ID)
		if err != nil {
			return err
		}
		input := &hookRun{ID: run.ID, Workflow: f.Workflow, Key: f.Key, Project: run.Project, Inputs: f.Inputs, Vars: f.Vars, State: f.State, Steps: []hookStep{}}
		readSet := map[string]int64{run.ID: run.Revision}
		for _, child := range children {
			readSet[child.ID] = child.Revision
			if child.Kind != store.KindStep {
				continue
			}
			step, err := stepHookInput(child)
			if err != nil {
				return err
			}
			input.Steps = append(input.Steps, step)
		}
		in := hookInvocation{Run: run.ID, RunInput: input, ReadSet: readSet, Digest: a.Digest, ParentInvocation: metaString(run.Meta, "automation", "invocation")}
		key := run.ID + "/start"
		if a.Version == 0 {
			in.Entry = a.OnStart
		} else {
			for _, child := range children {
				if child.Kind != store.KindStep || !store.StepFinished(child) || a.Completed[child.ID] == child.Revision {
					continue
				}
				step, err := stepHookInput(child)
				if err != nil {
					return err
				}
				in.Entry, in.Step, in.StepRevision, in.StepInput = a.OnStepCompleted, child.ID, child.Revision, &step
				key = run.ID + "/step/" + child.ID + "/" + strconv.FormatInt(child.Revision, 10) + "/" + strconv.FormatInt(a.Version, 10)
				break
			}
		}
		if in.Entry != "" {
			if err := c.putHook(ctx, key, in); err != nil {
				return err
			}
		}
	}
	return nil
}

// One bounded worker evaluates hooks without tickMu, runMu, or controller mu.
func (c *Controller) kickHooks() {
	if c.ctx == nil || c.ctx.Err() != nil {
		return
	}
	c.hookPending.Store(true)
	if !c.hookRunning.CompareAndSwap(false, true) {
		return
	}
	c.wg.Go(func() {
		c.hookPending.Store(false)
		defer func() {
			c.hookRunning.Store(false)
			if c.hookPending.Swap(false) {
				c.kickHooks()
			}
		}()
		if err := c.enqueueRunHooks(c.ctx); err != nil {
			c.log.ErrorContext(c.ctx, "enqueue workflow hooks", "err", err)
			return
		}
		col, err := c.store.Collection(c.ctx, hookInvocations)
		if err != nil {
			return
		}
		if !c.hookRecovered.Load() {
			if err := c.recoverHookClaims(c.ctx, col); err != nil {
				c.log.ErrorContext(c.ctx, "recover hook claims", "err", err)
				return
			}
			c.hookRecovered.Store(true)
		}
		for range 64 {
			lease := config.Duration(c.config().Automation.Timeout, 30*time.Second) + time.Minute
			rec, err := col.Claim(c.ctx, store.RecordFilter{States: []string{"pending", "planned"}}, "controller-hooks", lease)
			if errors.Is(err, store.ErrNotFound) {
				return
			}
			if err != nil {
				c.log.ErrorContext(c.ctx, "claim hook", "err", err)
				return
			}
			c.executeHookAttempt(c.ctx, col, rec)
			if err := c.enqueueRunHooks(c.ctx); err != nil {
				return
			}
		}
	})
}

func (c *Controller) saveHook(ctx context.Context, col store.Collection, rec store.Record, in hookInvocation, state string) (store.Record, error) {
	// Prepared evidence was sanitized with the accepted credentials, which may have rotated.
	data, err := json.Marshal(in)
	if err != nil {
		return rec, err
	}
	rec.Value, rec.State = data, state
	rec.LeaseUntil = time.Time{}
	return col.Update(ctx, rec)
}

func (c *Controller) hookFailure(ctx context.Context, col store.Collection, rec store.Record, in hookInvocation, err error) {
	in.Error = c.hookAuditText(err.Error())
	in.Public = c.hookPublicError(in.Public, in.Error)
	_, saveErr := c.saveHook(ctx, col, rec, in, "failed")
	c.log.ErrorContext(ctx, "automation hook failed", "invocation", rec.Key, "run", in.Run, "trigger", in.Trigger, "err", in.Error, "record_error", saveErr)
	if in.Run != "" {
		c.runMu.Lock()
		t, getErr := c.store.Get(ctx, in.Run)
		if getErr == nil && t.Status != store.StatusClosed {
			f, decodeErr := store.DecodeFields[store.RunFields](t)
			if decodeErr == nil && f.State == store.RunRunning {
				_, _ = c.setRunState(ctx, "controller", t, f, store.RunNeedsHuman, "")
			}
		}
		c.runMu.Unlock()
		_, _ = store.AddComment(ctx, c.store, "controller", in.Run, "Automation invocation "+rec.Key+" failed: "+in.Error)
	}
}
