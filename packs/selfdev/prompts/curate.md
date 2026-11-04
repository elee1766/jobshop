Step `curate` of a `curate` run: a curator pass over learnings (#273).
You run in your own worktree of Jobshop's checkout, detached at
origin/master. Learnings are corrections, failures and workarounds that
sessions and the controller recorded (`jobshop learn`). Your job is to
turn the recurring ones into small edits that stop them happening again.

1. Settle earlier passes. `jobshop learn list --status open --limit 500`
   and look at learnings that already have a `pr`. For each such PR,
   `gh pr view <url> --json state,mergedAt,mergeCommit`:
   - merged and not reverted (`gh pr list --state all --search "Revert
     <title>"` finds no merged revert): `jobshop learn update <ids>
     --status applied --merged-at <mergedAt>`. That records the edit for
     the curator gate (#274), which later compares its pools before and
     after the merge;
   - closed unmerged, or merged and then reverted:
     `jobshop learn update <ids> --status rejected --pr <url>`;
   - still open: leave them; do not propose them again.
2. Read what was turned down: `jobshop learn list --status rejected
   --limit 200`. Do not propose an edit a rejected PR already made, or
   the same rule in other words.
3. Gather the rest: open learnings with no `pr`, plus recent outcomes
   (`jobshop run list`, failed or needs_human runs and their step
   comments). Group learnings that describe the same problem. Skip one-off
   noise (a single flaky network error); a group needs at least two
   learnings, or one correction from a human.
4. For each group pick one bounded edit (add, replace or delete a few
   lines) in one of:
   - memory files: `memory/memory/global.md`,
     `memory/memory/repo/<owner>__<name>.md`,
     `memory/memory/pool/<name>.md`; `jobshop memory check memory/memory`
     must pass (each file under the cap). Replace or delete stale lines
     rather than only appending;
   - skills: `memory/skills/<name>/SKILL.md` for a multi-step procedure;
   - pack prompts under `packs/` when the prompt itself told agents to do
     the wrong thing.
   Memory is public: no secrets, tokens, hostnames of private machines or
   personal data. Keep the plain style of AGENTS.md.
5. Ship all edits of this pass as one PR through `pr-work`, as improvement
   sessions do (#249). File an issue on elee1766/jobshop titled
   `Curate: <short summary>` whose body lists each edit and the learning
   ids it addresses (`learnings 12, 15, 31`). Then
   `git switch -c issue-<n>`, edit, run
   `go build ./... && go vet ./... && go test ./...` and
   `jobshop memory check memory/memory`, commit with `(#<n>)` in the
   subject, `git push -u origin HEAD`, `gh pr create --base master --body
   "Closes #<n>. ..."` with the same list of edits and learning ids.
   Start its run: `jobshop run start pr-work --var issue=<n> --key
   issue-<n>`, find its `implement` step with `jobshop run show <run>`,
   and `jobshop comment -t <step id> "PR <url> head <sha>"` (sha from
   `git rev-parse HEAD`). The dev session adopts the PR and the refinery
   merges it.
6. Mark the learnings the PR addresses: `jobshop learn update <ids> --pr
   <url>`. They stay open until a later pass sees the PR merged or
   closed (step 1).

Never push to master, never merge, never force-push a branch you did not
create. If there is nothing worth an edit, say so in a comment and close.
When done: `jobshop comment "<PR url or 'no edits'>, learnings <ids>"`
then `jobshop close`. If you cannot finish, `jobshop report failed
"<why>"` and `jobshop close -r failed`.
