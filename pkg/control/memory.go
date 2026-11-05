package control

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/memory"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Local memory (#592). Memory lives in <root>/memory on the Jobshop host,
// outside any checkout. The curator writes it through memory.write, which
// records each write as a memory edit holding the text it replaced; the
// curator gate restores that text when the edit did not help. Prompt
// fixes stay PRs against packs/.

// memoryDir is the local memory directory.
func (c *Controller) memoryDir() string { return c.root.Path("memory") }

// memoryCap is the cap of the first prompt rule reading memoryDir, else
// the default.
func (c *Controller) memoryCap() int {
	dir := c.memoryDir()
	for _, r := range c.config().PromptRules {
		if r.Memory != nil && r.Memory.Dir != "" && c.root.Path(r.Memory.Dir) == dir && r.Memory.Cap > 0 {
			return r.Memory.Cap
		}
	}
	return memory.DefaultCap
}

// memoryRef is what a local write puts in its learnings' pr field, so the
// curator's settle pass and `learn list` show where each one went.
func memoryRef(id int64) string { return fmt.Sprintf("memory:%d", id) }

func (c *Controller) memoryWrite(ctx context.Context, sc scope, p rpc.MemoryWriteParams) (any, error) {
	if err := c.requireGrant(sc, "memory-write", "memory write"); err != nil {
		return nil, err
	}
	rel, err := memory.Target(p.Scope)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", store.ErrInvalid, err)
	}
	if strings.TrimSpace(p.Text) == "" {
		return nil, fmt.Errorf("%w: memory write needs text; to drop a file write one line saying so", store.ErrInvalid)
	}
	cited := make([]store.Learning, 0, len(p.Learnings))
	for _, id := range p.Learnings {
		l, err := c.store.GetLearning(ctx, id)
		if err != nil {
			return nil, err
		}
		cited = append(cited, l)
	}
	dir := c.memoryDir()
	c.memMu.Lock()
	defer c.memMu.Unlock()
	before, existed, err := memory.ReadRaw(dir, rel)
	if err != nil {
		return nil, err
	}
	now := c.now()
	pools := c.editPools(ctx, cited)
	base, err := c.measure(ctx, cited, pools, now.Add(-gateBaseline), now)
	if err != nil {
		return nil, err
	}
	if err := memory.Write(dir, rel, p.Text, c.memoryCap()); err != nil {
		if errors.Is(err, memory.ErrOverCap) {
			return nil, fmt.Errorf("%w: %w", store.ErrInvalid, err)
		}
		return nil, err
	}
	e := store.MemoryEdit{Path: rel, Before: before, Existed: existed, Pools: pools, MergedAt: now, Baseline: base}
	for _, l := range cited {
		e.Learnings = append(e.Learnings, l.ID)
	}
	out, err := c.store.AddMemoryWrite(ctx, e)
	if err != nil {
		// Keep the file and the record in step: undo the write.
		_ = memory.Restore(dir, rel, before, existed)
		return nil, err
	}
	for _, id := range out.Learnings {
		if _, err := c.store.UpdateLearning(ctx, id, store.LearningApplied, memoryRef(out.ID)); err != nil {
			return nil, err
		}
	}
	c.log.Info("memory write", "edit", out.ID, "path", rel, "learnings", out.Learnings, "session", sc.session)
	return out, nil
}

func (c *Controller) memoryShow(_ context.Context, _ scope, p rpc.MemoryShowParams) (any, error) {
	dir := c.memoryDir()
	want := ""
	if p.Scope != "" {
		rel, err := memory.Target(p.Scope)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", store.ErrInvalid, err)
		}
		want = rel
	}
	files, err := memory.List(dir)
	if err != nil {
		return nil, err
	}
	out := rpc.MemoryShowResult{Dir: dir, Files: []rpc.MemoryFile{}}
	for _, f := range files {
		if want != "" && f.Path != want {
			continue
		}
		out.Files = append(out.Files, rpc.MemoryFile{Path: f.Path, Chars: f.Chars, Text: f.Text})
	}
	return out, nil
}

// restoreEdit undoes a local write the gate found not helping: the file
// gets its previous text back, or is removed when the write created it,
// and the cited learnings become rejected so the curator does not write
// the same thing again. A newer write to the same path wins: restoring
// under it would drop that write, so the restore is skipped and logged.
func (c *Controller) restoreEdit(ctx context.Context, e store.MemoryEdit) error {
	if e.Path == "" {
		return nil
	}
	newer, err := c.store.MemoryEdits(ctx, "", 0)
	if err != nil {
		return err
	}
	for _, n := range newer {
		if n.Path == e.Path && n.ID > e.ID {
			c.log.Warn("memory edit not restored: a newer write replaced it", "edit", e.ID, "newer", n.ID, "path", e.Path)
			return c.rejectCited(ctx, e)
		}
	}
	c.memMu.Lock()
	err = memory.Restore(c.memoryDir(), e.Path, e.Before, e.Existed)
	c.memMu.Unlock()
	if err != nil {
		return err
	}
	c.log.Info("memory edit restored", "edit", e.ID, "path", e.Path, "existed", e.Existed)
	return c.rejectCited(ctx, e)
}

func (c *Controller) rejectCited(ctx context.Context, e store.MemoryEdit) error {
	for _, id := range e.Learnings {
		if _, err := c.store.UpdateLearning(ctx, id, store.LearningRejected, memoryRef(e.ID)); err != nil {
			return err
		}
	}
	return nil
}
