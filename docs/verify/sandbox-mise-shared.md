# Shared mise installs in sandboxes (#316)

With `sandbox.mise.tools = [ "node@24" ]` deployed:

1. `systemctl status jobshop-mise.service` succeeded;
   `ls /var/lib/jobshop-mise/installs/node` shows a 24.x dir.
2. In a sandboxed session: `echo $MISE_SHARED_INSTALL_DIRS` prints
   `/var/lib/jobshop-mise/installs`; `mise exec node@24 -- node -v` runs
   without a download; `touch /var/lib/jobshop-mise/x` fails (read-only).
3. Concurrency: in two sandboxed sessions at once run
   `mise install node@22` (not shared). Both finish, one waits on the
   other, and `node -v` works in both.
