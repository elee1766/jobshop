# JOB-172: packs are their own subsystem, config keeps the service

2026-11-09. Child of JOB-164. Behavior as built.

## What moved

| Before | After |
|---|---|
| `pkg/packcue` | `pkg/pack/cue` (one evaluator, no `config/pack/cue`) |
| `pkg/config/tmpl` | `pkg/pack/tmpl` |
| `pkg/cron` | `pkg/config/cron` (its callers are config validation and `control/triggers`) |
| pack assembly in `pkg/config` (`pack.go`, `rootpack.go`, `cuepack.go`, `textfile.go`, `scriptpath.go`, `syntax.go`, merge, prefixing, watch paths) | `pkg/pack` |
| CUE JSON Schema check in `pkg/orchestration/hooks/schema.go` | `pack/cue.CheckSchema`. `hooks.CheckStepOutput` keeps the size limit and calls it |

`pkg/pack` imports no config, store, automation or provider package, and
`pack/cue` imports nothing of the module. Both are enforced by archcheck rules
(`pack-is-independent`, `cue-only-in-pack-cue`). The `hooks -> cuelang.org/go`
exception is gone.

## The seam

`config.parseFile` reads the service file, applies includes, then calls
`pack.Assemble(Input, Options)` and decodes the returned YAML document itself.
`Options` carries what pack needs from config as data:

- `CheckLayer`: strict decode of one pack file against `packFile` or
  `rootPackFile`, so an unknown key still fails at that file.
- `LateVars`: the check output, event and scan vars kept as references.
- `EscalationVars` and `StepTemplateKeys`.

`Result` has the document, `Sources`, instances, warnings, `WatchPaths`, and
`Blame`, which marks an error that names a pack file as `ErrPack`. Hook program
validation, the last-good snapshot and `ScriptEntries` stay in config.

Pack types now live in pack: `Ref`, `Instance`, `WatchPaths`, `ErrPack`,
`FileName`, `CUEFileName`, `PrefixVar`. `rpc.PackInstance` is a defined type over
`pack.Instance` so the JSON and the generated `rpc.d.ts` are unchanged.

## Unchanged

Pack fixtures, imports and local replacements, source provenance, watch paths
and reload, instance labels and the last-good metadata format. The existing
config tests ran unchanged apart from the renamed identifiers. YAML packs and
the legacy hook surface are untouched: removing them is separate work.

## Check

`docs/verify/pack-boundary.md`.
