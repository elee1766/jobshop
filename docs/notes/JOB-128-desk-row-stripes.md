# Desk alternating row backgrounds (JOB-128)

2026-11-08. Repeated rows in Desk alternate between `--paper` and the
`--row-alt` token (`ui/shell/tokens.css`, one value per theme).

- Opt in with the `rows` class in `ui/shell/app.css`: on a table it
  stripes even `tbody` rows; on any other element it stripes even
  children (list items, articles, paragraphs) and gives them a small
  inline inset so text does not touch the stripe edge.
- The rule uses `:where`, so it has zero specificity: `--hi` hover,
  `.chosen`, `.stalled`, status colors and focus outlines always win.
- Applied to home (waiting, not moving, in flight), the chat picker,
  trace search results, trace relationships and history, learn
  (the rows of each stage group and the measures table) and sys (automation audit, pools,
  resources, ingest, scheduler, sessions). Markdown and chat transcript
  lists are not striped.
- Checked on an isolated controller with populated waiting items,
  relationships and history, desktop 1200px and Pixel 7, light and
  dark, with hover, a focused item and keyboard focus.
