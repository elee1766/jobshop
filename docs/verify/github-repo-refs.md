# Work references carry the repo (#183)

After a deploy that includes #183, on gasserver:

1. `jobshop ticket list --kind run | jq -r '.[] | select(.meta.github) | "\(.id) \(.meta.github)"'`
   Every line has `"repo":"gfx-labs/jobshop"`, including runs created
   before the deploy (the startup backfill).
2. The controller journal has `backfilled meta.github.repo` once, on the
   first start after the deploy, and not on later restarts.
3. A new `pr-work` run (`jobshop run show <id>`) has the repo on the run
   and every step.
