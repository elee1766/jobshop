# #262 Bubblewrap sandbox for agents

2026-11-04. Built on acphost v0.2.0 `Spec.Sandbox`.

- Config: `sandbox: {backend, ro, rw, hide, network}` under `defaults`,
  on agents and on pools. Backend and network: pool over agent over
  default. ro, rw and hide are appended in that order. `~/` is the
  controller's HOME. Top-level `shared_tools` are read-only in every
  sandboxed session. `backend: none` on a pool turns it off.
- Mounts the controller computes per session:
  - ro (skipped if missing): `/nix/store`, `/etc`, `/run/current-system`,
    `/bin`, `/usr/bin`, `/lib64` (nix-ld), `/run/systemd/resolve` and
    `/run/nscd` (DNS), the jobshop binary's dir, the agent command's dir,
    `shared_tools`, configured ro.
  - rw: the session dir (holds `cli.sock` and the worktree), the pool dir
    when not under it, the pool repo's git common dir (worktree commits),
    configured rw (created if missing).
  - Everything else, including `admin.sock`, `jobshop.db`, `secrets/` and
    other sessions, is not mounted.
- Env: bwrap clears it. The spec env plus HOME, locale, cert, Go, nix-ld,
  `GIT_*` and `MISE_*` names from the controller are kept.
- The sandbox config feeds `EnvHash`, so a reload respawns idle sessions
  as an env change does.
- bwrap missing, or a mount source missing, is a permanent spawn error:
  the pool is quarantined (#146). No fallback to unsandboxed.
- `jobshop session sandbox <id>` prints the mounts and bwrap argv.
- NixOS: `services.jobshop.sandbox.enable` adds bubblewrap to PATH, sets
  `defaults.sandbox` (backend and per-user rw caches: go, npm, jcode,
  mise) and the mise env (`MISE_ALL_COMPILE=false`,
  `MISE_PYTHON_COMPILE=false`, `NIX_LD*`, `MISE_TRUSTED_CONFIG_PATHS`).
  `mise` is in `defaultTools`. The host must enable `programs.nix-ld`.
- Not built yet (#316): the shared read-only mise install dir filled by a module
  unit (`MISE_SHARED_INSTALL_DIRS`); hosts set it through
  `environment` and `sandbox.sharedTools` for now.
