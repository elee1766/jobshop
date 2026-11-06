# JOB-51: read-only guidance mounts

2026-11-06.

`sandbox.guidance` (defaults, agent, pool, profiles; appended like `ro`)
lists the operator's guidance files and dirs, for example
`["~/.agents", "~/.jcode/prompt-overlay.md"]`. Each is bound read-only
at the same path inside the sandbox; missing ones are skipped.

- A mapped-user session gets its own rw `~/.jcode` from its home
  (JOB-19). acphost binds parents first, so a guidance file under it is
  bound over that rw dir and stays read-only.
- A guidance path equal to an `rw` path is refused (at equal depth the
  rw bind would win), and so is one that holds an agent `credentials`
  file: credentials are copied into the mapped home, never bound.
- Changing the list changes the env hash, so idle sessions respawn.

Not in this change: mapping the controller uid as namespace root
(acphost#162) and narrowing the pre-spawn mount check (JOB-53). Until
those land, a 0700 home still needs the host workaround.
