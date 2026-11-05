# trigger.list evaluates triggers in parallel (#522)

2026-11-05.

trigger.list (and the Desk pools view, through /api/pools) evaluated each
trigger in turn. Most of the time is check commands, often `gh` calls of
a second or more, so the total was the sum: about 4.7s on gasserver.

Each trigger is now evaluated in its own goroutine and the results are
returned sorted by name. The list stays live (no cache), so it still shows
what the next tick would do. Evaluation only reads the store and runs
checks, so parallel evaluation does not change results. The supervisor
tick (runTriggers) stays sequential because it fires.

Test: `TestTriggerListParallel` (two checks that each wait for the other).
