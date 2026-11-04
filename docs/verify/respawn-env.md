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

## Agent tool shell after a host restart (#253)

The process environ can be right while the agent's tool shell is not:
after a resume jcode runs tools as `bash -lc`, and NixOS's
`/etc/profile` used to reset PATH there. Restart a live session's host
(`jobshop session restart <id>`, or wait for a deploy refresh), then on
its next turn have the agent run, from its own shell tool:

```sh
command -v jobshop gh; echo "$__NIXOS_SET_ENVIRONMENT_DONE"
```

Pass: both paths print (jobshop from the deployed `bin`) and the marker
is `1`. Fail: `command not found`; check that `spec.json` `env` has
`__NIXOS_SET_ENVIRONMENT_DONE=1`.
