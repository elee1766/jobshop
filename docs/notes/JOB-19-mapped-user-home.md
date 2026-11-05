# Mapped user home and credentials (JOB-19)

2026-11-05. A pool with `isolation.user` runs its agent as a subuid that
is not the jobshop user, and the jobshop home (dataDir) is 0700, so the
module's default `sandbox.rw` paths under `~` could not be bound: bwrap
crash-looped with `Can't find source path ... Permission denied`.

- Each such pool gets a home at `<root>/homes/<pool>`
  (`<pool>-<slot>` with `per_session`), granted to the mapped uid and
  the pool's first group by POSIX ACL (as for #540 session dirs).
- `sandbox.rw` paths starting with `~/` bind from that home to the same
  path inside the sandbox (`~/go` is `<root>/homes/<pool>/go` outside,
  `$HOME/go` inside). Absolute rw paths are unchanged. Unsandboxed
  sessions get `HOME=<root>/homes/<pool>`.
- The root, `sessions` and `homes` dirs get an ACL entry giving the
  mapped uid search (x) only, so it can reach its own dirs but not list
  them.
- New agent field `credentials: ["~/..."]`: login files copied into the
  mapped home before every spawn (atomic rename), then shown at the same
  path inside. A missing file is a warning. Per agent:
  - jcode: `~/.jcode/auth.json` (its OAuth login; jcode also writes
    usage caches under `~/.jcode`, which the rw `~/.jcode` covers).
  - Refreshed tokens written by the agent stay in the mapped home; the
    next spawn overwrites them with the controller's copy.
- Before spawn the controller checks every mount source and the cwd
  against the mapped uid's mode bits and ACLs. A miss is one permanent
  error naming the path, mode and owner (pool quarantined once), not a
  bwrap crash loop.
- NixOS: `isolation.enable` adds nftables to the service PATH (pasta's
  `deny` runs nft).

Tests: `TestCheckMountsModes`, `TestGrantSearchKeepsEntries`,
`TestPrepareMappedHome`; `TestMappedUserHome` (skips without subuid,
newuidmap, bwrap) starts a mapped-user session from a 0700 root with a
`~/` rw path and checks the cache file is owned by the mapped uid.
