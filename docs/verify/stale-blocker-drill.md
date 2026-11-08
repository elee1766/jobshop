# Stale-blocker drill

Superseded 2026-11-02 by `scan-drill.md`: #143 removed `blocker-check`.
Kept as the record of #119.

Added 2026-11-02 (#119) for the #82 box "Blocker assessment", stale
path: a check finds the blocker gone and routes the work back. Make a
ticket that waits on a deploy that is already live, park it in a pool
with no sessions, and file a throwaway `needs-human` issue about it:

```sh
sha=$(jobshop version | tr -d '()' | cut -d' ' -f3)   # deployed commit
id=$(JOBSHOP_SOCKET= jobshop ticket create "drill: waiting on deploy of $sha" \
  -b "Blocked: waiting on deploy of $sha." --pool dev | jq -r .id)
JOBSHOP_SOCKET= jobshop ticket update "$id" --pool drill-none
n=$(gh issue create --title "drill: stale blocker $id" --label needs-human \
  --body "Ticket $id says it waits on deploy of $sha." | grep -o '[0-9]*$')
JOBSHOP_SOCKET= jobshop run start blocker-check --var ticket="$id" \
  --var issue="$n" --key blocker-drill-$(date +%F)
```

Check:

```sh
JOBSHOP_SOCKET= jobshop run show <run>     # state done, assess closed done
JOBSHOP_SOCKET= jobshop ticket show "$id"  # pool dev again, history shows the update
gh issue view "$n" --json labels,comments  # no needs-human, evidence comment
JOBSHOP_SOCKET= jobshop ticket list --label needs-human \
  | jq '[.[] | select(.title == "Escalate run <run>")] | length'   # 0
```

Pass: the run is `done`, the ticket is back in a real pool (`dev`, the
pool it held before parking), the label is gone and no escalation ticket
exists. Post the run id and outcome on #82. Clean up:
`JOBSHOP_SOCKET= jobshop close "$id"` and
`gh issue close "$n"`. Use a new date in the key to repeat.
