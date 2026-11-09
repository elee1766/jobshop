package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Trigger and scan evaluation lives in pkg/control/triggers (JOB-145). The
// controller keeps what writes: it starts the runs, files the tickets, runs
// trigger commands and records learnings, and it supplies the signals that
// read state other subsystems own. See docs/notes/JOB-145-triggers.md.

// triggerPack is the meta pack on tickets a trigger created.
const triggerPack = triggers.Pack

// findingMetaKey is meta.trigger.finding on the runs a scan trigger starts.
const findingMetaKey = triggers.FindingMetaKey

const (
	// needsHumanLabel marks tickets waiting on the operator.
	needsHumanLabel = triggers.NeedsHumanLabel
	// blockedLabel marks tickets that claim a blocker.
	blockedLabel = triggers.BlockedLabel
)

// triggerEngine is the trigger and scan engine, made on first use so
// controllers built in tests without Run get one too.
func (c *Controller) triggerEngine() *triggers.Engine {
	c.trigOnce.Do(func() { c.trig = triggers.New(triggerHost{c}) })
	return c.trig
}

// runTriggers fires every trigger whose conditions hold. Called each tick.
func (c *Controller) runTriggers(ctx context.Context, cfg *config.Config) {
	c.triggerEngine().Run(ctx, cfg)
}

// triggerHost is the controller as triggers.Host.
type triggerHost struct{ c *Controller }

func (h triggerHost) Config() *config.Config   { return h.c.config() }
func (h triggerHost) Now() time.Time           { return h.c.now() }
func (h triggerHost) Started() time.Time       { return h.c.started }
func (h triggerHost) Store() store.Store       { return h.c.store }
func (h triggerHost) Background() bool         { return h.c.backgroundChecks }
func (h triggerHost) Context() context.Context { return h.c.ctx }
func (h triggerHost) Go(fn func())             { h.c.wg.Go(fn) }
func (h triggerHost) Draining() bool           { return h.c.draining() }
func (h triggerHost) KickHooks()               { h.c.kickHooks() }

func (h triggerHost) Log() *slog.Logger { return h.c.log }

func (h triggerHost) Hooks() triggers.Callbacks { return h.c.hooks() }

func (h triggerHost) LockTick(ctx context.Context, who, trigger string) func() {
	return h.c.lockTickFor(ctx, who, trigger)
}

func (h triggerHost) RootSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return h.c.rootSpan(ctx, name, attrs...)
}

// TimedOut logs and counts an off-pass call that ran out its timeout
// (JOB-103).
func (h triggerHost) TimedOut(poller string, timeout time.Duration, args ...any) {
	c := h.c
	c.log.WarnContext(c.ctx, "off-pass call exceeded its timeout", append([]any{"poller", poller, "timeout", timeout}, args...)...)
	if c.metrics != nil {
		c.metrics.PollerTimeouts.Add(c.ctx, 1, metric.WithAttributes(telemetry.AttrPoller.String(poller)))
	}
}

// RunCheck runs a trigger check with sh in the root dir. Triggers have no
// pool: unisolated, through proc.Start for the process-group kill (#810).
func (h triggerHost) RunCheck(ctx context.Context, command string) (triggers.CommandRun, error) {
	var stdout bytes.Buffer
	var stderr strings.Builder
	ex, err := h.c.runIsolated(ctx, isolatedCmd{argv: shArgv(command), dir: h.c.root.Dir, stdout: &stdout, stderr: &stderr, name: "trigger-check"})
	run := triggers.CommandRun{Stdout: stdout.String(), Stderr: stderr.String()}
	if err == nil && (ex.ExitCode != 0 || ex.Signal != "") {
		run.Exit = exitText(ex)
	}
	return run, err
}

func (h triggerHost) EditsHold(ctx context.Context, w config.EditsWhen, now time.Time) (string, error) {
	return h.c.editsHold(ctx, w, now)
}

func (h triggerHost) Fire(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, key string, ev triggers.Event) (domain.Ticket, error) {
	return h.c.fireTriggerEvent(ctx, cfg, name, tr, key, ev)
}

func (h triggerHost) FireFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s triggers.Finding) (domain.Ticket, bool, error) {
	return h.c.fireFinding(ctx, cfg, name, tr, s)
}

func (h triggerHost) RecordFinding(ctx context.Context, s triggers.Finding, run string) {
	h.c.recordFinding(ctx, s, run)
}

func (h triggerHost) QuarantinedPools(cfg *config.Config) []triggers.Finding {
	return h.c.signalPoolQuarantined(cfg)
}

func (h triggerHost) JournalAt(session string) time.Time {
	_, at := host.JournalTail(h.c.root.SessionDir(session))
	return at
}

