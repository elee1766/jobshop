package stepexec

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/execution/sandbox"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// RunDir is a run's scratch dir for its command steps under root: the
// sandbox's own writable dir and the parent of the run worktree (#419).
func RunDir(root, runID string) string {
	return filepath.Join(root, "runs", runID)
}

// Place is where and how a command step runs.
type Place struct {
	Pool string
	Dir  string
	// Own is the run dir when the step is sandboxed or has a worktree.
	Own string
}

// Place resolves a command step's pool and working dir: the run
// worktree when the step asks for one, else the pool's dir, else its
// repo, else the root. It makes the worktree.
func (e Executor) Place(cfg *config.Config, s config.Step, sf domain.StepFields, t Templates) (Place, error) {
	pl := Place{Dir: e.Root}
	if s.Pool == "" {
		return pl, nil
	}
	name, err := cfg.ResolvePool(&t.W, s.Pool, t.Vars)
	if err != nil {
		return pl, err
	}
	pl.Pool = name
	p := cfg.Pools[name]
	switch {
	case p.Dir != "":
		pl.Dir = p.Dir
	case p.Repo != "":
		pl.Dir = p.Repo
	}
	if s.Worktree || cfg.SandboxFor(name).Backend != "" {
		pl.Own = RunDir(e.Root, sf.Run)
		if err := os.MkdirAll(pl.Own, 0o700); err != nil {
			return pl, err
		}
	}
	if s.Worktree {
		if p.Repo == "" {
			return pl, fmt.Errorf("worktree: pool %s has no repo", name)
		}
		pl.Dir = filepath.Join(pl.Own, "work")
		if err := e.Host.Worktree(p.Repo, p.Ref, pl.Dir, "jobshop run "+sf.Run, "pool "+name); err != nil {
			return pl, err
		}
	}
	return pl, nil
}

// RemoveRunDir deletes a run's worktree and scratch dir. gitMu
// serializes git commands on pool repos, as for addWorktree.
func RemoveRunDir(ctx context.Context, log *slog.Logger, gitMu sync.Locker, root, runID string) {
	dir := RunDir(root, runID)
	work := filepath.Join(dir, "work")
	if gd := GitCommonDir(work); gd != "" {
		gitMu.Lock()
		// Locked at add; two --force override the lock.
		if out, err := exec.CommandContext(ctx, "git", "--git-dir", gd, "worktree", "remove", "--force", "--force", work).CombinedOutput(); err != nil {
			log.WarnContext(ctx, "git worktree remove", "run", runID, "err", err, "out", strings.TrimSpace(string(out)))
		}
		gitMu.Unlock()
	}
	if err := os.RemoveAll(dir); err != nil {
		log.WarnContext(ctx, "remove run dir", "run", runID, "err", err)
	}
}

// GitCommonDir is repo's shared git dir, or "" if git does not know it.
func GitCommonDir(repo string) string { return sandbox.GitCommonDir(repo) }

// StaleRunDirs lists the run dirs under root whose run is closed or gone:
// a close that raced a running attempt, or a restart between close and
// removal.
func StaleRunDirs(ctx context.Context, st store.Store, root string) []string {
	ents, err := os.ReadDir(filepath.Join(root, "runs"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		t, err := st.Get(ctx, e.Name())
		if errors.Is(err, domain.ErrNotFound) || (err == nil && t.Status == domain.StatusClosed) {
			out = append(out, e.Name())
		}
	}
	return out
}
