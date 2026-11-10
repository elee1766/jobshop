# Verify: In flight activity (#1185)

Run on an isolated controller (`docs/verify/desk-redesign.md`).

1. Start a `pr-work` run. Claim its step and release it with
   `jobshop release --until 48h --time-wait -r "gather data"`. In flight shows
   the run as `running` and activity `deferred gather data · until <time> (in 2d)`.
2. `jobshop wait add 'github.pr:pr=<repo>#<n>,state=merged' --reason "merge"` on
   a step: activity reads `waiting on an event`, with the reason.
3. Label a step `needs-human`: activity reads `waiting on a person`.
4. Let an agent session hold a step: activity reads `in a turn` and the tool
   while a turn runs, `held, between turns` after it. Stop the session host:
   `host lost`.
5. Each row stays in the table through every change, and the run column keeps
   the run state.
