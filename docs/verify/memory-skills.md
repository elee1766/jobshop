# Memory and skills

Check that a session prompt carries the global memory (#266).

1. In the host `jobshop.yaml` (or the Nix `settings.packs` vars), set
   `memory_dir: /var/lib/jobshop/src/jobshop/memory` on the selfdev
   entry, then `jobshop reload`.
2. `JOBSHOP_SOCKET= jobshop memory check` prints nothing and exits 0.
3. `JOBSHOP_SOCKET= jobshop session restart <dev session>`, then
   `JOBSHOP_SOCKET= jobshop session prompt <dev session>` shows
   `=== rule:selfdev-memory ===` with `Memory (global,
   memory/global.md):` and the text of `memory/memory/global.md`.
4. For a ticket on elee1766/jobshop with a file at
   `memory/memory/repo/elee1766__jobshop.md`, the same prompt shows
   `Memory (repo, ...)`.
5. Append over 2200 characters to a scratch copy's `global.md` and run
   `jobshop memory check <copy>`: it lists the file and exits 1.
