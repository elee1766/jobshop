# Identities (#200)

2026-11-03. Jobshop held one GitHub identity: agent-credd minted
elee-agent App tokens for elee1766/jobshop outside jobshop. Phase 7
needs several Apps and Linear workspaces, mapped to what they may touch.

Config:

```yaml
identities:
  elee-agent:
    kind: github-app
    app_id: 123
    private_key_file: /run/secrets/elee-agent.pem
    installations: {elee1766: 456}   # owner -> installation id
    repos: ["elee1766/*"]            # path.Match globs
  acme-linear:
    kind: linear
    token_file: /run/secrets/linear
    workspace: acme
    teams: [ENG]                     # "*" for all
projects:
  js: {identities: [elee-agent]}
pools:
  triage: {identities: [acme-linear]}  # on top of the project's
```

Secrets are only files; yaml rejects inline tokens. Config errors name
unknown kinds, mixed GitHub/Linear fields and unknown identity names.

Built:
- `pkg/identity` signs the App JWT (RS256, stdlib) and mints installation
  tokens, cached per identity and owner and replaced 10m before expiry.
  A reload that changes `identities` drops the cache.
- RPC `identity.token {target}`: a session gets a token for an owner/repo
  or Linear team key through the first identity (by name) its project or
  pool lists that covers it. No match is `forbidden` with a message that
  names the target and the allowed identities, so the agent escalates.
- `jobshop identity token <target>` prints the token (for `gh` via
  `GH_TOKEN=$(jobshop identity token owner/repo)`, or `LINEAR_API_KEY`).
- `jobshop identity git-credential get` is a git credential helper; set
  `credential.https://github.com.useHttpPath=true` so git sends the repo
  path and two repos can use two identities in one session.
- `jobshop identity list` and `identity check <target>` (admin). Check
  mints a token, so a bad key or installation id fails there.

Not yet: the NixOS module still points git at agent-credd; switching
gasserver is the live check in `docs/verify/identities.md`.
