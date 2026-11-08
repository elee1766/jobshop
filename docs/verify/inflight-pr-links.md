# In-flight PR links for an issue-started run (JOB-131)

2026-11-08. Run only against an isolated controller (see
`compact-run-links.md`). Add a disposable workflow with an `issue` var and one
agent step in a pool with `max_sessions: 0`, so nothing runs:

```yaml
workflows:
  issue-accept:
    vars: [issue]
    issue: issue
    repo: acme/widget
    steps:
      - {id: implement, pool: dev, prompt: "Implement", title: "Implement {{.issue}}"}
```

```sh
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 DESK_ISSUE_WORKFLOW=issue-accept \
  npm --prefix ui run test:e2e -- run-links.spec.js -g JOB-131 --workers=1
```

The spec starts a run, checks the row has no PR link (a dash in the wide
layout), comments a mention and then a `PR <url> head <sha>` line on the step,
and checks the PR appears without a reload on desktop and mobile, and that the
mention does not. It cancels the run at the end.
