package control

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/knowledge/memory"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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

func (c *Controller) memoryWrite(ctx context.Context, sc scope, p rpc.MemoryWriteParams) (any, error) {
	if err := c.requireGrant(sc, "memory-write", "memory write"); err != nil {
		return nil, err
	}
	rel, err := memory.Target(p.Scope)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
	}
	if strings.TrimSpace(p.Text) == "" {
		return nil, fmt.Errorf("%w: memory write needs text; to drop a file write one line saying so", domain.ErrInvalid)
	}
	cited, err := knowledge.GetLearnings(ctx, c.store, p.Learnings)
	if err != nil {
		return nil, err
	}
	now := c.now()
	out, err := c.localMemory.Write(ctx, c.store, knowledge.LocalWrite{
		Dir: c.memoryDir(), Rel: rel, Text: p.Text, Cap: c.memoryCap(),
		Cited: cited, Pools: c.editPools(ctx, cited), Now: now, Baseline: gateBaseline,
	})
	if err != nil {
		if knowledge.IsOverCap(err) {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
		}
		return nil, err
	}
	c.log.InfoContext(ctx, "memory write", "edit", out.ID, "path", rel, "learnings", out.Learnings, "session", sc.session)
	return out, nil
}

func (c *Controller) memoryShow(_ context.Context, _ scope, p rpc.MemoryShowParams) (any, error) {
	dir := c.memoryDir()
	want := ""
	if p.Scope != "" {
		rel, err := memory.Target(p.Scope)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
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
func (c *Controller) restoreEdit(ctx context.Context, e domain.MemoryEdit) error {
	if e.Path == "" {
		return nil
	}
	res, err := c.localMemory.Restore(ctx, c.store, c.memoryDir(), e)
	if err != nil {
		return err
	}
	if res.Newer != 0 {
		c.log.WarnContext(ctx, "memory edit not restored: a newer write replaced it", "edit", e.ID, "newer", res.Newer, "path", e.Path)
		return nil
	}
	c.log.InfoContext(ctx, "memory edit restored", "edit", e.ID, "path", e.Path, "existed", e.Existed)
	return nil
}
