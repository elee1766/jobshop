# Session env after a self-deploy respawn (#244)

After a deploy that refreshes sessions (`refreshing host env` in the
journal), take a respawned session's new host pid from
`jobshop session list` and check:

```sh
tr '\0' '\n' < /proc/<host_pid>/environ | grep -E '^(PATH|GH_TOKEN)='
```

Pass:
- PATH starts with the deployed jobshop `bin` and contains gh (or the
  identities gh wrapper).
- Inside the session, `command -v gh jobshop` finds both and
  `gh issue view <n>` works (GH_TOKEN or the wrapper's identity token).
- A session-filed `jobshop ticket create` without `--pool` shows the
  session's pool in `jobshop ticket show`.
