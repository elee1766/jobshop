# selfdev ci-repair (#99)

`master-red` is a `run:` trigger: `{workflow: ci-repair, vars: {sha:
${check_output}}, key: ${check_output}}`. `ci-repair` (vars `[sha]`,
`escalate_pool: plan`) has `fix` (dev, 2 retries, `prompts/ci-fix.md`:
find or file the fix issue, `gh run view --log-failed`, open a PR) and
`merge` (refinery, 2 retries, `prompts/ci-merge.md`). It does not reuse
`merge.md`, which needs an `issue` var and refuses to merge while master
is red. With the SHA as both run key and trigger key, a SHA gets one run
and at most two fix attempts, then `needs_human` and an escalation ticket
(`90-step-retries-escalation.md`).
