# Attention view and rule (#185)

2026-11-03. Part of phase 6 (#192): take as little operator attention
as possible, and make what does need a human easy to find.

Built:
- `jobshop attention` (RPC `attention`, admin or pool grant
  `attention`) lists only what waits on a human, across all projects
  and repos, oldest first, and prints `{"items": []}` when healthy.
  Kinds: `ticket` (open, labeled `needs-human`: run escalations, pool
  quarantine escalations), `question` (open, labeled `question`, until
  #186 defines worker questions), `run` (`needs_human` with no open
  escalation ticket), `pool` (quarantined with no escalation ticket,
  since the controller started) and `rollback`. Items carry
  `meta.github` repo and issue or PR (#183).
- Rollback: config `attention.rollback_file`, the self-deploy state
  `failed` file. Listed while it is newer than `deployed` beside it, so
  the next good deploy clears it. The NixOS module sets it when
  self-deploy is enabled.
- Gateway `attention_digest` (cron, UTC; module option
  `attentionDigest`) opens one thread in `escalation_room` with the
  list, and posts nothing when it is empty.
- Pack rule (packs/github planner prompt): every @mention of the human
  names the decision needed and what happens if nobody answers. The
  planner and operator pools get the `attention` grant.

Tests: `TestAttention` (pkg/control), `TestDigest` (pkg/gateway),
gateway config cases, pack tests. Live check: `docs/verify/attention.md`.
