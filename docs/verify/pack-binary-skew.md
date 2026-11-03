# Live pack newer than the binary (#227)

1. `ls /var/lib/jobshop/config.last-good.yaml` exists after a start.
2. Sync guard with the controller down: `systemctl stop jobshop;
   systemctl start jobshop-workspace-jobshop-sync; git -C <checkout>
   rev-parse HEAD` is the installed `jobshop version` commit, not
   origin/master. Then `systemctl start jobshop`.
3. Simulate a pack the binary cannot load: in the checkout, add
   `${pack_unknown}` to a prompt in packs/github/prompts/dev.md, then
   `systemctl restart jobshop`. `systemctl is-active jobshop` stays
   active, the journal has "serving the last good config", and
   `jobshop attention` lists a `config` item. Revert the edit and
   `jobshop reload`; the item is gone.
4. After a manual switch to an older jobshop, the next
   `jobshop-self-deploy` run logs "switched outside self-deploy".
