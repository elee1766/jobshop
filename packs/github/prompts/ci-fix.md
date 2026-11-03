Step `fix` of a `ci-repair` run: master CI failed at ${sha}.

The `master-red` trigger started this run for that SHA; a new failing
SHA starts a new run.

1. Read the failure: `gh run list --branch master --commit ${sha}
   --json databaseId,conclusion,name`, then `gh run view <id> --log-failed`.
   If master is already green on a newer commit
   (`gh run list --branch master --limit 1`), comment that and
   `jobshop close`; the merge step then has nothing to merge and closes.
2. Find or file the fix issue: `gh issue list --search "${sha}"` and
   `gh issue list --search "master CI"`. If none is open, file one at
   priority 0 with the failing job, the log excerpt and the suspected
   commit. Comment `Issue #<n>` on this step.
3. If a PR for it is already open, work that PR (`gh pr checkout <n>`);
   comments on this step or the run's `merge` step list refinery findings
   (`jobshop run show <run>` from this ticket's `parent_id`).
   Otherwise `git switch -c ci-fix-<short sha>`, fix the cause (not the
   test, unless the test is wrong), run
   `${check_cmd}`, commit with
   `(#<n>)` in the subject, `git push -u origin HEAD`, and
   `gh pr create --base master --title "<summary> (#<n>)" --body "Closes #<n>. Master red at ${sha}. ..."`.
   PR body: at most ~10 lines, plain text: what changed, why, how it was
   tested. No long summaries, file lists or pasted output.
4. Wait for CI with `gh pr checks --watch`; fix and push again. Do not
   end the turn to wait: if a long command is moved to the background,
   keep polling (`gh pr checks <n>`, `sleep 60`) until CI finishes.

Do not merge. When CI is green: `jobshop comment "PR <url> head <sha>"`
(sha from `git rev-parse HEAD`), then `jobshop close`. Releasing this
step counts as a failed attempt, so wait inside your turn. If you cannot
deliver, `jobshop close -r failed` with a comment saying why; after two
attempts the run escalates to a human.
