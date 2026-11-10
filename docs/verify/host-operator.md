# Host-level operator

Check the operator pack on gasserver after it is included once at host
level (#248, #250). Run from the admin socket.

1. `JOBSHOP_SOCKET= jobshop config show | jq '[.config.pools | to_entries[] |
   select(.value.type == "chat") | .key]'` prints `["operator"]`: one
   operator pool, named `operator`, so Matrix threads and Desk keep
   working.
2. `JOBSHOP_SOCKET= jobshop config show | jq '.instances[] | {path,
   prefix, repo: .vars.repo, jobshop_repo: .vars.jobshop_repo}'` lists
   three instances: selfdev (repo `gfx-labs/jobshop`), `acphost-` (repo
   `elee1766/acphost`) and operator (`jobshop_repo: gfx-labs/jobshop`).
3. Start or pick an operator session (`jobshop send pool:operator --key
   verify:host-operator "list the instances you serve"`, or `jobshop
   session list`). `JOBSHOP_SOCKET= jobshop session prompt <id>` shows
   the operator prompt naming `jobshop config show`, `instances` and
   `<prefix>pr-work`; the session's reply names both the selfdev and the
   `acphost-` instance.
4. From the session's socket (`JOBSHOP_SOCKET=<session dir>/cli.sock`),
   `jobshop run list`, `jobshop session list` and `jobshop attention`
   succeed (the pool's admin grant).
5. Ask the session to start work on an acphost issue: the run it starts
   has workflow `acphost-pr-work` (`jobshop run show <run>`), and its
   steps land in `acphost-dev`.
