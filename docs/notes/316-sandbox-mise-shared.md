# Shared read-only mise installs for sandboxes (#316)

2026-11-04. Follows #262.

`services.jobshop.sandbox.mise.tools = [ "node@24" ... ]` (with
`sandbox.enable`) adds:

- `jobshop-mise.service`, a oneshot that runs `mise install --yes <tools>`
  as the jobshop user, outside the sandbox, with `MISE_DATA_DIR =
  sandbox.mise.dir` (default `/var/lib/jobshop-mise`). It runs before
  `jobshop.service` and again on `jobshop-mise.timer` (`sandbox.mise.interval`,
  default daily).
- `<dir>` in `shared_tools`, so it is mounted read-only in every sandboxed
  session.
- `MISE_SHARED_INSTALL_DIRS=<dir>/installs` in the controller env (passed
  through to agents by the `MISE_` env keep rule), so sessions use those
  installs and do not download them.

Empty `tools` (the default) adds nothing.

Concurrency: sessions never write the shared dir. A version that is not
shared is installed into the session's rw `~/.local/share/mise`, which all
sessions share. mise locks each install, so two sessions installing it at
once wait for each other. That was not measured under bwrap. Check it with
`docs/verify/sandbox-mise-shared.md`.
