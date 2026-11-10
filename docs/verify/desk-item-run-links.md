# Run links on Linear and GitHub items

2026-11-10. Needs a Desk with the Linear and GitHub views configured
(`docs/verify/desk-linear-github.md`).

1. Open a PR in `?view=github` that has a ship or review run. "linked work"
   lists each run with relation `run`. The run title shows link labels, not
   Markdown or URLs.
2. Click a run row: Desk opens `?view=trace&ref=<run id>`.
3. The PR's description names a Linear key that has an implement run: that
   run is listed with relation `via <KEY>`.
4. Open that Linear issue in `?view=linear`: its own run has relation `run`,
   and the runs of its attached PR have `via <owner/repo#n>`.
5. A PR with no runs shows no run rows.

`go test ./pkg/orchestration/control -run TestBrowseRunLinks` covers the
matching. `ui/e2e/browse.spec.js` covers the rows and the trace link.
