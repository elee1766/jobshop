# Reserved human slot

On a pool with `max_sessions: N` and N sessions open working tickets:

```sh
JOBSHOP_SOCKET= jobshop session new --agent jcode --pool <pool>   # starts: the human slot
JOBSHOP_SOCKET= jobshop session new --agent jcode --pool <pool>   # refused: max_sessions N plus one reserved human slot
```

- `session list` shows N+1 open sessions in the pool; the supervisor
  starts no more while they are open.
