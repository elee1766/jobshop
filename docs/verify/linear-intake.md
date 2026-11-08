# Linear intake live check (#287)

Needs a packs/github instance with `tracker: linear`, `linear_team` set,
and an identity covering that team (`jobshop identity check <TEAM>`
passes). Use a throwaway team or a label only the drill uses
(`linear_intake_labels: jobshop-drill`).

1. Create `<TEAM>-a` in a state from `linear_intake_states` with the
   intake label, and `<TEAM>-b` the same plus a `linear_skip_labels`
   label.
2. Within `linear_poll` plus one tick, `jobshop trigger list` shows
   `scan:linear_issue_no_run:<TEAM>-a` under the `scan` trigger, and
   not `<TEAM>-b`.
3. A `triage` run starts for `<TEAM>-a`; its `diagnose` step runs
   `jobshop linear eligible` and starts `pr-work` with key
   `linear-<TEAM>-a` (`jobshop run list` shows it, `meta.linear.issue`
   is the key).
4. After the next poll the finding is gone from `trigger list`, and no
   second triage run starts for `<TEAM>-a`.
5. Move `<TEAM>-a`'s state out of `linear_intake_states` and cancel its
   run: no new finding after the next poll.
6. `journalctl -u jobshop | grep linear_issue_no_run` shows no errors,
   and no token text.

Record the run ids and outcome on #287.
