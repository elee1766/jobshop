# Desk trace timeline

2026-11-09. Run on an isolated controller (see `desk-redesign.md`). Never
seed a live instance.

1. Build the branch (`npm --prefix ui run build`, then `go build -o
   /tmp/tl/jobshop ./cmd/jobshop`), `jobshop init` a root under /tmp, serve
   with `--tick=1s`, and serve Desk from the same binary.
2. Create a parent task with a child (`--parent`), add `needs-human` to the
   child, comment on it, close it. Open `?view=trace&ref=<parent>`, press
   `timeline`.
3. Expected: one row per ticket, the child nested under the parent. The child
   has queued then yellow human segments ending at its close, and a comment
   tick. An open row's bar reaches the dashed now line. `graph` returns to the
   graph and keeps the ref. Collapse and expand work from the keyboard. The
   detail under the chart lists every segment with exact times.
4. A step with a wait: park it (`jobshop wait add 'github.checks:...'`). Its
   bar shows a striped waiting segment, and the detail names the condition and
   links the PR.
5. Check 1000, 790 and 390 px, light and dark: no page overflow, the chart
   scrolls inside its own region, the label column stays while scrolling.
6. `DESK_URL=<desk> npm --prefix ui run test:e2e -- timeline.spec.js` passes
   on both projects, also with `DESK_PREFIX=1`.
7. `curl -s '<desk>/api/trace/<parent>?timeline=1' -H 'X-Jobshop-Desk: 1' |
   jq .timeline` returns the rows. Without `timeline=1` the key is absent.
