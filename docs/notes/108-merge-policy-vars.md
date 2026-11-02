# Merge policy vars (#108)

`packs/github` defaults `merge_method: squash`, `required_approver: ""`
and `required_checks: ""`; selfdev keeps the defaults. `review-merge`
and ci-repair `merge` run `gh pr merge --${merge_method}
--match-head-commit <head>`. A set approver must have an `APPROVED`
review whose `commit_id` is the current head (`gh api
repos/<repo>/pulls/<n>/reviews`, filtered by login, state and commit);
without it the step comments and closes `failed`, so it reopens and
checks again until the retry limit escalates. ci-repair gets no waiver:
a fix for red master needs the same approval. Empty `required_checks`
means every check must pass, otherwise the listed names must be
`SUCCESS`. The rules are prompt text, not controller code, like the rest
of the refinery's judgment.
