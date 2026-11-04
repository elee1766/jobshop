# Prompt rule injection dedup (#397)

2026-11-04. A session that keeps its conversation already holds the rule
text it was sent, so the controller no longer resends unchanged rules
every turn.

- Each runner keeps, per rule name, a hash of the text (rule text,
  command output, memory files) it last sent. A turn sends a rule only
  when it is new to the session or its text changed. A changed rule is
  sent as `Updated prompt rule "<name>":` plus the full new text.
- A rule the session received that is gone from config is announced once:
  `Prompt rule "<name>" no longer applies.`, just before the ticket or
  messages.
- Per rule `inject:` `changed` (default), `once` (send once per
  conversation, never updates) or `always` (every matching turn, the old
  behavior).
- The record is memory only. It resets on every host spawn (new
  conversation, journal-summary fallback #43, load after a respawn with
  the recovery note #189) and on a new ACP session; a controller restart
  also starts empty. A reset can only resend more, never less.
- Scope: prompt rules and memory rules (#266). The pool prompt and
  ticket text are unchanged.
- `jobshop session prompt <id>` labels rule parts with their status:
  `(new)`, `(changed)`, `(always)`, `(removed)`, or
  `(unchanged, not sent)`; `--json` has `status` on each part. `rules`
  lists only those that would be sent.
- Metric `jobshop.prompt.injected_bytes` by pool and `reason` (new,
  changed, always, removed, unchanged for bytes skipped).
- The `_meta` `jobshop/rules` list names only the rules sent that turn.
- Tests: `TestDedupeTurns` (three turns: unchanged, edited, removed;
  reset; once and always), `TestPromptRules`.
