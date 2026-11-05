package control

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// questionLabel marks a ticket where a worker asks a human (#185, #186).
const questionLabel = "question"

// proposedLabel marks the planner's proposed next target, which waits
// on a human to approve (#488).
const proposedLabel = "proposed"

// attention lists only what waits on a human, across all projects,
// oldest first (#185). Empty when healthy.
func (c *Controller) attention(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "attention", "attention"); err != nil {
		return nil, err
	}
	items, err := c.attentionItems(ctx)
	if err != nil {
		return nil, err
	}
	return rpc.AttentionResult{Items: items}, nil
}

func (c *Controller) attentionItems(ctx context.Context) ([]rpc.AttentionItem, error) {
	items := []rpc.AttentionItem{}
	// Runs and pools an open escalation already stands for are listed once,
	// as the escalation ticket.
	escalatedRuns, escalatedPools := map[string]bool{}, map[string]bool{}
	seen := map[string]bool{}
	for _, label := range []string{needsHumanLabel, questionLabel, proposedLabel} {
		ts, err := c.labeledOpen(ctx, label)
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			if id := metaString(t.Meta, runMetaPack, "id"); id != "" {
				escalatedRuns[id] = true
			}
			if p := metaString(t.Meta, poolMetaPack, "name"); p != "" {
				escalatedPools[p] = true
			}
			if seen[t.ID] {
				continue
			}
			seen[t.ID] = true
			kind := rpc.AttentionTicket
			switch {
			case label == proposedLabel && !slices.Contains(t.Labels, needsHumanLabel):
				kind = rpc.AttentionDecision
			case (label == questionLabel && !slices.Contains(t.Labels, needsHumanLabel)) || metaString(t.Meta, questionMetaPack, "ticket") != "":
				kind = rpc.AttentionQuestion
			}
			items = append(items, ticketItem(kind, t))
		}
	}
	runs, err := c.store.ListRuns(ctx, "", store.RunNeedsHuman, math.MaxInt32)
	if err != nil {
		return nil, err
	}
	// Waits past their timeout (#381), with reason and owner. Listed even
	// though the timeout labeled them needs-human, so the title says why.
	waits, err := c.waiting(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, t := range waits {
		w := waitOf(t)
		if w.Timeout.IsZero() || c.now().Before(w.Timeout) {
			continue
		}
		it := ticketItem(rpc.AttentionWait, t)
		it.Title = fmt.Sprintf("%s: wait past timeout: %s", t.Title, w.Reason)
		it.Owner, it.Since = w.Owner, w.Timeout
		if i := slices.IndexFunc(items, func(x rpc.AttentionItem) bool { return x.ID == t.ID }); i >= 0 {
			items[i] = it
		} else {
			items = append(items, it)
		}
	}
	for _, r := range runs {
		if !escalatedRuns[r.ID] && r.Status != store.StatusClosed {
			items = append(items, ticketItem(rpc.AttentionRun, r))
		}
	}
	res, err := c.resourceAttention(ctx)
	if err != nil {
		return nil, err
	}
	items = append(items, res...)
	ing, err := c.ingestAttention(ctx)
	if err != nil {
		return nil, err
	}
	items = append(items, ing...)
	items = append(items, c.journalLagAttention()...)
	stale, err := c.staleAttention(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range stale {
		// A needs-human or question ticket with no pool is listed once,
		// as itself.
		if !slices.ContainsFunc(items, func(x rpc.AttentionItem) bool { return x.ID == it.ID }) {
			items = append(items, it)
		}
	}
	c.mu.Lock()
	if c.configErr != "" {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionConfig, ID: "config", Title: "serving last good config, a pack failed to load: " + c.configErr, Since: c.configErrAt})
	}
	if c.watchErr != "" {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionConfigInvalid, ID: "config-invalid", Title: "a watched config change failed to load, running config kept: " + c.watchErr, Since: c.watchErrAt})
	}
	for p, why := range c.poolQuarantined {
		if !escalatedPools[p] {
			items = append(items, rpc.AttentionItem{Kind: rpc.AttentionPool, ID: p, Title: "pool " + p + " quarantined: " + why, Since: c.started})
		}
	}
	c.mu.Unlock()
	slices.SortStableFunc(items, func(a, b rpc.AttentionItem) int {
		if n := a.Since.Compare(b.Since); n != 0 {
			return n
		}
		return strings.Compare(a.ID, b.ID)
	})
	return items, nil
}

// ticketItem lists t. A run waits since it last changed (when it went
// needs_human), a ticket since it was filed.
func ticketItem(kind string, t store.Ticket) rpc.AttentionItem {
	it := rpc.AttentionItem{Kind: kind, ID: t.ID, Project: t.Project, Title: t.Title, Since: t.CreatedAt}
	if kind == rpc.AttentionRun {
		it.Since = t.UpdatedAt
	}
	var m struct {
		GitHub struct {
			Repo  string `json:"repo"`
			Issue any    `json:"issue"`
			PR    any    `json:"pr"`
		} `json:"github"`
		Linear struct {
			Issue string `json:"issue"`
		} `json:"linear"`
	}
	if len(t.Meta) > 0 && json.Unmarshal(t.Meta, &m) == nil {
		it.Repo = m.GitHub.Repo
		switch {
		case m.Linear.Issue != "":
			it.Ref = m.Linear.Issue
		case m.GitHub.Issue != nil:
			it.Ref = fmt.Sprintf("#%v", m.GitHub.Issue)
		case m.GitHub.PR != nil:
			it.Ref = fmt.Sprintf("PR #%v", m.GitHub.PR)
		}
	}
	return it
}
