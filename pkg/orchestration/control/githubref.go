package control

import (
	"context"
	"encoding/json"
	"math"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// backfillGitHubRepo adds meta.github.repo to runs created before work
// references carried the repo (#183), and to their steps and escalation
// tickets. The repo comes from the run's workflow config: the database
// does not know it, so this is a controller pass, not a SQL migration.
// Runs whose workflow is gone or names no repo are left alone. It is
// idempotent and runs at every start.
func (c *Controller) backfillGitHubRepo(ctx context.Context) error {
	cfg := c.config()
	runs, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun, Limit: math.MaxInt32})
	if err != nil {
		return err
	}
	n := 0
	for _, run := range runs {
		if !needsRepo(run.Meta) {
			continue
		}
		f, err := domain.DecodeFields[domain.RunFields](run)
		if err != nil {
			continue
		}
		repo := cfg.Workflows[f.Workflow].Repo
		if repo == "" {
			continue
		}
		related := []domain.Ticket{run}
		steps, err := store.Children(ctx, c.store, run.ID)
		if err != nil {
			return err
		}
		related = append(related, steps...)
		esc, err := c.store.List(ctx, store.Filter{MetaPack: runMetaPack, MetaKey: "id", MetaValue: run.ID})
		if err != nil {
			return err
		}
		related = append(related, esc...)
		for _, t := range related {
			mb, ok := withRepo(t.Meta, repo)
			if !ok {
				continue
			}
			if _, err := c.store.Update(ctx, "controller", t.ID, store.Update{Meta: mb, IfRevision: t.Revision}); err != nil {
				c.log.WarnContext(ctx, "backfill github repo", "ticket", t.ID, "err", err)
				continue
			}
			n++
		}
	}
	if n > 0 {
		c.log.InfoContext(ctx, "backfilled meta.github.repo", "tickets", n)
	}
	return nil
}

// needsRepo reports whether meta has a github object without a repo.
func needsRepo(meta json.RawMessage) bool {
	_, ok := withRepo(meta, "x/x")
	return ok
}

// withRepo sets meta.github.repo when meta.github exists and has none.
func withRepo(meta json.RawMessage, repo string) (json.RawMessage, bool) {
	var m map[string]map[string]any
	if len(meta) == 0 || json.Unmarshal(meta, &m) != nil {
		return nil, false
	}
	g, ok := m["github"]
	if !ok || g["repo"] != nil {
		return nil, false
	}
	g["repo"] = repo
	mb, err := json.Marshal(m)
	return mb, err == nil
}
