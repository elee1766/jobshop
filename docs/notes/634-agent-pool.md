# Agent pool pack (#634)

2026-11-05. `packs/agent` ports the gascity factory's agent pool from the
city config snapshot (gc HEAD 153951e): the implementation pool
(`routing.primary`, which replaced `gastown.polecat`), the independent
review and PR refinery pools it hands to, and the main-CI repair order.
Load it once per repo with a `prefix:` (gascity loaded it once per rig).
gasserver is not switched over: `docs/verify/agent-pool.md`.

Vars: `agent`, `checkout`, `repo`, `branch` (main), session caps,
`check_cmd`, `ci_quiet`. Sandbox and isolation are not vars (JOB-83): the
host attaches profiles to the pools by name
([JOB-83-profiles.md](JOB-83-profiles.md)).

| gascity | jobshop |
| --- | --- |
| rig (`[[rigs]]`, `default_branch`) | pack instance with `prefix:`, vars `repo`, `branch` |
| agent pool (`routing.primary`, `routing.review`, `gastown.refinery`) | pools `primary`, `review`, `refinery` |
| provider (`jcode`, `jcode-review`, `--model`) | `agents:` entry in jobshop.yaml, var `agent`; per-pool `model:` |
| `max_active_sessions`, `sleep_after_idle` | `max_sessions`, `idle_timeout` |
| `work_dir` worktree per agent | pool `repo` + `ref`: a worktree per session |
| `prompt_template` (`prompts/worker.md`, `refinery.md`, `sol-reviewer.md`) | pool `prompt_file` |
| formula `mol-gfx-pr-work` (one step, handed worker to refinery) | workflow `pr-work`: `deliver`, `review`, `merge` steps; repair is `run resume --step deliver` |
| formula `mol-main-ci-repair` | workflow `ci-repair` |
| order `main-ci-retail` (condition, `main-ci-due`) | trigger `main-ci` (`check`, key per failing sha) |
| bead, source + step pair, `gc hook --claim` | ticket and run step, claimed by the supervisor |
| `bd update --append-notes`, `gx task finish` | `jobshop comment`, `jobshop close` |
| `hold:mayor`, `gc mail send gastown.mayor` | `jobshop report waiting`, escalation to `escalate_pool` |
| watched gate (`pr-refinery` gate) | `jobshop wait add 'github.checks:...' --resume` |
| `gx learn log` | `jobshop learn` (prompt rule `agent-learn`) |
| order `turn-guard` (400 calls / 120 min) | pool `turn_timeout: 2h` (no tool-call cap) |

No equivalent, not built here:
- Model routing by PR (`gx pr review-route` to Sonnet or Sol): one
  review pool; a second pool needs a routing step.
- `gx` helpers (`task context`, `pr state`, `pr request-review`,
  `pr carry-approval`, `emulator request`): prompts use `gh` and
  `jobshop` directly. Device checker pools (android-emulator,
  ios-simulator) and their resource orders are not ported.
- Per-rig merge policy tables (`ops/policy/review.md`: required
  approvers, human sign-off categories) and the `ops/policy/*` files the
  worker loads: the refinery prompt states the general gate only.
- Linear label intake, Design-only holds, funded test accounts and the
  money/wallet rules: Retail specific, left to a per-repo prompt rule.
- Orders without a pool: `blocked-revalidator`, `checker-unwedge`,
  `intake-queue`, `prune-testcontainers`, `retail-resources-*`,
  `improvement-sweep`, `review-lessons` (selfdev's curate covers part).
- The worker name pool (`namepool`).

Test: `TestAgentPoolPack` loads the pack, then runs `pr-work` end to end with the fake agent.
