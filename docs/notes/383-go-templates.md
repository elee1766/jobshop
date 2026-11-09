# Go text/template for pack templates (#383)

2026-11-04. Behavior as built.

- A pack opts in with `template: go` at the top of `pack.yaml`. The
  in-repo packs (github, selfdev, operator) are converted; packs without
  the key still use `${name}` / `$$` and load unchanged. Each pack's own
  syntax applies to its own files, so a `${}` pack can extend or sit next
  to a `go` pack.
- `pkg/pack/tmpl` (was `pkg/pack/tmpl` (was `pkg/config/tmpl` until JOB-172) until JOB-172) parses with `missingkey=error` and a small FuncMap:
  `default`, `join`, `quote`, `lower`, `indent`, `toYAML`. No shell or
  file access.
- Same two passes as before. Load time renders pack vars into every
  string (an error names the field, e.g. `pools.dev.prompt: var "x" is
  not set`). Step titles and prompts wait for run start; trigger run vars
  and keys wait for the trigger to fire (`{{.check_output}}`, stall
  vars). The loader stores the syntax on each workflow and trigger run
  (`template`), and `Workflow.Render` / `TriggerRun.Fill` use it.
- A string that is exactly `{{.name}}` takes the var's value with its
  type, so `max_sessions: "{{.n}}"` stays a number. Quote it in YAML:
  an unquoted `{{` starts a flow mapping.
- Literal `{{` is written `{{"{{"}}` (GitHub Actions `${{"{{"}} ... }}`).
- `${name}` left in a `go` pack is a load warning.
- Load-time checks of step and trigger references walk the parse tree
  (fields under `range` / `with` bodies are not top-level vars).
- `TestPacksGolden` (`pkg/config/testdata/packs.golden`) renders all
  in-repo packs, steps and trigger runs. It was recorded from the `${}`
  packs and passes unchanged on the converted ones. Rerun with
  `go test ./pkg/config -run TestPacksGolden -args -update` when a prompt
  change is intended.

Next: once hosts' own packs have moved, drop `${}`, `varPattern` and the
`template` key (a follow-up release).

2026-11-05: done in #566, see `docs/notes/566-drop-dollar-templates.md`.
