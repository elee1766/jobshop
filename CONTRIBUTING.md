# Contributing

PRs are squash-merged and the PR title is the commit subject, so it must be
a [conventional commit](https://www.conventionalcommits.org/):
`<type>(<scope>): <summary> (#<issue>)`, e.g. `fix(store): retry busy writes (#123)`.

- Types: feat, fix, perf, refactor, docs, test, ci, build, chore, revert.
  Scope is the package or area (`control`, `desk`, `store`, `packs/github`).
- The summary starts lowercase. A breaking change to the CLI, config,
  protocol or store schema gets `!` after the scope or a `BREAKING CHANGE:`
  line in the body. The `pr-title` check enforces the format.

Build, test and lint rules are in [AGENTS.md](AGENTS.md).
