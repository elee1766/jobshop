# Home sections across widths

2026-11-08. Run only against an isolated controller. Define a disposable
workflow accepting `repo`, `pr` and `issue` variables that stays open, for
example one command step with `run: 'false'` (see `compact-run-links.md`).
Rebuild the frontend and the Go binary, restart Desk, then:

```sh
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  DESK_LINK_WORKFLOW=link-acceptance \
  npm --prefix ui run test:e2e -- home-sections.spec.js run-links.spec.js --workers=1
```

Expected on desktop and mobile projects:

- Waiting, not-moving and in flight are all visible on home at 412 and 1100
  pixels, and resizing both ways keeps the same page and the in-flight filter.
- The phone nav `in flight` and `waiting` buttons scroll to that section and
  add `#inflight` or `#waiting`. Back and forward return to them.
- `?view=home#inflight` opens scrolled to in flight and shows the same
  section at desktop width.
- Going to another page and back keeps the filter.

Also look at a populated 412 pixel view: the nav highlights the chosen tab and
no home section is hidden. Against a build from before JOB-130 the spec fails.
