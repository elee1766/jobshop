# Per-session isolation (#540)

Added 2026-11-05. Needs the setup from `isolation.md` first.

1. Set `per_session: {range: 2}` on a pool's `isolation.user`, with
   `groups: [jobshop]` and `max_sessions: 3`. Queue three tasks.
2. Two sessions start, not three. `jobshop session sandbox <id>` shows
   slots 0 and 1, mapped to `map_uid` and `map_uid+1`.
3. From a session, `ls` the other session's worktree: permission denied.
4. `getfattr -n system.posix_acl_access <worktree>` (or `getfacl`)
   lists `user:<map_uid+slot>:rwx`, `group:jobshop:rwx`, `other::---`.
5. End one session; the third starts and reuses the free slot.
6. `go test ./pkg/orchestration/control -run TestPerSessionIsolation -v` as the
   jobshop user passes (not skipped).
