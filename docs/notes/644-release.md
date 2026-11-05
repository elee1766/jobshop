# Releases with release-please (#644)

2026-11-05. Operator decision: semver 0.x.y from conventional commits,
never 1.0.0.

- PR titles are conventional commits (#646) and become the squash
  commit subject. `.github/workflows/release.yml` runs release-please on
  every push to master: it keeps one release PR open with the next
  version, `CHANGELOG.md` and `version.txt`. Ordinary merges only update
  that PR. A release happens when a human merges it: release-please tags
  `v<version>` and creates the GitHub Release.
- Version rule (`release-please-config.json`): `bump-minor-pre-major`
  makes a breaking change (`!` or `BREAKING CHANGE:`) bump the minor,
  `bump-patch-for-minor-pre-major` makes `feat` bump the patch. Never
  set `release-as: 1.0.0`. The first release is 0.1.0 through
  `release-as`; remove that line once it is out.
- On a release the workflow builds `jobshop_<v>_<os>_<arch>.tar.gz` for
  linux amd64 and arm64 (plain `go build`, `CGO_ENABLED=0`; the sandbox
  is Linux only),
  uploads them with `checksums.txt`, and checks `nix build` at the tag
  reports the version.
- Nix: `nix/package.nix` reads `version.txt`, so
  `nix build github:elee1766/jobshop/v0.1.0` is `jobshop-0.1.0`. Builds
  between releases report the last release; the commit in
  `jobshop version` tells them apart. Self-deploy matches on the commit,
  so it is unaffected.
- release-please uses an elee-agent App token (secrets `RELEASE_APP_ID`,
  `RELEASE_APP_KEY`): a PR opened with `GITHUB_TOKEN` gets no CI runs,
  so the required checks would never report on the release PR.
