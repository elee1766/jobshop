# Identities on gasserver (#200)

After a deploy that includes #200, with `identities.elee-agent` set
(the existing App key file, installation for elee1766) and
`projects.js.identities: [elee-agent]`:

- `jobshop identity list` shows elee-agent with project js.
- `jobshop identity check gfx-labs/jobshop` returns `identity:
  elee-agent` and an expiry about an hour out.
- `jobshop identity check nobody/else` fails with `no identity covers`.
- From a session: `GH_TOKEN=$(jobshop identity token gfx-labs/jobshop)
  gh api repos/gfx-labs/jobshop -q .full_name` prints the repo, and
  `git -c credential.helper= -c credential.helper='!jobshop identity
  git-credential' -c credential.useHttpPath=true ls-remote
  https://github.com/gfx-labs/jobshop HEAD` succeeds.
- Second identity: add another installation or App covering a second
  repo, `config reload`, and `identity check` on that repo names it.
