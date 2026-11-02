# Stall drill

Added 2026-11-02 (#143) for the #82 stall box: stall-check resolves an
open issue with no run and a stale `needs_human` run without the
operator. Run on gasserver after the #143 deploy is live (`jobshop
version` at or after the merge commit). Replaces the stale-blocker drill.

Set up both stalls:

```sh
n=$(gh issue create --title "stall drill: no run $(date +%F)" \
  --body "Throwaway issue for docs/verify/stall-drill.md. Close when the drill is done." \
  | grep -o '[0-9]*$')
# A drill run fails at once and sits in needs_human. It names the target
# (#82), not $n: a run naming $n would count as $n's run.
d=$(JOBSHOP_SOCKET= jobshop run start drill --var issue=82 \
  --key stall-drill-$(date +%F) | jq -r .run.id)
```

The issue title must not start with "drill" (the `issue_no_run` check
skips drills). Then wait: `issue_no_run` fires once the issue is 30m old,
`needs_human` once the drill run has been idle `stall_human_after` (4h).

Check:

```sh
JOBSHOP_SOCKET= jobshop trigger list | jq '.[] | select(.name == "stall")'
JOBSHOP_SOCKET= jobshop run list | jq '.[] | select(.title | startswith("Run stall-check"))
  | {id, title, status}'
JOBSHOP_SOCKET= jobshop run list | jq -r --arg n "$n" '.[] | select(.title == "Run pr-work issue-"+$n) | .id'
JOBSHOP_SOCKET= jobshop run show "$d"     # cancelled (stale drill) or resumed
```

Pass: two stall-check runs, keys `stall:issue_no_run:$n` and
`stall:needs_human:$d`, both closed `done`; a `pr-work` run with key
`issue-$n` exists; the drill run is no longer `needs_human`; no new
escalation ticket for either stall-check run. Post the run ids on #82.

Clean up: `JOBSHOP_SOCKET= jobshop run cancel <pr-work run>`,
`gh issue close "$n"`.
