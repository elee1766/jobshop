# Escalation drill

Added 2026-11-02 (#114) for the #82 boxes "Retry limits" and
"Escalation". File a throwaway issue, then start the drill against it:

```sh
JOBSHOP_SOCKET= jobshop run start drill --var issue=<n> --key drill-$(date +%F)
```

A dev session claims the step, comments and closes it `failed`. Check:

```sh
JOBSHOP_SOCKET= jobshop run show <run>          # state needs_human, run still open
JOBSHOP_SOCKET= jobshop ticket list --label needs-human \
  | jq -r '.[] | select(.title == "Escalate run <run>") | "\(.id) \(.pool)"'
gh issue view <n> --comments                    # step-history comment
```

Pass: the run is `needs_human`, exactly one "Escalate run <run>" ticket
is in `plan`, and the issue has the step-history comment. Post the run
id and links on #82, then close the escalation ticket and the throwaway
issue. The same key returns the existing run, so use a new date to
repeat.
