# Isolation runbook: clear sandbox errors (#677)

2026-11-05.

- The isolated-sessions block in `docs/verify/isolation.md` used to pipe
  `jobshop session sandbox` output into jq and skip failures. Run by a
  caller without the admin socket (selfdev `live-check`, run
  js-00339h6z98002) it printed a jq parse error per session and ended
  with a misleading "no open session of an isolated pool".
- Now a failed `session list` or `session sandbox`, or non-JSON output,
  stops the block with one `FAIL: cannot read sandbox: <err>` line.
  It also passes `--json`, which it relied on but did not ask for.
- The block needs the admin socket (`session sandbox` is admin only).
  A command step runs in its pool's sandbox, which binds no admin
  socket, so under live-check the block fails with that line: run it as
  the jobshop user on the host.
