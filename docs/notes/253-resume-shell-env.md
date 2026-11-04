# Resumed session: tool shell PATH (#253)

2026-11-04. Cause: after a resume or reload, jcode runs tool commands
as `bash -lc` (its `JCODE_RELOAD_DETACH_COMMAND`). A login shell reads
NixOS's `/etc/profile`, which sources `set-environment` unless
`__NIXOS_SET_ENVIRONMENT_DONE` is set, and that file exports the system
PATH. The host, `jcode acp` and `jcode serve` had the right env; only
the login shell lost PATH. `JOBSHOP_*` survived.

Fix: the host spec env sets `__NIXOS_SET_ENVIRONMENT_DONE=1`, so the
profile keeps the inherited PATH (jobshop's bin, gh and the rest of the
controller's PATH). Other set-environment vars come from the controller
env as before. Being part of the spec env, the change alters the env
hash, so live hosts are refreshed on the deploy that carries it.

Test: `TestResumeShellEnv` (a stand-in profile, plus the real
`bash -lc` when `/etc/profile` is NixOS's). Runbook:
`docs/verify/respawn-env.md`.
