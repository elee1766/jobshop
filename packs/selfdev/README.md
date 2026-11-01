# selfdev pack

Jobshop agents that develop Jobshop through GitHub issues and PRs.

- `dev` pool: resolves one issue per ticket (branch, PR, CI, squash merge).
  Prompt: `prompts/dev.md`.
- `plan` pool (1 session): reads the `target` issue, ticks verified items,
  files the next 2-4 issues and dev pointer tickets. Prompt:
  `prompts/plan.md`.
- `replan` trigger: when `dev` has no ready tickets, creates a planning
  ticket in `plan` (10m cooldown after the last one closes).

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | Main checkout of github.com/elee1766/jobshop; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `dev_sessions` | `max_sessions` for the dev pool |

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/selfdev
    vars: {repo: /var/lib/jobshop/src/jobshop, agent: jcode, dev_sessions: 2}
agents:
  jcode: {command: jcode, args: [acp]}
```

Pools use the host's `default_project`. Prompts are read when the
controller loads config, so a merged prompt change applies on restart.
