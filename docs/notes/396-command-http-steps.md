# Command and http steps (#396)

2026-11-04.

A workflow step has `type: agent` (default), `command` or `http`. The
controller runs command and http steps itself: no session, no turn.

```yaml
steps:
  - id: wait-ci
    type: command
    run: 'gh pr checks ${pr} --required || exit 75'   # list = argv, string = sh -c
    timeout: 10m          # per attempt, default 10m
    retry_after: 2m       # after "not yet", default 1m
    concurrency: 4        # attempts of this step at once, default 4
    on_restart: retry     # or fail
  - id: enqueue
    type: http
    url: http://127.0.0.1:9000/merge
    method: POST          # default
    headers: {X-Repo: "${repo}"}
    header_files: {Authorization: /run/secrets/mq}   # trimmed file contents
    body: '{"sha":"${steps.wait-ci.json.sha}"}'
  - id: review
    pool: refinery
    title: Review
    prompt: ...
```

- Ticket: the step ticket has no pool and `fields.type` command or http,
  so no session claims it and the orphan report skips it. Its project is
  the optional `pool`'s, else `default_project`; that pool also gives
  the working dir (pool dir, else repo, else the root).
- Each tick the controller claims ready ones as holder `controller`
  (lease = timeout + 1m) and runs one attempt in a goroutine.
- Results: exit 0 / 2xx closes done. Exit 75, or HTTP 409, 425, 429, 503
  (honoring `Retry-After`), releases with `defer_until`, not counted as a
  failure. Anything else, or a timeout, closes failed: counted against
  `retries` and escalated as for agent steps (#90).
- Env: the controller's, plus `JOBSHOP_VAR_<NAME>` per run var,
  `JOBSHOP_STEP_KEY` and `JOBSHOP_TICKET`. http sends the same key as
  `Idempotency-Key`. The key is `<run>/<step>/<fence>`: a new claim gets
  a new key, a resumed one keeps it.
- Output: stdout or the response body, first 64KiB, redacted like
  conversation export, in `fields.output`; `fields.command` and
  `fields.result` (ok, not_yet, failed); a comment per attempt with the
  stderr tail on failure. `jobshop run show` prints them.
- Outcome: a JSON `outcome` field or a last stdout line `outcome: <word>`
  goes to `fields.outcome`, for conditional steps (#385).
- Templates in run, url, body and headers: run vars, pack vars,
  `${steps.<id>.output}`, `${steps.<id>.outcome}` and
  `${steps.<id>.json.<dotted.path>}` of an earlier step, checked at load.
  In a `template: go` pack (#383) the outputs are the `steps` var:
  `{{.steps.probe.output}}`, `{{.steps.probe.json.sha}}`, or
  `{{(index .steps "wait-ci").outcome}}` for ids with a dash.
- Restart: a step held by `controller` with no attempt in the new
  process is run again under the same claim and key (`on_restart:
  retry`), or released as one failed attempt (`fail`). An attempt cut by
  shutdown is not recorded.
- Metrics: `jobshop.step.duration` and `jobshop.step.results` by step,
  type and result.
- Not built: sandboxing command steps and a per-run worktree (#419).
- Tests: `TestAutoStepsMixedRun`, `TestAutoStepGoTemplate`, `TestAutoStepHTTPRetryAfter`,
  `TestAutoStepFailureEscalates`, `TestAutoStepRestart`,
  `TestAutoStepHTTPRestart`.
