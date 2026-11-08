# In-flight reference columns follow the table's width (JOB-129, 2026-11-08)

The in-flight table showed the Linear issue as an unlabeled line under the
title on desktop, and its PR column and stacked PR fallback switched on the
viewport (899px), which ignored the open chat rail: at 1000px with the rail
open the table is 600px wide.

- Two labeled columns, `issue` and `pr`, show when the table itself is at
  least 720px wide (a container query on `.table-scroll.refs`, so chat rail
  open or closed and window size all count through the space the table has).
  The title cell then holds only the title, which still opens the trace.
- Narrower, both columns hide and the same links stack under the title. Each
  link is in the DOM twice and CSS shows one, so a row never shows one link
  twice. The `.wide` viewport rule no longer applies to these columns.
- Measured on an isolated Desk with Linear and PR references (several PRs,
  another repo, and a run with no PR): columns at table widths 800 and up,
  stacked at 700 and below, no horizontal overflow from 1500px down to 360px.
- `ui/e2e/run-links.spec.js` asserts the labeled headers on desktop, the
  stacked links on mobile, one visible link per reference and no overflow.

Not changed: the `session` column keeps its viewport rule, and trace search
tables are untouched.
