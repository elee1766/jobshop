# Linear as a tracker (#201)

2026-11-03. Work specs can live in Linear while code still goes
through GitHub PRs.

Built:
- Workflows take `tracker: linear` and `linear_workspace: <name>`. The
  `issue` var then holds a key like ENG-123 (uppercased, checked at run
  start). Runs, steps and escalations get
  `meta.linear = {workspace, issue}` and `meta.github = {repo}` for the
  code repo only, so subscriptions and `issue_no_run` never read a
  Linear key as a GitHub number. Attention items show the key as ref.
- `pkg/linear`: a small GraphQL client (issue with comments, comment,
  move state by name). No new dependency.
- `jobshop linear issue|comment|state <KEY>`: the CLI asks the
  controller for the token of the key's team through `identity.token`
  (#200), so only sessions whose project or pool names a Linear
  identity covering the team can act, and others get `no identity
  covers ENG`.
- Status sync stays as Gas City had it: prompts comment the PR link
  and move the issue to Done on merge with these commands. No
  issue-to-ticket sync.

Not yet: pack prompts that use it (#232) and Linear comments in
subscriptions (#233). Tests: `pkg/linear` against recorded API
responses, `TestRunMetaLinear`, workflow config cases. Live check:
`docs/verify/linear-tracker.md`.
