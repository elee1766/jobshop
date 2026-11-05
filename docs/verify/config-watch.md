# Config watch (#533)

On a host with `reload.watch: true` (or `services.jobshop.reloadOnChange`):

1. `jobshop config show | jq .watch` shows `on: true` and nonzero counts.
2. Edit a prompt file in a live pack. Within the debounce (default 2s)
   the journal logs `config reloaded (watch)` naming the file.
3. Break a `pack.yaml` (e.g. `pools: [`). The journal logs `config reload
   failed`, `jobshop attention` lists `config-invalid`, pools are unchanged.
4. Fix it. Attention clears after the next reload.
5. Open a pack file in vim and leave it: the `.swp` file does not reload.
