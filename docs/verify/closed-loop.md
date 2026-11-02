# Closed-loop evidence

Added 2026-11-01 (#76) for the #36 item "Planner sees a closed loop". One
cycle: a trigger files a plan ticket, the planner files dev tickets while
holding it, one of them goes through the refinery and its PR is merged.
Needs a binary with `ticket trace` (#53).

Pick the plan ticket: the newest closed `replan` ticket whose dev work has
had time to merge.

```sh
JOBSHOP_SOCKET= jobshop ticket list --pool plan --status closed \
  | jq -r '.[] | select(.meta.trigger.name == "replan") | "\(.closed_at) \(.id) \(.title)"' \
  | sort -r | head
```

Run the check and paste its full output on #36, plus the trace if it fails:

```sh
JOBSHOP_SOCKET= deploy/verify-closed-loop.sh <plan-ticket-id>
JOBSHOP_SOCKET= jobshop ticket trace <plan-ticket-id>   # on failure
```

Pass: every line `ok`, ending `closed-loop check passed`. On failure the
`FAIL` line names the missing link: no trigger, plan ticket not closed, no
dev ticket from the planner session, none reached `refinery`, or no
merged PR linked. A dev ticket that is still in `dev` or `refinery` means
the cycle has not finished yet. Pick an older plan ticket or wait for it.
`--trace FILE` checks saved trace JSON offline.
