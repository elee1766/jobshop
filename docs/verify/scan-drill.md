# Scan drill

Added 2026-11-02 (#143), renamed 2026-11-08 with the scan rename: the
`triage` workflow resolves an open issue with no run and a stale
`needs_human` run without the operator. Run on gasserver after the deploy
is live (`jobshop version` at or after the merge commit).

Set up both findings:

```sh
n=$(gh issue create --title "scan check: no run $(date +%F)" \
  --body "Throwaway issue for docs/verify/scan-drill.md. Close when the drill is done." \
  | grep -o '[0-9]*$')
# A drill run fails at once and sits in needs_human. It names the target
# (#82), not $n: a run naming $n would count as $n's run.
d=$(JOBSHOP_SOCKET= jobshop run start drill --var issue=82 \
  --key scan-drill-$(date +%F) | jq -r .run.id)
```

The issue title must not start with "drill" (the `issue_no_run` scan
script skips drills). Then wait: `issue_no_run` fires once the issue is
30m old, `needs_human` once the drill run has been idle
`scan_human_after` (4h).

Check:

```sh
JOBSHOP_SOCKET= jobshop trigger list | jq '.[] | select(.name == "scan")'
JOBSHOP_SOCKET= jobshop run list | jq '.[] | select(.fields.key | startswith("scan:"))
  | {id, title, key: .fields.key, status}'
JOBSHOP_SOCKET= jobshop run list | jq -r --arg n "$n" '.[] | select(.fields.key == "issue-"+$n) | .id'
JOBSHOP_SOCKET= jobshop run show "$d"     # cancelled (stale drill) or resumed
```

Pass: two triage runs, keys `scan:issue_no_run:$n` and
`scan:needs_human:$d`, both closed `done`; a `pr-work` run with key
`issue-$n` exists; the drill run is no longer `needs_human`; no new
escalation ticket for either triage run. Post the run ids on #82.

Clean up: `JOBSHOP_SOCKET= jobshop run cancel <pr-work run>`,
`gh issue close "$n"`.
