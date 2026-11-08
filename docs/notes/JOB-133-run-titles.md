# Run titles from the source issue or PR (JOB-133)

2026-11-08.

- A run of a workflow with an `issue` or `pr` var is titled with that
  issue or PR's own title (one line, at most 200 characters), read when
  the run starts: GitHub `issues/<n>` or `pulls/<n>` through the identity
  covering the repo, or the Linear issue through the identity covering
  its team. Before: `Run <workflow> <key>`.
- The title is cosmetic. The dedup key, the workflow, and
  `meta.github` / `meta.linear` (the links) are unchanged. A start that
  returns the open run for its key reads nothing.
- A failed read does not stop the start. The run is titled by its source:
  `Issue owner/repo#n`, `PR owner/repo#n` or `Linear ENG-1`.
- From a pass (trigger, hook handler) the read goes through `offTick`, like
  an intake read: the start returns `errCheckPending` and the replay has
  the title, so the tick never waits on the network.
- `RunStartParams.Title` and `jobshop run start --title` set a title
  explicitly and skip the read. Hook `start_run(title=...)` already did
  and still wins.
- A scan finding about an issue titles its run with the issue's title
  whatever the workflow: `issue_no_run` (subject is a number of the
  signal's repo) and `linear_issue_no_run` (subject is the key). So
  triage runs read as the issue they triage. Other signals (stuck work,
  needs_human, ...) keep `Run <workflow> <key>`.
- `jobshop run start` from a triage or plan prompt and `batch start` get
  the source title the same way, with no change to the prompts.
- github pack: `master_red` titles its run `Repair red <branch> CI at
  <sha7>`. selfdev: `evaluate` and `curate` carry `args.title`
  (`Evaluate recent runs and edits`, `Curate learnings`).
- The acphost instance reuses the github pack's workflows and triggers,
  so it gets the same titles with no change of its own.
- `docs/verify/scan-drill.md` matched runs by title; it now matches
  `fields.key`.
