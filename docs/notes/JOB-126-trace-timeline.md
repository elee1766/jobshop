# Trace timeline, compact cards and bars (JOB-126)

2026-11-08.

- Time axis: one scale (`graphLayout` start and px/ms) places cards and the
  axis. The axis is pinned to the top of the graph viewport and positioned
  through the same camera transform as the canvas, so ticks stay on their
  cards while panning and zooming. Tick marks are CSS rules on the axis line,
  with faint guides into the canvas.
- Labels: `start <date>` and `now <date>` (any ticket open, dashed running-color
  mark) or `ended <date>` (all closed, at the last close). Endpoint labels stay
  in view at the edges, with an arrow when their time is off-screen;
  intermediate labels are offsets from start (`+1h 30m`),
  picked so they sit about 88px apart at the current zoom.
- Exact time: each label is a button. Hover, keyboard focus or a tap shows the
  full date with seconds and time zone; Escape or another tap hides it. The
  button's accessible name holds the same text.
- Cards: title first, then one row with the status letter (color plus a
  `role=img` label such as `closed · done`), ID, kind, pool and step, then one
  row of created → closed times. Only what the letter does not say (a run's
  state, the resolution) is shown as text. The inspector header drops its
  duplicated status word the same way.
- Bars: the graph heading is a tag bar like `/trace`. Inputs and selects in tag
  bars take the button height. The trace search field is about 36 characters
  wide and shrinks before the bar wraps. Search behavior is unchanged: no
  dropdown.

Check: `docs/verify/trace-axis.md`.
