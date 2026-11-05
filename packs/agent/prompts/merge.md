Deliver the PR for: {{.task}}

1. Read the run (`jobshop run show <run>`): the deliver step's PR and
   head, the review step's verdict. Fetch fresh state: `gh pr view <n>
   --repo {{.repo}} --json headRefOid,state,reviewDecision,mergeable` and
   `gh pr checks <n> --repo {{.repo}}`.
2. Head moved since review, or evidence missing: `jobshop run resume
   <run> --step review` (new head) or `--step deliver` (repairs), comment
   the findings, close this step done.
3. Only CI or review pending: wait with `jobshop wait add
   'github.checks:pr={{.repo}}#<n>,state=success' ... --any --resume`
   and end the turn.
4. Every gate passed: re-fetch, then `gh pr merge <n> --repo {{.repo}}
   --squash --match-head-commit <full sha>`. Confirm GitHub reports it
   merged, comment the merge commit, head and evidence, `jobshop close`.
5. A PR closed unmerged is not shipped: `jobshop report waiting` for the
   decision, do not close done.
