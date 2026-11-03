# Two repos on one host (#184)

Add a second packs/github entry with `prefix: <p>-` and the second
repo's vars, then `jobshop reload`. `jobshop config show` lists
`<p>-dev`, `<p>-plan`, `<p>-refinery`, `<p>-deploy-check`, workflows
`<p>-pr-work` etc. and triggers `<p>-replan`, `<p>-master-red`,
`<p>-stall`, while selfdev's names are unchanged. `jobshop run start
<p>-pr-work --var issue=<n> --key issue-<n>` creates steps in the
`<p>-` pools with `meta.github.repo` set to the second repo, and the
selfdev pools keep working.