func (h triggerHost) WaitsOnQuestion(ctx context.Context, t domain.Ticket) bool {
	return h.c.waitsOnQuestion(ctx, t)
}

func (h triggerHost) ResourceWaits(ctx context.Context, now time.Time, after time.Duration) ([]triggers.Finding, error) {
	return h.c.signalResourceWait(ctx, now, after)
}

func (h triggerHost) Ingest(ctx context.Context, cfg *config.Config, signal string) ([]triggers.Finding, error) {
	return h.c.signalIngest(ctx, cfg, signal)
}

func (h triggerHost) LinearIssues(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]triggers.Finding, error) {
	return h.c.signalLinearIssues(ctx, cfg, sig)
}

func (h triggerHost) LinearIssueBusy(ctx context.Context, key string) (bool, error) {
	return h.c.linearSvc().HasOpenRun(ctx, key)
}

func (h triggerHost) Script(ctx context.Context, cfg *config.Config, name, entry string, timeout time.Duration, args map[string]any) (map[string]any, error) {
	res, err := h.c.callSync(ctx, cfg, name, entry, timeout, args)
	return res.Output, err
}

// signalPoolQuarantined: pools paused for crash-looping sessions. The
// subject carries the pause end, so each quarantine is its own finding and
// max_runs does not carry over from the last one.
func (c *Controller) signalPoolQuarantined(cfg *config.Config) []triggers.Finding {
	var out []triggers.Finding
	for _, p := range slices.Sorted(maps.Keys(cfg.Pools)) {
		if c.poolPaused(p) {
			c.mu.Lock()
			until := c.poolPausedUntil[p]
			_, permanent := c.poolQuarantined[p]
			why := c.poolPauseWhy(p)
			c.mu.Unlock()
			if permanent {
				continue // already escalated as "Pool <p> quarantined" (#146)
			}
			out = append(out, triggers.Finding{Signal: config.SignalPoolQuarantined, Subject: p + "@" + until.UTC().Format("20060102T150405Z"), Detail: "pool " + p + ": sessions crash-looping, paused until " + until.UTC().Format(time.RFC3339) + ": " + why})
		}
	}
	return out
}

// waitsOnQuestion: the ticket's last report is waiting on a question that
// is still open. The question itself is a needs-human ticket, so the
// needs_human signal covers it and the step or session is not idle.
func (c *Controller) waitsOnQuestion(ctx context.Context, t domain.Ticket) bool {
	rm, ok := reportOf(t)
	if !ok || rm.State != rpc.ReportWaiting || rm.Question == "" {
		return false
	}
	q, err := c.store.Get(ctx, rm.Question)
	return err == nil && q.Status != domain.StatusClosed
}

// metaString is meta[pack][key] when it is a string.
func metaString(raw json.RawMessage, pack, key string) string {
	return domain.MetaString(raw, pack, key)
}

// fireTriggerEvent fires tr; ev is the event that fired it, nil for a
// poll. Its fields are the event template var (empty on a poll).
func (c *Controller) fireTriggerEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, key string, ev triggers.Event) (domain.Ticket, error) {
	m := map[string]string{"name": name}
	if key != "" {
		// Prefixed with the trigger name so keys of different triggers
		// do not collide in the meta filter.
		m["key"] = name + "/" + key
	}
	meta, err := json.Marshal(map[string]map[string]string{triggerPack: m})
	if err != nil {
		return domain.Ticket{}, err
	}
	out, _, _ := strings.Cut(key, "\n")
	tv := map[string]any{config.CheckOutputVar: strings.TrimSpace(out), config.EventVar: ev.Vars()}
	if tr.Run != nil {
		return c.fireRun(ctx, cfg, tr.Run, m, tv)
	}
	if tr.Command != nil {
		return c.fireCommand(ctx, cfg, name, tr, meta, tv)
	}
	tt := *tr.Ticket
	if key != "" {
		tt.Body = strings.TrimSpace(tt.Body + "\n\nCheck output: " + key)
	}
	return store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: tt.Project, Kind: domain.KindTask, Title: tt.Title, Body: tt.Body,
		Priority: tt.Priority, Pool: tt.Pool, Labels: tt.Labels, Meta: meta,
	})
}

