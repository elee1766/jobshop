# Curator gate: cited learnings after the merge (#1061)

2026-11-08. `gateMetrics` counts recurrence: learnings in the window that
repeat a cited learning (same kind, Jaccard >= 0.5 on the summary), and
a cited learning itself (`c.ID == l.ID`).

The session that ships a memory edit records a learning about the fix
after the merge. In the after window that learning is cited and inside
the window, so it counted as the problem coming back. Edit 61 (#1037)
was judged `not_helped` for that reason alone (1 in 23 steps after, 1 in
1056 before).

Now `gateMetrics` takes `after`. In the after window (`editDue`, so
`learn evaluate` and the trigger) a cited learning is never a
recurrence, by ID or by similarity to another cited one. Learnings that
are not cited still count when they repeat a cited one. The baseline
window (before the merge, `recordEdit` and `memory.write`) is unchanged:
there the cited learnings are the problem, so they count.

Edits already decided stay as they are. Edit 61 is not re-evaluated.
