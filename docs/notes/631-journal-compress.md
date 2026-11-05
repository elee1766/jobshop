# Journal compression (#631)

2026-11-05.

- Config `journal.compress.after` (Go duration or Nd; empty is off, the
  default) and `journal.compress.level` (zstd 1-22, 0 means 3). These are
  host config. The NixOS module sets them through
  `services.jobshop.journal.compress.{after,level}`.
- The controller tick starts a sweep on start and then hourly. The
  sweep lists closed sessions. For each one closed before `now - after`
  that still has a plain `journal.ndjson`, it calls
  `host.CompressJournal(dir, level)` (acphost v0.7.0, acphost#74). That
  call writes `journal.ndjson.zst` and `journal.idx`, then removes the
  plain file. The sweep runs in its own goroutine, one session at a
  time, on a locked OS thread set to the idle IO class (`ioprio_set`,
  Linux only). It never runs twice at once, and the tick does not wait
  for it. Each compressed journal logs `bytes_before` and `bytes_after`.
- Readers: `session.subscribe`, `session.journal`, recall indexing,
  journal shipping, `jobshop session attach`, the stall and last-seq
  checks (`journalTail`) and chat all go through acphost's
  `ReadJournal`, `ReadFrom`, `FollowJournal` and `JournalTail`. These
  open the `.zst` when the plain file is gone. They still pass the
  `journal.ndjson` path, and acphost resolves it. jobshop's own
  read-from-the-end tail (#614) is removed: acphost's `JournalTail` now
  reads from the end or from the index for both forms.
- A compressed journal is final. acphost refuses to start a host in
  that session (`ErrJournalCompressed`), and only closed sessions are
  swept.

Tests: `pkg/control/journalcompress_internal_test.go` (sweep cutoff,
off by default, and session.journal, session.subscribe with since and
tail, plus recall on a compressed journal) and
`pkg/control/journaltail_test.go`.

Rollout on gasserver: set `after: 24h, level: 3` in the host config repo
after deploy, then compare `du -sh /var/lib/jobshop/sessions` (about
1.6 GB before). Check with `docs/verify/journal-compress.md`.
