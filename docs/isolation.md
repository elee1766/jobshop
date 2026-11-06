# Isolation

Moved from the README (2026-11-05, #551).

- **Bubblewrap sandbox per session.** Set `sandbox:` under `defaults`, on
  agents or on pools. acphost runs bwrap and Jobshop computes the mounts,
  which are an allow list: system dirs listed in config (`ro`, or a
  named `sandbox_profiles` entry, #625) are read-only, and the session dir
  (with its worktree), the pool repo's git dir and configured `rw` paths
  are writable. `admin.sock`, `jobshop.db`, `secrets/` and other sessions
  are not mounted. The env is cleared to the configured `env_keep` list. A missing bwrap or
  mount source quarantines the pool. There is no unsandboxed fallback.
  `jobshop session sandbox <id>` shows the mounts and argv. On NixOS
  `services.jobshop.sandbox.enable` puts bwrap on PATH and nothing else;
  the sandbox config, agent tools and caches are set by the host ([nixos.md](nixos.md#agent-tools)). See [notes/262-sandbox.md](notes/262-sandbox.md) and
  [notes/625-sandbox-config.md](notes/625-sandbox-config.md). `overlay:`
  merges layer dirs onto one target (ro, tmp or persist), appending
  layers across levels: [notes/JOB-55-sandbox-overlay.md](notes/JOB-55-sandbox-overlay.md).
- **Socket scoping.** A session talks only to its own `cli.sock`, which
  acts as that session (its ticket, pool and project). The admin socket
  is not reachable from a sandbox.
- **Credentials.** Secrets are files, never inline in yaml. A session gets
  a short-lived scoped token with `jobshop identity token <owner/repo>`
  instead of holding a key. Agent text leaving the controller is checked
  with betterleaks and the host's `redact` rules (#574, #578): session events are masked, comments, reports,
  learnings and artifacts are stored masked, and `jobshop gh`/`linear`
  writes holding a finding are refused.
- **Guardrail hooks.** `hooks` gate claims and outside actions (push, PR,
  merge, comments) with allow, deny or ask a human. A broken hook denies.
- **Per-pool users, cgroup limits and network.** Optional: a mapped user,
  cgroup limits and pasta networking per pool. See
  [notes/388-pool-isolation.md](notes/388-pool-isolation.md).
- **Limits.** Without per-pool isolation, all processes run as one Unix
  user, and network is on by default (`sandbox.network: host`). The
  sandbox guards against mistakes and limits blast radius. It is not a
  boundary against a hostile agent. See
  [notes/users-security.md](notes/users-security.md).

```yaml
defaults:
  sandbox:
    backend: bubblewrap
    rw: [~/.cache/go-build]
```
