Triage failing {{.branch}} CI on {{.repo}} at {{.sha}}. Triage only:
no branch, commit, push, PR, revert or merge here.

1. Re-read fresh state: `gh api repos/{{.repo}}/commits/{{.branch}}
   --jq .sha` and `gh run list -R {{.repo}} -b {{.branch}} -e push -L 10`.
   Head green or still running: comment that, `jobshop close --outcome
   green`.
2. Find the first failing push run since the last green one and the PR
   that introduced it (`gh run view <id> --log-failed`, `gh pr list
   --state merged --search <sha>`). An infrastructure flake (runner lost,
   network, cache): `gh run rerun <id> --failed` once, comment it, close
   with `--outcome flake-rerun`.
3. Deduplicate: `jobshop run list` for an open pr-work run whose task
   names this failure, and open PRs that fix it. Found: comment the new
   evidence on it, close with `--outcome duplicate`.
4. Otherwise start one repair: `jobshop run start pr-work --var
   task="Fix {{.branch}} CI: <cause>. Failing run <link>, job <job>,
   breaking PR <link> <sha>. Acceptance: {{.branch}} CI green." --key
   ci-fix-{{.sha}}`. Prefer a forward fix over a revert unless the fix is
   unclear. Close with `--outcome routed`.