// fireRun starts the trigger's workflow run. meta.trigger goes on the run
// ticket, so the open run holds the trigger and its key dedups like a
// ticket's. A run already open under the same run key is not an error;
// the trigger stays held by that run only if it carries the trigger meta.
func (c *Controller) fireRun(ctx context.Context, cfg *config.Config, r *config.TriggerRun, m map[string]string, tv map[string]any) (domain.Ticket, error) {
	p := rpc.RunStartParams{Workflow: r.Workflow, Vars: map[string]string{}}
	for k, v := range r.Vars {
		s, err := r.Fill(v, tv)
		if err != nil {
			return domain.Ticket{}, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = s
	}
	k, err := r.Fill(r.Key, tv)
	if err != nil {
		return domain.Ticket{}, fmt.Errorf("run key: %w", err)
	}
	p.Key = k
	res, err := c.startRun(ctx, cfg, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return domain.Ticket{}, err
	}
	return res.(rpc.RunResult).Run, nil
}

// fireFinding starts the trigger's workflow for one finding with key
// <key_prefix>:<signal>:<subject>, or the trigger's run key filled with
// the scan vars. existing is true when an open run already had that key
// (one started by hand or by another run, JOB-33).
func (c *Controller) fireFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s triggers.Finding) (run domain.Ticket, existing bool, err error) {
	tv := map[string]any{config.ScanSignalVar: s.Signal, config.ScanSubjectVar: s.Subject, config.ScanDetailVar: s.Detail}
	key := s.KeyWith(tr.ScanKeyPrefix())
	p := rpc.RunStartParams{Workflow: tr.Run.Workflow, Vars: map[string]string{}, Key: key}
	if tr.Run.Key != "" {
		if p.Key, err = tr.Run.Fill(tr.Run.Key, tv); err != nil {
			return domain.Ticket{}, false, fmt.Errorf("run key: %w", err)
		}
	}
	// A finding about an issue titles its run with the issue's (JOB-133).
	if p.Title, err = c.findingTitle(ctx, cfg, tr, s, p.Workflow, p.Key); err != nil {
		return domain.Ticket{}, false, err
	}
	for k, v := range tr.Run.Vars {
		out, err := tr.Run.Fill(v, tv)
		if err != nil {
			return domain.Ticket{}, false, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = out
	}
	m := map[string]string{"name": name, "key": name + "/" + key}
	if tr.Run.Key == "" {
		// The #704 hold covers runs keyed on the finding. A run with its
		// own key (pr-<n>, JOB-33) never took part: leave it out.
		m[findingMetaKey] = s.ID()
	}
	res, err := c.startRun(ctx, cfg, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return domain.Ticket{}, false, err
	}
	rr := res.(rpc.RunResult)
	return rr.Run, rr.Existing, nil
}

// commandLabel marks the record of a trigger command run.
const commandLabel = "trigger-command"

// fireCommand runs tr.Command. The run is recorded as a message ticket
// carrying the trigger meta, so it holds the trigger while the command
// runs and dedups its key. Success closes it; failure keeps it open as
// needs-human with the output, so a broken command is reported once
// and the trigger waits for a human to close it.
func (c *Controller) fireCommand(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, meta json.RawMessage, tv map[string]any) (domain.Ticket, error) {
	cm := tr.Command
	script, err := config.Fill(cm.Run, tv)
	if err != nil {
		return domain.Ticket{}, fmt.Errorf("command: %w", err)
	}
	fields, err := json.Marshal(domain.MessageFields{To: questionTo, From: "controller"})
	if err != nil {
		return domain.Ticket{}, err
	}
	// A root with several projects has no default; the record only needs
	// one to live in.
	project := cfg.DefaultProject
	if projects := slices.Sorted(maps.Keys(cfg.Projects)); project == "" && len(projects) > 0 {
		project = projects[0]
	}
	if project == "" {
		return domain.Ticket{}, errors.New("command: no project to record the run in")
	}
	t, err := store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: project, Kind: domain.KindMessage, Title: "Trigger " + name + ": command",
		Body: "```\n" + script + "\n```\n", Labels: []string{commandLabel}, Meta: meta, Fields: fields,
	})
	if err != nil {
		return domain.Ticket{}, err
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
func (c *Controller) finishCommand(ctx context.Context, name string, t domain.Ticket, script string, timeout time.Duration) {
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

func (c *Controller) triggerList(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "trigger-list", "trigger list"); err != nil {
		return nil, err
	}
	// Triggers are evaluated in parallel: most of the time is check
	// commands (often gh), and one after another they made the Desk
	// pools view take ~5s (#522). Evaluation only reads, so order does
	// not matter.
	cfg := c.config()
	trs := cfg.Triggers
	names := slices.Sorted(maps.Keys(trs))
	out := make([]rpc.TriggerInfo, len(names))
	errs := make([]error, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Go(func() { out[i], errs[i] = c.triggerEngine().Info(ctx, cfg, name, trs[name]) })
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	labels := map[string][]string{}
	for _, in := range cfg.Instances {
		for _, name := range in.Triggers {
			labels[name] = config.InstanceLabels(in)
		}
	}
	for i := range out {
		out[i].Labels = labels[out[i].Name]
	}
	return out, nil
}
