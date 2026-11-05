# Drop ${name} pack templates (#566)

2026-11-05. Finishes #383.

- Go `text/template` is the only pack syntax. `SyntaxDollar`,
  `varPattern`, the `$$` escape, `FillTemplate` and the `${steps.<id>...}`
  step refs are gone; command and http steps use `{{.steps.<id>.output}}`.
- `template: go` still loads, as a no-op for one release. Any other value
  is a load error naming the file and #566. The in-repo packs dropped it.
- The loader no longer stores a `template` field on workflows, escalation
  blocks or trigger runs. A last-good config saved by an older binary
  with `template: go` on those would no longer parse strictly; it is only
  read when a pack fails, and the next good load rewrites it.
- `${name}` anywhere in a pack file or a pack's text file is a load
  error (it was a warning in Go packs), so `config check`, reload and
  self-deploy reject a stale pack instead of rendering literal text.
  The check runs on the pack as written, before vars fill in, so a var
  value that holds shell `${...}` is fine.
- `${SESSION_DIR}` is allowlisted in that check. It is not pack syntax:
  the controller expands it in agent args at session start
  (`pkg/control/session.go`). An allowlist, rather than skipping agent
  args, keeps the check on every other field of an agent.
- Host `jobshop.yaml` is not a pack and never had substitution, so
  `${SESSION_DIR}` there needs no escape.
