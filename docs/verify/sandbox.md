# Sandbox check (#262)

Run as the jobshop user after enabling `services.jobshop.sandbox.enable`
and `sandbox: {backend: bubblewrap}` on one pool (first `acphost-dev`).

1. `jobshop session sandbox <id>` for a session of the pool: lists ro,
   rw and argv, no `error:` line.
2. In the session (ask the agent to run, or `session attach --rw`):
   - `ls /var/lib/jobshop` shows only `sessions/<own id>` and the
     workspace repo's `.git`; no `admin.sock`, `jobshop.db`, `secrets`.
   - `jobshop ticket show` works (cli.sock).
   - `git commit --allow-empty -m t && git push` in the worktree works.
3. Tools through mise, inside the session:
   - `mise exec go@1.25.4 -- go version`
   - `mise exec node@24 -- npx -y cowsay hi`
   - `mise exec python@3.13 -- python -c 'import ssl, sqlite3'`
   - `mise exec aqua:astral-sh/uv -- uv --version`
   - a `mise.toml` with `[tools] node = "24"` in the worktree, then
     `mise exec -- node -v` there runs without a trust prompt.
4. Failure mode: set `bwrap` off PATH (or a missing `ro` that is not
   skipped) and spawn: the pool is quarantined with a `sandbox` reason.
5. After a day without sandbox-caused failures, enable for all pools and
   record the evidence on #262.
