# Memory files and skills (#266)

2026-11-04. A prompt rule can name a memory directory instead of (or as
well as) text or a command:

```yaml
prompt_rules:
  - name: memory
    at: system
    memory: {dir: /var/lib/jobshop/src/jobshop/memory, cap: 2200}
```

- Layout (`pkg/memory`): `memory/global.md` (every matching session),
  `memory/repo/<owner>__<name>.md` (turns whose ticket has
  `meta.github.repo`), `memory/pool/<name>.md` (the session's pool) and
  `skills/<name>/SKILL.md` (agentskills.io frontmatter `name`,
  `description`). Files go in that order, labeled `Memory (<scope>,
  <path>)`, then `Skills (read the file when one applies):` with one line
  per skill: name, description, absolute path. The agent reads a skill
  on demand.
- Read at prompt time, not at load, so a `git pull` of the directory is
  live without a reload. A relative dir resolves against the root dir.
  An empty dir or a missing one adds nothing.
- Cap: characters per file, default 2200. A longer file is cut at the cap
  and logged at WARN (`memory file over cap, truncated`).
  `jobshop memory check [dir] [--cap]` lists files over the cap and exits
  1; with no dir it checks every rule's `memory.dir` in jobshop.yaml. It
  reads files only, no controller.
- Packs: selfdev (rule `selfdev-memory`, `chat: false`) and operator
  (`operator-memory`, the operator pool) take vars `memory_dir` and
  `memory_cap` (2200).
- 2026-11-05: memory is local to the host, `<root>/memory` by default,
  and written by `jobshop memory write`; the repo's `memory/` is gone
  (#592, `docs/notes/592-local-memory.md`).
- Tests: `pkg/memory` (scopes, cap, skills index, missing dir),
  `TestRuleMemory`, `TestSelfdevPackLoads`.
- Live check: `docs/verify/memory-skills.md`.
