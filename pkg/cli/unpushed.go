package cli

import (
	"fmt"
	"os/exec"
	"strings"
)

// unpushedCommits lists commits on the current branch of the git work
// tree at dir that no remote branch has (JOB-70). A session that
// releases its ticket can be ended and its worktree removed, so a commit
// only it has is easy to lose. Nil outside a work tree, on a detached
// HEAD, or when git is missing: the check is advice, not a gate on
// non-git work.
func unpushedCommits(dir string) []string {
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}
	if b, err := git("symbolic-ref", "-q", "--short", "HEAD"); err != nil || b == "" {
		return nil
	}
	out, err := git("log", "--format=%h %s", "HEAD", "--not", "--remotes")
	if err != nil || out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// checkPushed refuses a release or handoff from a work tree with
// unpushed commits unless force is set.
func checkPushed(dir, verb string, force bool) error {
	if force {
		return nil
	}
	c := unpushedCommits(dir)
	if len(c) == 0 {
		return nil
	}
	if len(c) > 5 {
		c = append(c[:5], fmt.Sprintf("... and %d more", len(c)-5))
	}
	return fmt.Errorf("%s: the current branch has commits no remote has:\n  %s\npush them first (a released session may end and remove this worktree), or pass --force",
		verb, strings.Join(c, "\n  "))
}
