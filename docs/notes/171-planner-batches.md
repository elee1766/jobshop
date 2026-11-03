# Planner starts dependent issues as a batch (#171)

2026-11-03. Follow-up 4 of `docs/notes/148-batches.md`.

`packs/github/prompts/plan.md` step 8: issues that depend on each other
start with `jobshop batch start --key <slug> --issues ... --order ...`
instead of `run start` with `--after`. Single independent issues stay on
`run start pr-work`. Step 11: progress comments on the target quote
`jobshop batch show <key>` instead of listing runs. Prompt change only;
`TestSelfdevPackLoads` checks the rendered plan prompt.

Evidence still to collect on #145: one planner pass that started a batch
and quoted `batch show`.
