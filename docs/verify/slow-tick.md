# Slow tick (JOB-212)

How to catch and read a tick that takes a second or more. Quiet ticks are
covered in `tick-latency.md`.

1. **Find it.** On the host, `journalctl -u jobshop | grep 'slow tick'`
   (or the controller's log). One line per tick of `Options.SlowTick` (1s)
   or more, with `seconds`, `slowest_step` and `hints`. The same fields are
   attributes `jobshop.tick.*` on that tick's `supervisor tick` span, next to
   its `tick <step>` children and the `store.*` spans of calls over 50ms.
   A healthy instance logs none.
2. **Read `hints`.**
   - `process_stalled`: the stall monitor ran 40ms or more late during the
     tick (`stalls`, `stall_max_seconds`), so the whole process was not
     running. Look at `psi_cpu_seconds` (host CPU pressure), `throttled_seconds`
     (cgroup throttling, `cpu.stat` of the controller's cgroup) and
     `gc_pause_seconds`. Stalls of about 100ms with `throttled_seconds`
     above zero are CFS throttling.
   - `writer_queue`: callers waited for the single writer connection
     (`pool_writer_wait_seconds`). Look for `housekeeping:<step>` in the same
     hints: that step held the writer.
   - `db_file_written` without `process_stalled`: a WAL checkpoint ran. With
     `wal_delta_bytes` negative or small and `io_pressure`, the disk was slow.
   - `none`: no candidate matched. Compare `slowest_step` with the `store.*`
     spans; a step slow on its own (a command, an HTTP call) is not a store
     problem.
3. **Drill on a throwaway controller** (isolated-controller skill): run
   `go test ./pkg/orchestration/control -run 'TestTickProbeReport|TestSlowTickLogged'`.
   The first feeds fixture `/proc`, cgroup and database files and expects each
   hint; the second runs a real pass with the threshold at 1ns and expects the
   `slow tick` record.
4. **Live check after deploy.** Nothing logs while ticks are quiet; that is
   expected. To see one record, pin the controller to one CPU and load it:
   `taskset -pc 0 <controller pid>`, then six `taskset -c 0 yes >/dev/null`.
   A tick that passes 1s logs `slow tick` with `process_stalled` and
   `cpu_pressure`; undo with `taskset -pc 0-5 <pid>` and kill the `yes` jobs.
   If ticks stay under 1s nothing logs, which proves nothing either way.
5. **Record it.** Paste the first real `slow tick` record and its cause into
   the "Trace" section of `docs/notes/JOB-212-slow-tick-probe.md`.
