# Rebase route-back without a failure (#424)

2026-11-04.

`jobshop close -r rebase` on a step ticket reopens it the way `-r failed`
does (open, unheld, still blocked behind anything it depends on) but does
not bump `fields.failures`, so it never reaches the retry limit or puts
the run into needs_human. The store adds the note "Closed for a rebase"
to the step. On a non-step ticket `rebase` is an ordinary closed
resolution.

The github pack's review-merge prompts (github and linear) use it when a
PR has non-docs conflicts with master and no review finding: comment the
conflicted files, `run resume --step implement`, `close -r rebase`. A
conflict together with a real finding, and every other case 4, still
closes `-r failed` and counts.

Tests: `TestConformance/StepFailures` (store), `TestStepRebaseNoFailure`
(control).
