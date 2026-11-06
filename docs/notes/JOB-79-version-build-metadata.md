# Version build metadata (JOB-79)

2026-11-06. `version.txt` only changes in the release-please commit, so a
self-deployed master commit after a release reported that release
(e7083d2 said `0.4.1` and shipped in 0.4.2).

Built: `nix/package.nix` sets the binary version to
`<version.txt>+<rev>` (semver build metadata), e.g.
`jobshop 0.4.4+e7083d2 (e7083d2)`. `rev` is the flake's `shortRev` or
`dirtyShortRev`; with no git info (`unknown`) it stays plain
`version.txt`. The derivation `version` is unchanged, so store paths and
`pname-version` do not move per commit. The release.yml tarballs keep the
plain tag version.

Consumers: the module's workspace sync, rollback and health checks and
the selfdev `deploy_check` read the commit in parentheses, not the
version. release.yml's Nix check matches `jobshop <version>` followed by
`+` or a space.
