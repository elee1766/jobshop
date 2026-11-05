# Memory and skills

Check that a session prompt carries the local memory (#266, #592).

1. The selfdev and operator packs read `<root>/memory` by default
   (`/var/lib/jobshop/memory` on the NixOS module). `ls
   /var/lib/jobshop/memory` exists and is owned by the service user.
2. `jobshop memory show` prints that directory and its files;
   `JOBSHOP_SOCKET= jobshop memory check` prints nothing and exits 0.
3. `jobshop memory write global --from-file <f>` with a one-line file,
   then `JOBSHOP_SOCKET= jobshop session restart <dev session>` and
   `JOBSHOP_SOCKET= jobshop session prompt <dev session>` shows
   `=== rule:selfdev-memory ===` with `Memory (global,
   memory/global.md):` and that line. `jobshop memory log` lists the edit.
4. For a ticket on elee1766/jobshop, `jobshop memory write
   repo/elee1766__jobshop ...` makes the same prompt show
   `Memory (repo, ...)`, and `=== rule:selfdev-guide ===` appears.
5. A write over 2200 characters fails with `over the memory cap` and
   leaves the file as it was.
6. Curator: `jobshop pool scale learn auto`; after one curate run,
   `jobshop memory log` shows its writes and `gh pr list --search Curate`
   shows no new PR.
