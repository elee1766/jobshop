# #261: config check before self-deploy switches (2026-11-04)

- `jobshop config check [--file F] [--replace OLD=NEW]` loads a config
  and its packs offline and exits 1 on error. `--replace` reads packs
  whose path starts with OLD from NEW.
- Self-deploy fetches the new rev's source, builds its package, and runs
  that binary's `config check` on the host's jobshop.yaml with the
  selfDeploy workspace dir replaced by the new source. On failure it
  does not switch: the rev goes to `failed`, the error to
  `failed.reason`, and `jobshop attention` shows "self-deploy skipped
  <rev>, config check failed: <error>". The rev is not retried; a newer
  green commit or a host config change plus a newer commit moves on.
- pack.yaml `renamed_vars: {old: new}`: a host still setting the old name
  gets the value under the new name and a logged warning (controller log,
  `config check` stderr). The operator pack maps `repo` to `jobshop_repo`.
- The refinery's merge prompt asks a PR that removes or renames a pack
  var for `renamed_vars` or a list of hosts that set it.
