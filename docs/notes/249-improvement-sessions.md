# Operator improvement sessions (#249)

2026-11-04. An improvement session is an `operator` pool session keyed
`improve:<slug>`. Start one from the admin socket (`jobshop send`
gained `--key`, which sends by pool and key through `session.send` and
prints the session id; later sends with the same key reach it):

```sh
jobshop send pool:operator --key improve:dev-prompt "<what to improve>"
```

Desk's "New session" keys sessions `desk:<id>`, so they get no
improvement prompt; use the command above.

- Prompt: the pack's prompt rule `operator-improve` (`at: after`,
  `match: {pools: [operator], keys: ["improve:*"]}`) adds
  `packs/operator/prompts/improve.md` after the pool prompt on the
  session's first message turn. Prompt rules gained `match.keys`: globs
  over the session key; the rule command event carries `key`.
- Worktree: the pool's `repo: ${checkout}`, so the session works in its
  own worktree of Jobshop's checkout at origin/master.
- Paths: (a) file the issue on the right repo (from `config show`
  instances) and start `<prefix>pr-work`; (b) branch `issue-<n>`, push,
  `gh pr create`, start `<prefix>pr-work` and comment `PR <url> head
  <sha>` on its `implement` step, so the dev session adopts the PR and
  the refinery merges it (as PR #246 for #208).
- Never master: packs may now declare `hooks:` (appended to the host's,
  pools prefixed like rules). The operator pack adds an `on: action`
  hook for `push` in the `operator` pool that denies a target of
  `master` or `main` in any ref form (`origin/master`, `HEAD:master`,
  `refs/heads/main`) and a push with no target. The prompt has the
  session run `jobshop guard push --target <branch>` before each push.
  The guard is advisory like every action hook (#188): it stops an agent
  that follows its prompt, not one that skips the call. Branch
  protection on the repo is the hard stop.
- Tests: `TestImprovementSession` (real pack: prompt, worktree, a matrix
  key without the prompt, push denials), `TestOperatorPackLoads`.
