Step `route` of a `pr-work` run for Linear issue {{.issue}}.

Pick who implements it. Do not write code, open a PR or change the issue.

Read the issue: `jobshop linear issue {{.issue}}` (description and
comments) and `jobshop linear context {{.issue}}` (labels).

- A `model:sonnet` or `model:opus` label on the issue wins: `scoped` or
  `primary`.
- `scoped`: clear spec and acceptance checks, one area, known patterns,
  tests, docs or config, mechanical renames, a bug fix with a known
  cause.
- `primary`: needs investigation or design, cross-cutting (store schema
  plus controller plus RPC), concurrency, leases or isolation, unclear
  root cause, security.
- Unsure: `primary`.

Comment the choice and a one-line reason on the issue
(`jobshop linear comment {{.issue}} "route=<word>: <reason>"`), then
`jobshop close --outcome <word>`. The run starts `implement` in the pool
for that word. A failed attempt or a review resume moves the step to the
primary pool on its own.
