# JOB-173: storage/store and storage/journal

2026-11-09. Child of JOB-164. Moves persistence and journal ownership under
`pkg/storage`. No behavior change: the same SQL, migrations, RPC and files.

## Layout

| Before | After |
|---|---|
| `pkg/store` and its `sqlite`, `postgres`, `storerules`, `storemigrate`, `storetest` | `pkg/storage/store/...` |
| `pkg/writekey` | `pkg/storage/store/writekey` |
| `pkg/journal` (frame mapper, VictoriaLogs client, recall index) | `pkg/storage/journal` |
| `pkg/control/journalread` (paging, search, recall summary) | files `read.go` in `pkg/storage/journal` |
| `pkg/control/journalship` | `pkg/storage/journal/ship` |
| `pkg/control/journalcompress` | `pkg/storage/journal/compress` |
| `pkg/journal/vltest`, `pkg/control/journalvl` | `pkg/storage/journal/vltest`, `.../journalvl` (test harnesses) |

`storage` is a namespace with no Go package and no Storage interface. SQL
stores, journal streams and artifacts keep their own narrow APIs.

## Decisions

- **journalread merged into journal.** It had one dependency beyond what
  `journal` already imports (`rpc` types, `domain`) and the same callers, so
  a separate leaf had no dependency reason. `journal.Run` is now
  `journal.RunSearch` because `Client.Run` and `Run` shared a name in one
  package. `ship` and `compress` stay leaves: they own goroutines, cursors
  and config, and `compress` needs neither the VictoriaLogs client nor `rpc`.
- **`journalread.Env` moved to control** (`sessionEnv` in `journal.go`). It
  read `spec.json` through `hostspec`, which made the journal family import
  session ownership. The controller already owns the redactor, so it builds
  the env and hands the redactor in. The JOB-165 exception for this edge is
  deleted.
- **writekey moved under `storage/store`**, as JOB-165 decided from callers:
  it is a durable store collection and control is the only caller. The
  controller keeps scoping a key by project and the claim and finish RPCs.
- **Migrations are renamed, not edited**: the goose files keep their bytes
  and numbers, so an existing database reads the same. `//go:embed
  migrations/*.sql` is relative to the package and needed no change.
- **Not touched**: authorization (`canReadJournal`), the ship and compress
  schedules and the redactor stay in control.

## Checked

- `go list -deps ./pkg/storage/journal/...` reaches no `control`, `hostspec`
  or `sessionrt`.
- archcheck ownership rows and rules point at the new paths; the family rule
  `family-journal` no longer needs an exception.
- CI's skipped-test filter (`store/postgres|store/storemigrate|journalvl`)
  still matches the new paths.
