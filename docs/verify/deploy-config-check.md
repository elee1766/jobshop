# Deploy config check

Added 2026-11-04 (#261). Checks that a pack var removed on master, with
the host config not updated, stops self-deploy before the switch.

On gasserver, as root, with a branch whose `packs/operator/pack.yaml`
references a var the host does not set (e.g. `${nope}` in a pool field)
and green CI on it, point a manual run at it:

```sh
src=$(nix flake metadata --json "git+https://github.com/elee1766/jobshop?rev=<sha>" | jq -r .path)
pkg=$(nix build --no-link --print-out-paths "path:$src#jobshop")
runuser -u jobshop -- $pkg/bin/jobshop config check \
  --file /var/lib/jobshop/jobshop.yaml --replace /var/lib/jobshop/src/jobshop=$src
```

Expect exit 1 and `var "nope" is not set` naming the pack file. With
master instead of the branch, expect `ok: N pools, ...`.

On the next real self-deploy after this merges, `journalctl -u
jobshop-self-deploy` shows `config check: ok: ...` before `rebuild`.
If a check ever fails there, expect no switch, `failed` and
`failed.reason` in /var/lib/jobshop-deploy, and `jobshop attention`
listing one ticket "self-deploy skipped <rev>: config check failed"
(`deploy:config-check`, #577), closed by the next good deploy.

Rename shim: with the host setting `repo` on the operator pack entry,
`config check` prints a `warning: ... var "repo" is renamed to
"jobshop_repo"` line and succeeds.
