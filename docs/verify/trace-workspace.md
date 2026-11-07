# Trace workspace and transcript provenance

2026-11-07

Use an isolated controller with an operator chat pool, escalation `chat_pool`,
and attention prestarts enabled. Do not run mutation specs against Jobcity.
Build the frontend before building the Go binary that embeds it, then start
both that controller and its Desk server.

```sh
npm --prefix ui run check
npm --prefix ui run build
go build -o "$JCODE_SCRATCH_DIR/desk-workspace-jobshop" ./cmd/jobshop
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  DESK_ROOT="$JCODE_SCRATCH_DIR/desk-workspace-root" \
  JOBSHOP_BIN="$JCODE_SCRATCH_DIR/desk-workspace-jobshop" \
  npm --prefix ui run test:e2e -- trace-session.spec.js live.spec.js desk.spec.js chat-dismiss.spec.js --workers=1
```

The trace-session spec uses the admin CLI to create a real claim. Desk's RPC
allowlist stays unchanged. It checks conversion without resolution, exact human
text (including literal wrapper-like text), collapsed injected context, live and
replayed read-only transcripts, desktop split geometry and mobile dismissal.
It saves populated session/history screenshots in Playwright's output directory.

Check a configured workflow run with GitHub `repo`/`pr` variables and either
explicit Linear metadata or a configured Linear workspace identity. Its Home
PR and issue links must open the canonical external pages, while the ticket
title opens Trace. An ambiguous/missing Linear workspace must not produce a
fabricated URL.

On a ticket with multiple dependencies, inspect each direction separately:
parent above the title, children together, `blocked by`, `blocks`, and waits in
their own sections. O/C/R statuses have distinct colors and retain text labels.
History should show action, step, time, reason and navigable references without
opening raw metadata. Raw diffs and complete model prompts remain expandable.
Older journal prompts without presentation metadata remain collapsed legacy
prompts, not heuristically rewritten messages.
