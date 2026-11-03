# NixOS module uses jobshop identities (#216)

2026-11-03. `services.jobshop.identities.<name>` is written to
`settings.identities` (key paths only; the files come from host secrets).
`gitIdentity.jobshopIdentities = true` replaces `credentialHelper`:

- git: `credential.https://github.com.helper` is
  `!<jobshop> identity git-credential` with `useHttpPath=true`, so each
  repo gets the identity that covers it.
- gh: a wrapper first on PATH sets `GH_TOKEN` from
  `jobshop identity token <repo>`, where repo is `--repo`/`-R` or the
  cwd's origin. A preset `GH_TOKEN` wins; no repo runs gh without one.

Setting both options fails an assertion. agent-credd can be removed from
the host once `docs/verify/module-identities.md` passes.
