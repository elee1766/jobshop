# Desk trace detail pane

2026-11-10. Extends `desk-timeline.md` and `desk-timeline-turns.md`. Run on an
isolated controller (see `desk-redesign.md`). Never seed a live instance.

1. Build the branch (`npm --prefix ui run build`, then `go build -o
   /tmp/tp/jobshop ./cmd/jobshop`), `jobshop init` a root under /tmp, serve with
   `--tick=1s`, serve Desk from the same binary.
2. Make a run with a step that fails once: a workflow whose step uses the fake
   agent pool (`isolated-controller` steps 4 and 5) and `retries: 2`. Let the
   first attempt close `failed`, the second `done`. Add a comment and a `PR
   https://github.com/acme/w/pull/7` comment to the step.
3. Open `?view=trace&ref=<run>` and choose the step. Expected, at 1000 px with
   no scrolling in the pane: the title and status letter, `step · <pool> ·
   step <name>`, facts `started`, `ended`, `took`, `attempts` (`2 · failed 1 of 2
   allowed`) and `outcome`, a time by state bar whose parts add up to the
   duration, and an `ended` block with the close reason.
4. `events` tab: newest first, one entry per claim with `ended: ...` saying
   how it ended and why, the PR marker once (not also as a comment), and the
   close. Tick `every history row`: the raw rows appear and the grouped list
   goes. Untick it.
5. Choose the failed bar on the chart: the URL drops `tab=`, the `events` tab
   is shown and that entry is highlighted. There is no second list under the
   chart.
6. Tabs: `input / result` shows `step`, `run` and the last `command` and
   `output` (long output behind its first line), `relations N` lists the parent
   and children with a status letter and a duration, `body`, `comments N` and
   `raw` work. Click one: `?tab=` is in the URL, reload keeps it, `?tab=x` opens
   `events`. With a tab focused, Left, Right, Home and End move between tabs.
7. While a step runs: `now` names the running tool (`turn running: <tool>`),
   `holder` shows the session and its live state, and `view session` opens it.
   Kill the host: `now` turns orange, `host not alive`.
8. A task with `needs-human`: `now` shows the label and `read the question`
   opens the body tab. A `jobshop release --until 1h --time-wait` ticket says
   `deferred until <time>`. A step waiting on CI lists each condition with its
   state and a link.
9. A command or script step has no `attempts` and no `holder`. A batch and a run
   show `n of m closed`. A ticket whose history was compacted says part of its
   history is not kept.
10. Change the ticket (add a comment) while the tab is `comments`: the count
    rises, the tab and page scroll stay.
11. Check 1000, 790 and 390 px, light and dark: no page overflow, the tab bar
    stays in the viewport, every tab and link reachable by keyboard.
12. `DESK_URL=<desk> npm --prefix ui run test:e2e -- trace-detail.spec.js
    timeline.spec.js` passes on both projects, also with `DESK_PREFIX=1`. With
    `DESK_MUTATE=1` also `live.spec.js`, `title-links.spec.js` and
    `trace-session.spec.js` (`live.spec.js` and `trace-axis.spec.js` have
    failures that predate this change, #1166).
    `node ui/e2e/trace-detail-shots.mjs <dir> <prefix>` writes the screenshots.
