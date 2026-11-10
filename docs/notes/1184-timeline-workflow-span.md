# Timeline: a run is not "in a turn"

2026-11-10. Issue #1184.

A run or batch has a `working` segment for its whole life and no actor. Desk
now treats a `working` span with no actor as a workflow span: its text reads
`running, no agent holds it` (not `working`), and its bar and swatch are a
light fill with an outline (`data-workflow`), not the solid working color that
a turn uses. Child steps keep their own bars, so a deferred step still shows
as deferred under a running run.

Claims about an active turn come only from a held span (`actor` set) with turn
rows and the live holder state (`holderState`), as before. No server change.
