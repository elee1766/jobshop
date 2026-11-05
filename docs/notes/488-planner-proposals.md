# Planner proposes the next target (#488, 2026-11-05)

Pools sat idle between phases because only humans filed targets. With pack
var `propose_targets` (default `true`) the `packs/github` planner prompt
gains a "Proposing a target" section:

- Propose when no target is open, or every open target item is covered or
  waits on time or a human and fewer than 3 issues without `needs-human`
  have a running pr-work run.
- At most one open issue labeled `proposed`, titled `Proposed target: ...`,
  in the target body format with 4 to 8 items, each citing its source.
  The planner creates the label if it is missing.
- A human approves by relabeling `proposed` to `target`. The planner never
  sets `target`, never files work for a proposal and never edits it after a
  human commented. A closed proposal is rejected.
- The planner files a Jobshop ticket labeled `proposed` with
  `meta.github` for the issue. `jobshop attention` lists open `proposed`
  tickets as kind `decision`.

`propose_targets: false` renders the prompt as before: no target, close.
