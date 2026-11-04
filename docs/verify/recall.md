# Recall live check (#265)

After the self-deploy that ships #265, on gasserver:

1. `jobshop recall vendorHash` as admin returns comment hits from tickets
   written before the deploy (backfill worked).
2. `jobshop comment "recall-check $(date +%s)"` on any ticket, then
   `jobshop recall recall-check` returns it without a restart.
3. After a dev session finishes a turn, `jobshop recall --scope journal
   <a word from its ticket>` returns a hit with `ref` = that session id.
4. A dev session's system prompt includes the `recall` rule (the
   `prompt_rules` list in `jobshop config show`).

Post the output of 1 to 3 on #265.
