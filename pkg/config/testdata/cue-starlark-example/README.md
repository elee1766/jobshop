# Offline CUE and Starlark example

This module is a runnable loader/runtime fixture for JOB-112. See the
[configuration guide](../../../../docs/configuration/README.md) and
[callback reference](../../../../docs/configuration/callbacks.md).

From the repository root:

```sh
go run ./cmd/jobshop config check --file pkg/config/testdata/cue-starlark-example/jobshop.yaml
go run ./cmd/jobshop config resolve --file pkg/config/testdata/cue-starlark-example/jobshop.yaml
go test ./pkg/config -run '^TestCUEPack' -count=1
go test ./pkg/execution/automation -run '^TestDocumented(Hook|Capability)Examples$' -count=1 -v
```

No controller, external registry or live credentials are needed. Tests
use the actual files, local HTTP/Jev test endpoints and `/bin/printf` when
available. The shell subtest reports a skip if that executable is absent.
A passing local capability example does not grant a running service access.

| File | Role |
| --- | --- |
| `jobshop.yaml` | Service entry, no HTTP/shell/Jev grants |
| `cue.mod/module.cue` | Checked-in local module and language version |
| `pack.cue` | Root imports, required var binding and enabled override |
| `packs/demo/pack.cue` | Reusable constrained/defaulted pack, static and callback workflows |
| `packs/demo/prompt.md` | Pool prompt |
| `packs/demo/implement.md` | Runtime Go-template prompt using explicit task input |
| `packs/demo/workflow.star` | Start one step, validate summary output, explicitly finish |
| `packs/demo/events.star` | Validate custom task payload, return run actions, explicitly ignore timer |
| `packs/demo/capabilities.star` | Independently tested HTTP/shell/Jev functions, not registered callbacks |
| `event.json` | Custom source payload, not an ingest envelope |

The resolved pool `demo-dev` has capacity zero. The custom trigger
`demo-demo-event` is disabled. Do not raise capacity, enable the trigger
or deliver the payload to a live root just to test documentation. Copy
the module to a disposable root and follow the
[operator runbook](../../../../docs/verify/cue-starlark.md) for deliberate
end-to-end testing. The demo is not a production implementation/review
pipeline or an authoritative reconciliation source.
