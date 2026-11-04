# Linear App identity (#278)

2026-11-04. Target #258, item "Linear App identity".

- Config: `kind: linear-app` with `client_id_file`, `client_secret_file`,
  `workspace`, `teams`, optional `scopes` (one per item; default `read,
  write, issues:create, comments:create`) and `token_url` (tests).
  `token_file` and GitHub App fields are rejected; `kind: linear` (a
  static key) is unchanged and rejects the app fields.
- Minting: `identity.Minter.Token` posts the OAuth `client_credentials`
  grant to `https://api.linear.app/oauth/token` and caches the token per
  identity and scope set until 10 minutes before `expires_in` runs out,
  the same `refreshBefore` and cache as github-app. Errors carry
  Linear's `error`/`error_description`, never the secret or a token.
- API calls: `pkg/linear` sends a `lin_api_` key as is and any other
  token as `Bearer <token>`, which OAuth tokens need.
- Everything that took a Linear identity (subscriptions, activity feed,
  `jobshop linear`, `ActivityTeams`) accepts both kinds through
  `config.IsLinear`.
- `jobshop identity check <TEAM>` mints and then reads `viewer` for both
  Linear kinds, so a refused token fails there; the result has `viewer`.
- nix: `services.jobshop.identities` already passes any keys through to
  settings; the option example and description now cover the secret
  file paths.
