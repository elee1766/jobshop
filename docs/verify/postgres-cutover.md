# Postgres cutover (#632)

Move a host's ticket store from SQLite to Postgres. Not yet run on
gasserver.

1. Prepare: a Postgres 17 database owned by the jobshop user, and a DSN
   file readable only by it (e.g. `/run/secrets/jobshop-pg-dsn`,
   `postgres:///jobshop?host=/run/postgresql`).
2. Stop: `systemctl stop jobshop` (gateway, desk and ingest can stay;
   they reconnect). Check `jobshop whoami` fails to connect.
3. Backup: `sqlite3 /var/lib/jobshop/jobshop.db ".backup /var/lib/jobshop/jobshop.db.pre-pg"`
   (or copy `jobshop.db`, `-wal` and `-shm` together while stopped).
4. Migrate, as the jobshop user:
   `jobshop --root /var/lib/jobshop admin store migrate --to postgres --dsn-file /run/secrets/jobshop-pg-dsn`.
   Expect a row count per table and `ok: row counts match`. It refuses a
   non-empty target; rerun with `--force` only to replace an earlier
   attempt.
5. Switch config: `services.jobshop.store = { driver = "postgres"; dsnFile = "/run/secrets/jobshop-pg-dsn"; };`
   and `nixos-rebuild switch` (or set `store:` in a user-owned
   jobshop.yaml).
6. Start: `systemctl start jobshop`. Check `jobshop ticket list`,
   `jobshop run show` on an open run, `jobshop recall <word>` and Desk
   show the same data as before; check new comments and claims work.
7. Rollback: stop jobshop, unset `services.jobshop.store` (driver back to
   sqlite), switch, start. The SQLite file was never written by the
   migrate, so it holds the state at step 2; anything written to
   Postgres after cutover is not carried back.
