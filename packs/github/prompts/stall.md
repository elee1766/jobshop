Step `diagnose` of a `stall-check` run: signal `${signal}` on `${subject}`.

The controller saw work stop moving: ${detail}

Find out why from current state, not from what tickets claim, and get
the work moving again. Nobody else is watching; if you do nothing the
pool stays idle.

What the signal means and where to start:
- `ready_unclaimed`: ticket `${subject}` is ready in its pool but no
  session took it. `jobshop ticket trace ${subject}`, `jobshop session
  list`, `jobshop ticket list --orphaned`. Is the pool drained, full of
  stuck sessions, or failing to spawn?
- `pool_quarantined`: the pool before `@` crash-looped and is paused.
  Read the last session dirs' exit reason and journal (`jobshop session
  list`, `jobshop ticket trace` of what they held).
- `step_stalled`: step `${subject}` is held but nothing was recorded on
  it for a long time. `jobshop ticket show ${subject}`, its run (`jobshop
  run show <parent_id>`), the holder's journal.
- `session_silent`: session `${subject}` holds a ticket but its agent
  wrote nothing. `jobshop session list`, its journal; restart it with
  `jobshop session restart ${subject}` if the agent is hung.
- `needs_human`: run or ticket `${subject}` has waited on a person for
  a long time. Read the escalation (`jobshop ticket list --label
  needs-human`), `jobshop run show`, the issue. The blocker may be gone.
- `blocked`: ticket `${subject}` is labeled `blocked`. Re-check the
  claim (below).
- `issue_no_run`: issue #${subject} is open with no run. `gh issue view
  ${subject} --comments`.
- `pr_unmerged`: PR #${subject} has green checks but is not merged.
  `gh pr view ${subject}`, its run's `review-merge` step.
- `no_merge`: nothing merged to master for a long time while work is
  open. Look at what the open runs are waiting on.

Checking a blocker claim: deploys (${deploy_check}), pools (`jobshop
ticket list --orphaned`), permissions and credentials (the read-only
form of what failed, e.g. `gh auth status`, `gh api repos/${repo}`),
and any merged PR, closed issue or run it names.

Fix in place when you can, one or more of:
- Start a missing run: `jobshop run start pr-work --var issue=<n> --key
  issue-<n>` (skip issues labeled `target`, `needs-decision`, drills and
  tracking issues; comment why instead).
- Requeue or reroute a ticket: `jobshop ticket update <id> --pool
  <pool>`, `jobshop ticket update <id> --defer 0`, `jobshop label rm <id>
  blocked`.
- Resume a run whose blocker is gone: `jobshop run resume <run>`, and
  drop `needs-human` from its issue (`gh issue edit <n> --remove-label
  needs-human`).
- A time-gated step that is waiting on purpose (a 24h window, a
  scheduled run): leave it and say until when; `jobshop release --until`
  belongs to the session holding it.
- Rebase a conflicted PR: `gh pr checkout <n>`, `git rebase
  origin/master`, push with `--force-with-lease`.
- Cancel a stale drill run: `jobshop run cancel <run>`.
- Restart a stuck session: `jobshop session restart <id>`.
- A code bug in this repo: file it (`gh issue list --search` first, then
  `gh issue create`) and start its run.

Do not merge PRs and do not write code here.

Escalate instead, with `jobshop close -r failed` (this step has one
attempt, so the run goes to `needs_human` at once and the operator is
asked), when:
- the fix needs a person (a decision, a credential, access), or
- it is a host config problem: a bad repo path, a missing secret or
  token, an agent command that does not start. Always escalate these; do
  not work around them.
Before closing failed, comment the evidence: what you checked (commands
and short output), what is wrong, what would unblock it.

Record what you found and did with `jobshop comment` on this step, and
on the issue if there is one (`gh issue comment <n>`). Then `jobshop
close` when the work is moving again. Earlier stall-check runs for this
stall are listed by `jobshop run list` with key `stall:${signal}:${subject}`;
if the same fix did not hold last time, escalate.

Do not release this step (a release counts as the failed attempt).
