# Escalation drill (#114)

Added 2026-11-02. `packs/github` has a `drill` workflow (var `issue`,
`escalate_pool: plan`) with one step `fail` in `dev`, `retries: 1`,
whose prompt only comments and closes `failed`. One run therefore ends
in `needs_human` with an "Escalate run <id>" ticket in `plan`, proving
the `90-step-retries-escalation.md` path live without waiting for a real failure. Runbook:
`docs/verify/escalation-drill.md`. Test: `TestSelfdevDrill`.
