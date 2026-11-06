# Pool labels and a grouped Pools view (JOB-81, 2026-11-06)

## Labels

- `pools.<name>.labels: [..]` is a list of strings, cosmetic: nothing
  schedules on it. A label is non-empty, at most 64 bytes, with no
  whitespace or comma. A number or map is rejected at load (the typed
  decode would turn `1` into `"1"`, so `checkRawPoolLabels` checks the
  raw YAML first).
- A pool a packs entry declares gets `instance:<name>` and
  `pack:<pack>`. `<name>` is the entry's prefix without trailing `-` or
  `_` (`sap-` gives `sap`), or the pack directory name with no prefix;
  `<pack>` is the pack directory name. The project is not a label: the
  RPC returns it as `project` and Desk can group on it.
- A host entry for a pack pool unions its labels with the pack's, where
  other fields merge per field (`unionLabels`). Labels are sorted and
  deduplicated after load, so the RPC order is stable. Host pools that
  no packs entry declares get only what they set.
- `pool.list` (and `api/pools`) return `labels` and `project` per pool,
  and `labels` per trigger (the instance and pack labels of the entry that
  declares it). `pool.list` takes `label: [..]`: every one must match,
  `k=v` matching the label `k:v`. A filtered list drops the quarantined
  and paused pools it filters out. `jobshop pool list --label x
  --label k=v` sends it.

## Desk view

- Rows group by `instance` (default), `pack`, `project`, any other label
  key in use (`k` groups by the value of `k:v`), or `none`. State is in
  the URL: `group`, `q` (name substring), `label` (comma-joined labels,
  all required) and `open` (groups expanded; `t:<name>` for triggers).
- A group is one summary row: pools, desired/open/working/idle totals
  and counts of quarantined, paused, drained and target-overridden
  pools. Quiet pools (no sessions, nothing queued, no target, not
  quarantined, paused or drained) fold into "N idle pools"; a group whose
  pools are all quiet is one row. Click a group row or the fold row to
  expand. Quarantined and paused pools show their reason in the row, not
  in a second table. "Auto all" returns a group's target overrides to
  auto.
- Triggers group the same way. A trigger that fires next or has stalls
  always shows; waiting ones fold, and stalled groups sort first.
- `ui/e2e/pools.spec.js` mocks about 40 pools in 3 instances (group,
  fold, expand, search, label chips, group by project and none). It
  needs a Desk to serve the UI but no live controller data.

## Not done

Label candidacy (a ticket or step selecting pools by label) is not
implemented. The design question is on #569.
