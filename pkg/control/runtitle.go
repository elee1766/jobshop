package control

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Run titles (JOB-133). A run of a workflow with an issue or PR var is
// titled with that issue or PR's own title, read when the run starts, so
// Desk and trace say what the work is. The title is cosmetic: dedup keys,
// workflow identity and meta.github / meta.linear (the navigable
// references) do not depend on it, and a failed read falls back to a
// label naming the source instead of stopping the start.

// runTitleTimeout bounds one title read, made off the tick like a check
// (JOB-103).
const runTitleTimeout = 20 * time.Second

// maxRunTitle is the longest run title in runes.
const maxRunTitle = 200

// runSource is the issue or PR a run works on.
type runSource struct {
	kind string // runSourceIssue, runSourcePR or runSourceLinear
	repo string // owner/name; empty for Linear
	ref  string // issue or PR number, or the Linear key
}

const (
	runSourceIssue  = "issue"
	runSourcePR     = "pr"
	runSourceLinear = "linear"
)

// runSourceOf is the source of a run of w with vars, if w has one.
func runSourceOf(w config.Workflow, vars map[string]any) (runSource, bool) {
	name, kind := w.Issue, runSourceIssue
	if w.PR != "" {
		name, kind = w.PR, runSourcePR
	}
	if name == "" {
		return runSource{}, false
	}
	ref := strings.TrimSpace(fmt.Sprint(vars[name]))
	if w.Tracker == config.TrackerLinear && kind == runSourceIssue {
		return runSource{kind: runSourceLinear, ref: strings.ToUpper(ref)}, ref != ""
	}
	ref = strings.TrimPrefix(ref, "#")
	if n, err := strconv.Atoi(ref); err != nil || n <= 0 || w.Repo == "" {
		return runSource{}, false
	}
	return runSource{kind: kind, repo: w.Repo, ref: ref}, true
}

// findingSource is the issue a scan finding is about: an issue_no_run
// finding's subject is a GitHub issue number of the signal's repo, a
// linear_issue_no_run finding's a Linear key. Other signals have none.
func findingSource(tr config.Trigger, s triggers.Finding) (runSource, bool) {
	for _, sig := range tr.When.Scan {
		if sig.Signal != s.Signal {
			continue
		}
		switch s.Signal {
		case config.SignalIssueNoRun:
			if n, err := strconv.Atoi(s.Subject); err == nil && n > 0 && sig.Repo != "" {
				return runSource{kind: runSourceIssue, repo: sig.Repo, ref: s.Subject}, true
			}
		case config.SignalLinearIssueNoRun:
			return runSource{kind: runSourceLinear, ref: strings.ToUpper(s.Subject)}, true
		}
	}
	return runSource{}, false
}

// label names the source for a run whose title could not be read.
func (s runSource) label() string {
	switch s.kind {
	case runSourceLinear:
		return "Linear " + s.ref
	case runSourcePR:
		return "PR " + s.repo + "#" + s.ref
	}
	return "Issue " + s.repo + "#" + s.ref
}

func (s runSource) key() string { return s.kind + "\x00" + s.repo + "\x00" + s.ref }

// sourceTitle is the title of src, or its label when the read fails. A
// start from a pass returns triggers.ErrCheckPending and the replay has the read
// (JOB-103).
func (c *Controller) sourceTitle(ctx context.Context, cfg *config.Config, src runSource) (string, error) {
	res, err := triggers.OffTick(ctx, triggers.Need{
		Key: "runtitle\x00" + src.key(), Poller: "run_title", Timeout: runTitleTimeout,
		Run: func(ctx context.Context) triggers.Result {
			t, err := c.readSourceTitle(c.linearSvc().PollerCtx(ctx), cfg, src)
			return triggers.Result{Out: t, Err: err}
		},
	})
	if err != nil {
		return "", err
	}
	if res.Err != nil {
		c.log.WarnContext(ctx, "run title: read failed, using the source label", "source", src.label(), "err", res.Err)
		return src.label(), nil
	}
	if t := cleanTitle(res.Out); t != "" {
		return t, nil
	}
	return src.label(), nil
}

// readSourceTitle reads the title of src as the identity covering it.
func (c *Controller) readSourceTitle(ctx context.Context, cfg *config.Config, src runSource) (string, error) {
	if src.kind == runSourceLinear {
		cl, err := c.linearClientIn(cfg, src.ref)
		if err != nil {
			return "", err
		}
		return cl.IssueTitle(ctx, src.ref)
	}
	cl, err := c.githubForRepo(cfg, "", src.repo)
	if err != nil {
		return "", err
	}
	path := "/repos/" + src.repo + "/issues/" + src.ref
	if src.kind == runSourcePR {
		path = "/repos/" + src.repo + "/pulls/" + src.ref
	}
	var v struct {
		Title string `json:"title"`
	}
	if err := cl.Get(ctx, path, &v); err != nil {
		return "", err
	}
	return v.Title, nil
}

// cleanTitle puts s on one line and caps its length.
func cleanTitle(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > maxRunTitle {
		s = string(r[:maxRunTitle-1]) + "…"
	}
	return s
}

// planTitle is the title of the run p starts: p.Title, else its source's.
// "" keeps the default title. A start that will return the open run for
// its key reads nothing.
func (c *Controller) planTitle(ctx context.Context, cfg *config.Config, w config.Workflow, vars map[string]any, p rpc.RunStartParams) (string, error) {
	if t := cleanTitle(p.Title); t != "" {
		return t, nil
	}
	src, ok := runSourceOf(w, vars)
	if !ok {
		return "", nil
	}
	if p.Key != "" {
		if _, err := store.OpenRun(ctx, c.store, p.Workflow, p.Key); err == nil {
			return "", nil
		}
	}
	return c.sourceTitle(ctx, cfg, src)
}

// findingTitle is the title of the run a scan trigger starts for s: its
// issue's own, "" when the finding has no issue (stuck work keeps its
// descriptive default) or the key already has an open run.
func (c *Controller) findingTitle(ctx context.Context, cfg *config.Config, tr config.Trigger, s triggers.Finding, workflow, key string) (string, error) {
	src, ok := findingSource(tr, s)
	if !ok {
		return "", nil
	}
	if _, err := store.OpenRun(ctx, c.store, workflow, key); err == nil {
		return "", nil
	}
	return c.sourceTitle(ctx, cfg, src)
}
