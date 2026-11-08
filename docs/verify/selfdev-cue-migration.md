# Self-dev native pack cutover

2026-11-07

Scope: the `jobshop` user's self-development instance on gasserver. Do not change
jobcity. The prefixed acphost and operator packs remain YAML consumers.

## Before selecting the native pack

1. Keep `packs/selfdev/pack.yaml` beside the new CUE and Starlark files until the
   native configuration and workflow lifecycle pass verification.
2. Compare resolved legacy and CUE-static configurations with the deployed vars.
   Check pool prompts, escalation prompts, prompt rules, workflow names, retries,
   review loops, trigger keys, cooldowns and caps. Compare command bytes for
   `live-check`; it remains a durable command step, not an inline `shell()` call.
3. Exercise the actual callback workflows with a disposable controller. Check
   selected branches, request-changes retries and terminal outcomes. Do not send
   synthetic events to the live instance.
   Add an unrelated pool and workflow while a callback run is active, reload,
   and verify it can finish. The same run must not gain references introduced
   by that reload. Changed code, grants or credential identities must still
   stop execution. Verify prepared child-step retry routes against live policy,
   not just the parent callback's direct actions.
4. Install a binary containing the native packs before changing pack selection.
   Check the running executable, not a session's older `jobshop` binary. Wait for
   installed-source synchronization and verify the checkout contains `cue.mod`,
   `packs/selfdev/pack.cue` and the selected scripts before activating the host root.

## Cutover

Use the declarative host configuration in `elee1766/gasserver.nix`.
The generated host CUE root imports the local Jobshop module and supplies the
same self-dev vars. Select it through `jobshop.yaml`'s `pack` field and remove the
self-dev entry from `livePacks` in the same activation. Do not activate both
versions of its triggers. Set `#Config.starlark_workflows: false` for the first
activation and verify CUE-static parity before selecting callback workflows.
Keep `#Config.starlark_triggers: false` to preserve the existing maintenance
trigger checkpoints, cooldowns and caps. Handler triggers are a separate opt-in.

Drain existing self-dev workflow runs before switching static workflows to
callbacks. Keep their logical workflow and step names unchanged. A static run
must not be resumed against a callback-only workflow definition.

The host root and runtime asset paths must remain stable across unrelated binary
upgrades. Use the installed checkout for runtime imports and assets, not a new
Nix store source directory for each release. Keep `followInstalled` synchronization
active even when `livePacks` is empty.

Preflight a candidate source without changing the live checkout. Copy the small
host CUE root into a disposable directory, set its dedicated `source.cue` and
local module replacement to the candidate checkout, and replace both the host
root and checkout paths for `config check`. Rewriting only legacy pack paths does
not redirect native CUE imports. Verify this with a disposable candidate containing
an invalid imported CUE package: preflight must reject that candidate and name its
source path, while the unchanged candidate passes.

## After selection

- Confirm the service is active and the running version is the intended commit.
- Confirm exactly one copy of each self-dev trigger and the unchanged peer packs.
- Inspect callback invocations, selected step tickets, rendered prompts and
  resulting run outcomes through `automation list` / `automation show` and Desk.
- Confirm a source sync and ordinary binary upgrade preserve unchanged hook
  identities for in-flight runs.
- Only then remove the obsolete self-dev YAML. Keep shared GitHub YAML while
  other deployed consumers still load it.

After YAML removal, use `starlark_workflows: false` for static CUE workflows.
A YAML rollback requires restoring a pre-removal source revision as well as
changing host selection. Drain callback runs before reverting to static workflows.
Do not reset or replace the ticket database as part of either direction.

## Isolated acceptance evidence

The generated host configuration was checked with native static workflows and
declarative maintenance triggers. All resolved sections match the legacy host
configuration except workflow `pack_vars`; executable workflow fields match.
Eight disposable-controller starts without supplied inputs verify the actual
default-input path preserves self-dev prompt bytes, pools and retry counts.
The differing `repo` metadata is unused by those self-dev prompts and commands.
This does not replace live cutover verification.

## Live cutover evidence, 2026-11-08

At 04:01 UTC, gasserver's `jobshop` user was running version `0.8.26+1f83223`
with the stable CUE root selected and all ten selfdev workflows using callbacks.
Maintenance triggers remained declarative. Live `live-check` run
`js-00343fhs4g002` was closed with state `done`; its audit contained successful
start, script-step and completion records. This verified native lifecycle allows
removing selfdev YAML. Shared GitHub and operator YAML remain for peer consumers.
