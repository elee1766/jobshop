# Dev pool routing: one week of numbers (JOB-177)

2026-11-13. Follow-up to `JOB-59-dev-pool-routing.md` and
`pr-work-route-step.md`. Window: `pr-work`, `pr-work-linear` and
`acphost-pr-work` runs created 2026-11-06 06:17Z (first run with
`dev_pool`) to 2026-11-13 06:18Z: 212 runs. Until 2026-11-10 the planner
chose the pool (`dev_pool` var, or none for the default `dev`); after
that the `route` step did (42 runs, 13 `scoped`, 29 `primary`).

| start pool | runs | cancelled | closed no PR | merged | moved to retry pool | failed attempts | review changes | tokens / merged PR | fresh in / merged | out / merged | median tokens / PR | median claim to close |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| dev (Opus) | 143 | 1 | 6 | 134 | n/a | 1 | 5 | 14.0M | 390k | 178k | 9.3M | 26 min |
| dev-sonnet | 28 | 0 | 3 | 25 | 1 (3.6%) | 0 | 0 | 3.9M | 166k | 48k | 1.7M | 10 min |
| acphost-dev | 23 | 1 | 8 | 14 | n/a | 0 | 0 | 5.4M | 250k | 93k | 3.2M | 10 min |
| acphost-dev-sonnet | 18 | 2 | 5 | 11 | 0 | 0 | 0 | 3.8M | 194k | 24k | 2.6M | 6 min |

- Runs to a Sonnet pool: 28 of 171 jobshop runs (16%), 46 of 212 with
  acphost (22%). Via the route step: 13 of 42 (31%).
- The one dev-sonnet run moved to `dev` was a review resume: the PR
  duplicated a fix already on master (PR 922 vs 924) and the retry closed
  it as a duplicate. No Sonnet attempt failed and no Sonnet PR got
  changes requested. `dev` steps were resumed (stall check or admin)
  26 times against 1 for dev-sonnet; a resume does not move a `dev` step.
- Refinery tokens per merged PR are about equal (dev 0.93M, dev-sonnet
  0.76M, acphost 0.70M / 0.77M), so Sonnet PRs did not push cost into
  review.
- Queue wait (step created to first claim) median 0 in every pool; p90
  2 min on dev-sonnet with its 1 session, 25 min on acphost-dev.

Method. Steps: `jobshop ticket list --parent <run>` and `jobshop ticket
show <implement step>` (pool, `retry_pool`, `failures`, reopen history,
"Moved from ... to retry pool" comments, request-changes reopens).
Merged: implement outcome `pr` and review-merge closed `done`. Tokens:
sum of `usage.totalTokens` from each ACP `session/prompt` result in the
session journals, attributed to the ticket named in that prompt
("claimed Jobshop ticket ..."). Total is dominated by cache reads;
"fresh in" is input plus cache write tokens. 192 message turns (37M of
3.5G tokens, 1%) name no ticket and are left out. Models confirmed from
the journals: `dev` claude-opus-5-5, `dev-sonnet` claude-sonnet-5-5 high.

Caveat: the pools did not get the same work. Sonnet got scoped issues
by design, so the token gap (3.6x per merged PR) is partly the work, not
the model. Post-merge defects per pool were not measured.

Recommendation: keep routing and keep `retry_pool` `dev`. The retry
rate to `dev` is near zero, so the rubric is if anything too cautious.
Next steps, each a separate change: let the route prompt send more
clear single-package fixes, test and docs work to `scoped`, and raise
`dev_sonnet_sessions` to 2 if its p90 queue wait passes ~15 min. Measure
again after a week of route-step-only runs, adding post-merge reverts
or follow-up fixes per pool.
