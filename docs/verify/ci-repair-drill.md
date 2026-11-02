# CI repair drill

Added 2026-11-02 (#117) for the #82 box "CI repair". It uses a red
throwaway branch, so master stays green.

1. Push a branch `drill/red-ci` from master with one failing test (e.g.
   `func TestDrillRed(t *testing.T) { t.Fatal("ci drill") }` in
   `pkg/ids`). Do not open a PR.
2. Wait for its CI run to fail and take its SHA:

   ```sh
   gh run list --branch drill/red-ci --workflow ci --status completed --limit 1 \
     --json conclusion,headSha --jq '.[] | select(.conclusion == "failure") | .headSha'
   ```

3. Start the run by hand (or point a drill workspace's `ci_branch` at
   `drill/red-ci` and let `master-red` start it):

   ```sh
   JOBSHOP_SOCKET= jobshop run start ci-repair --var sha=<red sha> --key drill-ci-$(date +%F)
   ```

Check:

- The `fix` step is claimed by a dev session and comments on its ticket
  (`jobshop run show <run>`).
- At most 2 attempts on `fix` for that SHA: `failures` never exceeds 2
  and a repeated start with the same key returns the same run.
- The failing test is not on master, so the fix cannot land there. The
  run ends `needs_human` with one "Escalate run <run>" ticket in `plan`.

Post the run id and outcome on #82. Then close the escalation ticket,
close any PR the drill opened, and delete the branch:
`git push origin --delete drill/red-ci`.
