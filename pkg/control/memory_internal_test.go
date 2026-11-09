package control

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/knowledge/memory"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestMemoryWrite: memory.write writes under <root>/memory, records the
// edit with the previous text, marks the cited learnings applied, and
// refuses a file over the cap without touching it (#592).
func TestMemoryWrite(t *testing.T) {
	t.Parallel()
	c, _, dir := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	l, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningWorkaround, Summary: "retry gh on 502", Scope: "pool:dev"})
	require.NoError(t, err)
	id := l.(domain.Learning).ID

	out, err := c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "global", Text: "Retry gh on 502.\n", Learnings: []int64{id}})
	require.NoError(t, err)
	e := out.(domain.MemoryEdit)
	assert.Equal(t, filepath.Join("memory", "global.md"), e.Path)
	assert.False(t, e.Existed)
	assert.Empty(t, e.PR)
	assert.Equal(t, []string{"dev"}, e.Pools)
	assert.Equal(t, domain.EditPending, e.State)
	b, err := os.ReadFile(filepath.Join(dir, "memory", "memory", "global.md"))
	require.NoError(t, err)
	assert.Equal(t, "Retry gh on 502.\n", string(b))
	got, err := knowledge.GetLearning(ctx, c.store, id)
	require.NoError(t, err)
	assert.Equal(t, domain.LearningApplied, got.Status)
	assert.Equal(t, knowledge.MemoryRef(e.ID), got.PR)

	// A second write records the first text as before.
	out, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "global", Text: "Retry gh twice on 502.\n"})
	require.NoError(t, err)
	e2 := out.(domain.MemoryEdit)
	assert.True(t, e2.Existed)
	assert.Equal(t, "Retry gh on 502.\n", e2.Before)

	// Over the cap: refused, file unchanged, nothing recorded.
	_, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "global", Text: strings.Repeat("x", memory.DefaultCap+1)})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "cap")
	b, err = os.ReadFile(filepath.Join(dir, "memory", "memory", "global.md"))
	require.NoError(t, err)
	assert.Equal(t, "Retry gh twice on 502.\n", string(b))
	edits, err := knowledge.ListEdits(ctx, c.store, "", 0)
	require.NoError(t, err)
	assert.Len(t, edits, 2)

	// Bad scopes and sessions without the grant are refused.
	for _, s := range []string{"../x", "repo/../../etc", "pool/", "other"} {
		_, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: s, Text: "x"})
		require.ErrorIs(t, err, domain.ErrInvalid, s)
	}
	_, err = c.memoryWrite(ctx, scope{pool: "dev", session: "s1"}, rpc.MemoryWriteParams{Scope: "global", Text: "x"})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	res, err := c.memoryShow(ctx, admin, rpc.MemoryShowParams{Scope: "global"})
	require.NoError(t, err)
	sh := res.(rpc.MemoryShowResult)
	assert.Equal(t, filepath.Join(dir, "memory"), sh.Dir)
	require.Len(t, sh.Files, 1)
	assert.Equal(t, "Retry gh twice on 502.\n", sh.Files[0].Text)
}

// TestMemoryWriteRestore: the gate restores the previous text of a write
// that did not help, and removes a file the write created (#592).
func TestMemoryWriteRestore(t *testing.T) {
	t.Parallel()
	c, clk, dir := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	repoFile := filepath.Join(dir, "memory", "memory", "repo", "o__r.md")
	poolFile := filepath.Join(dir, "memory", "memory", "pool", "dev.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(repoFile), 0o755))
	require.NoError(t, os.WriteFile(repoFile, []byte("old line\n"), 0o644))

	// The cited problem recurs after each write, so both are not_helped.
	step := mkTicket(t, c, store.NewTicket{Kind: domain.KindTask, Title: "x", Pool: "dev"})
	l, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "go vet failed on shadowed err variable", Ticket: step.ID})
	require.NoError(t, err)
	id := l.(domain.Learning).ID
	clk.advance(time.Hour)
	_, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "repo/o__r", Text: "new line\n", Learnings: []int64{id}})
	require.NoError(t, err)
	_, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "pool/dev", Text: "pool line\n", Learnings: []int64{id}})
	require.NoError(t, err)
	clk.advance(time.Hour)
	for i := range 2 {
		res, err := c.startRun(ctx, c.config(), admin, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": fmt.Sprint(i)}, Key: fmt.Sprint("m", i)}, nil)
		require.NoError(t, err)
		if i == 0 {
			s := res.(rpc.RunResult).Steps[0]
			_, err = c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: "go vet failed again on shadowed err variable", Ticket: s.ID})
			require.NoError(t, err)
		}
	}
	clk.advance(time.Minute)

	out, err := c.learnEvaluate(ctx, admin, rpc.LearnEvaluateParams{Runs: 2, Age: 72 * time.Hour})
	require.NoError(t, err)
	res := out.([]domain.MemoryEdit)
	require.Len(t, res, 2)
	for _, e := range res {
		assert.Equal(t, domain.EditNotHelped, e.State, e.Path)
	}
	b, err := os.ReadFile(repoFile)
	require.NoError(t, err)
	assert.Equal(t, "old line\n", string(b), "previous text restored")
	_, err = os.Stat(poolFile)
	assert.ErrorIs(t, err, os.ErrNotExist, "a file the write created is removed")
	got, err := knowledge.GetLearning(ctx, c.store, id)
	require.NoError(t, err)
	assert.Equal(t, domain.LearningRejected, got.Status)
	assert.Contains(t, got.Gate, "not_helped: ")
}
