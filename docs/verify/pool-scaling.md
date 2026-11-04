# Pool scaling live check (#281)

On gasserver after the deploy that ships #281.

1. With dev busy (8 working): `jobshop pool scale dev 2`. Watch
   `jobshop pool list` and `jobshop session list`: 6 sessions show
   `cordon: scale`, each finishes its ticket and closes `scaled_down`.
   No ticket is released with `session crashed` and no turn is cut.
2. `jobshop session cordon <id>` on an idle dev session: it claims no new
   ticket (`session list` shows `cordon: manual`) and stays open past
   idle_timeout. `jobshop session uncordon <id>` returns it.
3. With a target and a cordon set, `systemctl restart jobshop`. After the
   restart `pool list` still shows the target and `session list` the
   cordon. Then `jobshop pool scale dev auto`.
