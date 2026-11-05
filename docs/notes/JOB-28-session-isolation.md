# Per-session isolation: home freed with the session (JOB-28)

2026-11-05. Per-session ids, ACL'd worktrees and per-slot homes were
built in #540 and JOB-19 (`540-per-session-isolation.md`,
`JOB-19-mapped-user-home.md`). This closes the gap left:

- A `per_session` session's home `<root>/homes/<pool>-<slot>` (its
  copied credentials and caches) is deleted when the session ends, with
  its worktree. The next session given that slot starts with a fresh
  home and cannot read what the previous one left. Per-pool homes
  (no `per_session`) are shared by design and kept.
- `TestPerSessionIsolation` now also checks that each agent reads its
  own home and not the other session's, and that both homes are gone
  after the sessions close. It skips without bwrap, newuidmap and a
  subuid range, like the other isolation tests.
- Config: `pools.<p>.isolation.user.per_session.range` (set through
  `services.jobshop.settings` on NixOS; the module asserts the range
  fits `isolation.subIdStart/subIdCount`).
