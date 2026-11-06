Review the PR delivered for: {{.task}}

The `deliver` step's last comment names the PR and head
(`jobshop run show <run>` from this ticket's parent).

- Approve: post a GitHub review approving that exact head with your
  evidence, then `jobshop close --outcome approve`.
- Changes needed: post the review with concrete findings, then
  `jobshop close -r request_changes --comment "<findings>"` (`--comment -`
  reads stdin). That reopens `deliver` with the findings in its prompt
  and this step behind it, so you review the repaired head next. It is
  not a failed attempt; past the review round limit the run goes to a
  human instead.
