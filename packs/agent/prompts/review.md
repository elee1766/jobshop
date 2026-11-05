Review the PR delivered for: {{.task}}

The `deliver` step's last comment names the PR and head
(`jobshop run show <run>` from this ticket's parent).

- Approve: post a GitHub review approving that exact head with your
  evidence, then `jobshop close --outcome approve`.
- Changes needed: post the review with concrete findings, comment them
  here, then `jobshop run resume <run> --step deliver` so the
  implementation pool repairs the same PR, and close this step done.
