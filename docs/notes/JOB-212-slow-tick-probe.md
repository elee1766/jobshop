# Slow tick probe (JOB-212)

2026-11-14. `JOB-191-store-latency.md` ends with one tick that took 7.3s
(10:00:33 UTC on selfdev) while every store call, even a 4ms reader-pool list,
took about 90ms or a multiple of it. This note records what is known, what
the controller now captures, and why the cause is not named yet.

## What is known

No trace of that tick exists: the spans of a tick say which step and which
store call were slow, not why. Two readings fit "every call 90ms or a
multiple":

- **The database stalled.** A WAL checkpoint (an fsync of a 1.1 GB database),
  a writer queued behind a long write, or a housekeeping step holding the
  writer (`PRAGMA optimize` took 1.6 to 2.4s on this file). This does not
  fit the reader pool calls: they use their own connections, take no write
  lock and do not wait on a checkpoint.
- **The process stalled.** It was not scheduled, so everything it did took
  longer, whatever the call was. Causes: CPU starvation, cgroup CFS
  throttling (its period is 100ms, so stalls come in multiples of it), a GC
  stop-the-world, memory reclaim.

The reader pool calls point at the second. The host agrees: at the time of
writing, `/proc/loadavg` is 17.6 on 6 cores and `/proc/pressure/cpu` reads
`some avg60=12`, with agent sessions and builds sharing the CPUs. That is
a lead, not a finding. Nothing was changed in the scheduling settings, because
no measurement ties a stall to them.

## What changed

A tick that takes `Options.SlowTick` (default 1s; a quiet tick is under
0.1s) logs one warning, `slow tick`, and puts the same fields on the
`supervisor tick` span (`jobshop.tick.*`). The fields are the evidence for
each candidate, taken as the difference of two snapshots around the tick:

| Field | Candidate it tests |
|---|---|
| `stalls`, `stall_max_seconds`, `stall_total_seconds` | Process not running. A monitor goroutine expects to run every 20ms and records each run that is 40ms or more late. |
| `psi_cpu_seconds`, `psi_io_seconds`, `psi_memory_seconds` | Host pressure: time some task waited (`/proc/pressure`, host-wide). |
| `throttled_seconds` | The controller's cgroup was CPU-throttled (`cpu.stat` `throttled_usec`). |
| `gc_cycles`, `gc_pause_seconds` | Go GC stop-the-world time. |
| `db_written`, `wal_bytes`, `wal_delta_bytes` | The database file was written during the tick: a checkpoint copies WAL pages into it. |
| `pool_writer_waits`, `pool_writer_wait_seconds` (and `pool_reader_*`) | Calls queued for a connection, so a busy writer. |
| `housekeeping` in `hints` | A housekeeping step (outside `tickMu`) ran during the tick: `housekeeping:<step>`. |
| `slowest_step`, `slowest_step_seconds` | Which tick step. |

`hints` lists the candidates the numbers support: `process_stalled`,
`cpu_pressure`, `cgroup_throttled`, `io_pressure`, `memory_pressure`, `gc`,
`db_file_written`, `writer_queue`, `housekeeping:<step>`, or `none`. A hint is
a threshold on the numbers (see `hints` in `tickprobe.go`), not a verdict:
read the fields.

How to read it is in `docs/verify/slow-tick.md`. A pattern like the 7.3s
tick reads as `process_stalled` with `cpu_pressure` or `cgroup_throttled`
when the process was starved, and as `writer_queue` or `db_file_written`
without `process_stalled` when the database was.

## Not done

The cause of the 7.3s tick is not recorded. The next slow tick on selfdev
carries the evidence; add its `slow tick` record below, with the hints, and
then fix it or close JOB-212 with the cause. The probe does not change any
setting.

## Trace

None yet. Format for the entry (from the journal, `grep 'slow tick'`):

    <date> seconds=<s> slowest_step=<step> hints=<...> stalls=<n> stall_max_seconds=<s> psi_cpu_seconds=<s> ...
    cause: <one line>
EOF
