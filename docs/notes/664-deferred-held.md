# Deferred held tickets skip the unclosed check (#664)

2026-11-05. `TestSessionRestartIdle` claims a deferred ticket for an
idle session by id. When a turn (here the first ticket's, or the one
after the env refresh) ended right after that claim, `afterTurn` listed
every in-progress ticket the session held, sent the unclosed reminder,
and the fake agent closed "the current ticket": the deferred one.

`afterTurn` now ignores tickets whose `defer_until` is in the future,
the same rule `ClaimNext` uses for what it prompts. Such a ticket is not
reminded about or released; it stays claimed until its deferral passes,
then the session is prompted with it as usual.
