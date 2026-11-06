# Self-deploy starts on CI success (#826, 2026-11-06)

- The deploy pack has a `self-deploy` trigger (JOB-62 event triggers):
  a `workflow_run` completed delivery with conclusion success for the
  `ci_workflow` on the deploy branch runs `start_command`, which starts
  `jobshop-self-deploy.service` with `--no-block`. A merge deploys within
  a minute of CI instead of up to `selfDeploy.interval`.
- The timer is unchanged and is the fallback for a lost webhook. The
  deploy script still picks the newest green commit itself, so a start
  while a deploy runs does nothing and the commit merged meanwhile is
  picked up by the next start or timer run (the second merge of #774 and
  #775).
- The module passes the pack `repo` (the selfDeploy workspace's repo),
  `branch` (`selfDeploy.branch`) and `start_command`
  (`systemctl [--user] start --no-block jobshop-self-deploy.service`).
- On NixOS the module adds a polkit rule that lets the jobshop user start
  only that unit (verb `start`). Under home-manager the controller
  already is the user that owns the unit.
- Needs the repo's GitHub webhook to send `workflow_run` events to
  ingest. Without it the trigger never fires and the timer deploys as
  before.
- Each start is a closed `trigger-command` message ticket. A failing
  start (for example the polkit rule missing) stays open as needs-human
  with the output.
