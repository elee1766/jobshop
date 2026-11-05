# Verify: journal compression (#631)

On gasserver after deploying a build with #631:

1. Before: `du -sh /var/lib/jobshop/sessions` (about 1.6 GB on
   2026-11-05).
2. Set `services.jobshop.journal.compress.after = "24h";` (level 3 is
   the default) in the host config repo and switch.
3. `journalctl -u jobshop | grep 'journal compressed'` shows one line per
   old session, with `bytes_before` and `bytes_after`. The first sweep
   runs right after start.
4. After: `du -sh /var/lib/jobshop/sessions` is well below the starting
   size. `ls /var/lib/jobshop/sessions/<old id>/` shows
   `journal.ndjson.zst` and `journal.idx`, and no `journal.ndjson`.
5. In Desk, open a session closed more than a day ago. The transcript
   loads in under 500ms (browser network tab, the session.subscribe
   tail). `jobshop session attach <old id> --tail 20` prints its last
   frames.
6. `jobshop recall <a word from that session>` still finds it.
