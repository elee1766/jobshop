# Desk Linear and GitHub views live check (JOB-163)

Needs a controller with a Linear identity for a team (here `JOB`) and a GitHub
identity for a repo (here `elee1766/jobshop`). Use an isolated root, not the
live one.

1. `curl -s -H x-jobshop-desk:1 -X POST $DESK/api/rpc/browse.scope -d '{}'`
   lists the team and repo, and no token.
2. Hold one run per provider so the views have real links. Add two workflows
   with one command step `run: 'exit 75'` and `retry_after: 24h`: `gh-hold`
   (`vars: [pr]`, `pr: pr`, `repo: elee1766/jobshop`) and `lin-hold`
   (`vars: [issue]`, `issue: issue`, `repo: elee1766/jobshop`,
   `tracker: linear`, `linear_workspace: gfx-labs`). Then
   `jobshop run start gh-hold --var pr=<open pr>` and
   `jobshop run start lin-hold --var issue=JOB-163`.
3. Open `?view=linear`: rows match `jobshop linear issues`, the filters and
   search work, and `?view=linear&item=JOB-163` shows comments, relations and
   the held run under "Linked work", opening its trace.
4. Open `?view=github`: pulls and issues of the repo, a pull request with
   reviews and checks, and the held run on its pull request. An item with no
   run shows no run link.
5. `?view=github&repo=someone/else` is refused with the scope message.
6. Post a comment on the open Linear issue. The view shows it within about a
   minute without a reload, and a "new" marker for unread comments.
7. In the chat composer add `linear:JOB-163` or `github:elee1766/jobshop#<n>`
   and send: the agent sees the title and body. A ref outside the scope fails
   with 400.
8. No response or console line contains a token (`grep -r lin_api\|ghp_\|github_pat_`
   over saved responses).

Specs:

```
cd ui
DESK_URL=http://127.0.0.1:<port> npx playwright test browse.spec.js          # mocked, any Desk
DESK_URL=http://127.0.0.1:<port> npx playwright test browse-live.spec.js     # needs steps 1 and 2
```

`node e2e/run-isolated.mjs` runs both in CI. The live spec skips there.
