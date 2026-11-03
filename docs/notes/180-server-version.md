# Server version for deploy checks (#180)

2026-11-03. A session's `jobshop` binary is fixed when the session
starts, so after a self-deploy `jobshop version` kept reporting the old
commit (b9477e8 from 02:11Z to 02:39Z while the controller ran 085c635).
Post-merge steps then waited on, or failed for, a deploy that was live.

Built: an RPC `version` open to every caller returns the controller's
version and commit (`serve` passes them in `control.Options`).
`jobshop version --server` prints it in the same
`jobshop <version> (<commit>)` form; plain `jobshop version` is unchanged.
The selfdev pack's `deploy_check` uses `--server`. `TestVersionServer`
in `pkg/cli` checks the server commit is reported.
