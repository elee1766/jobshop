# GitHub personal access tokens (#341)

2026-11-04. People can run jobshop as their own GitHub user, not only
through the elee-agent App.

- Identity kind `github-token`: `token_file` (the PAT), `repos` globs,
  optional `api_url`. No installations. App and Linear fields are a
  load error.
- The token is read from the file on each use and passed through as is
  (`identity token`, git credential helper, activity poller).
- `identity check` reads `GET /user` and reports the login as `viewer`.
- Subscriptions drop comments by the login of any `github-token`
  identity covering the thread's repo, as they drop `[bot]` logins for
  the App. The login is cached until a reload changes identities.
- Linear personal API keys were already kind `linear`; not changed here.
