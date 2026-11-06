# JOB-40: pure ticket rules in pkg/store/storerules

2026-11-06.

`pkg/store/storerules` holds the I/O-free rules both stores must apply
the same way: claim checks and the fence (`CanClaim`, `StillCandidate`,
`Claim`, `Holds`), release, handoff, close, reopen and expire transitions
with step failure counting, the revision bump (`Bump`), update diffs
(`ApplyUpdate`) and the fields-derived columns (`FieldColumns`). Each
returns the history diff and the controller note; the store writes them.

`pkg/store/sqlstore` calls these and keeps only SQL, transactions and
glue. No behavior change. When the store splits into separate SQLite and
Postgres packages (JOB-34), both call the same rules instead of copying
them. The package has no SQL and imports only `pkg/store`.
