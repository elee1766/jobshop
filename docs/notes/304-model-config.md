# Model and effort per pool (#304)

2026-11-04.

- Config: `model`, `effort` and `config_options: {<id>: <value>}` on
  agents and pools. `model` and `effort` are shorthand for the ACP ids
  `model` and `reasoning_effort`. The pool overrides the agent per id.
  Values pass through as strings (any jcode route, e.g.
  `openai-oauth:<model>`, `claude-oauth@<account>:<model>`).
- After `session/new` or `session/load` the runner records the ids the
  agent advertised in `configOptions` (`fields.config_ids`) and its
  current `model` and `effort`. Before the first claim, and between every
  turn after that, `applyConfig` (`pkg/control/sessionconfig.go`) calls
  `session/set_config_option` for each wanted value that differs from
  what it last set (`fields.options`). Never during a turn.
- An id the agent did not advertise is logged once per session and
  skipped. A set that fails is retried once, then the session ends
  `config_failed` with the agent's message as the close comment, before
  any turn runs. `config_failed` counts toward the pool crash-loop pause
  like `crashed`.
- Reload: options are not in the host spec or its env hash, so a change
  never respawns. Each open session picks it up at its next between-turn
  point. Each set is logged with old and new values, recorded as a
  `session config` span, and published on `events.subscribe` as
  `kind: session, op: config` with `option`, `old`, `new`.
- `jobshop session config <id> model=<v> [effort=<v>] [--reset]`
  (admin, or the `session-config` pool grant; the operator pack has it)
  stores `fields.override`, applied the same way. `--reset` drops it and
  the pool's values are set again.
- Visibility: `session list` and Desk Sessions show `model`, `effort`
  and `override`. Lifecycle spans carry `jobshop.session.model` and
  `jobshop.session.effort`; prompts carry them in `_meta`
  (`jobshop/model`, `jobshop/effort`), so conversation export records
  have them too.
- Live check: `docs/verify/model-config.md`.
