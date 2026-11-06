Implement: {{.task}}

1. Read the task and every comment on this ticket and its run
   (`jobshop ticket show`). If an earlier attempt or a refinery repair
   left a PR, resume its branch: `git fetch origin pull/<n>/head && git
   checkout --detach FETCH_HEAD`, push with `git push origin
   HEAD:<branch>`. Address each finding.
2. Otherwise branch from origin/{{.branch}}, meet the acceptance
   criteria, run `{{.check_cmd}}`, and open a Conventional Commits PR on
   {{.repo}}.
3. Wait for CI with a `jobshop wait` (see your pool prompt), not polling.
   Red: fix, push, wait again.
4. Record the PR URL, exact head SHA, validation evidence and limitations
   with `jobshop comment`. Then `jobshop close`: that hands the PR to
   independent review. Publication alone is not done: close only with
   evidence for every acceptance criterion on that head.
5. A genuine wait (a decision, credential, another team): `jobshop report
   waiting "<what>" --question "<decision needed>" --context "<what you
   did and found, what each answer leads to>"`, with `--option
   "<Title>|<what it does>"` for concrete alternatives. An event (CI, a
   merge, a deploy, a tag): `jobshop wait add` with its kind (`jobshop
   wait add --help`). A real time: `jobshop release --until <time>
   --time-wait -r "<why>"`. Cannot deliver: comment why, then
   `jobshop close -r failed`.
