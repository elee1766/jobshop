# Prompt rules (#205)

2026-11-03.

Operator guidance added to session prompts from config, without editing
pack prompts.

```yaml
prompt_rules:
  - name: no-master            # required, unique across all files
    text: Never push to master.
    at: system                 # system | before | after | turn (default turn)
  - name: dev-pr
    file: rules/dev.md         # read at load, relative to this file
    match: {pools: [dev], repos: ["elee1766/*"], steps: [implement]}
  - name: policy
    command: ./policy-hook     # sh -c in the root dir; JSON event on stdin
    timeout: 5s                # default 10s
```

- Sources: packs and the host file can both declare `prompt_rules`. The
  lists are concatenated in pack order, host last. Names must be unique,
  so a pack cannot replace or remove a host rule (a duplicate name is a
  load error). `config show` lists the rules, and its sources include
  `prompt_rules.<name>`.
- Match: each set field must match, and a list matches when any glob
  (`path.Match`) does. An empty match is every session. Fields: `pools`,
  `agents`, `packs` (dir names of the packs that set the session's
  pool), `projects`, `repos` (`meta.github.repo`, set by #183),
  `workflows`, `steps`, `labels`, `chat`. Repos, workflows, steps and
  labels describe the turn's ticket, so they never match message turns
  or a ticket without that field.
- Order, top to bottom: system rules, before rules, pool prompt, after
  rules, turn rules, then the ticket or messages. Within a position,
  rules keep config order.
- `system` goes on the first turn of a host and again on the first turn
  after a respawn (restart, crash, deploy). `before` and `after` go
  wherever the pool prompt goes: every ticket turn, and a chat session's
  first message turn. `turn` goes on every ticket and message turn.
- `command`: its stdout follows `text`, capped at 16 KiB. A non-zero
  exit or a timeout adds nothing from the hook and logs WARN
  `prompt rule command failed` or `timed out`. The event on stdin is
  `{rule, at, session, pool, agent, project, turn, ticket, run, step,
  repo}`, where `turn` is `ticket` or `message`.
- The prompt's ACP `_meta` carries `jobshop/rules` (the rule names), and
  the host journals it with the prompt, so the journal records which rules
  applied to each turn. The controller also logs `prompt rules applied`.
- `jobshop session prompt <id>` (admin) prints the next turn's prompt,
  with each part under `=== <source> ===` (`pool`, `rule:<name>`,
  `ticket`, `messages`). The turn uses the ticket the session holds,
  else its pool's next ready ticket, else a message turn. Nothing is
  claimed, and a pending system turn stays pending. Rule commands run.
  `--json` prints the parts.
- Reload: config is read each turn, so added, changed or removed rules
  apply from the next turn. Rule files are read again on reload.
- Tests: `TestRuleMatch`, `TestPromptRulesMerge`, `TestPromptRuleErrors`
  (config). `TestAssembleOrder`, `TestRuleCommand` and `TestPromptRules`
  (control) cover respawn, journal, `session prompt`, reload and
  `config show`.
