# Agent pool pack next to gascity (#634)

On gasserver, with isolation deployed (`docs/verify/isolation.md`) and
gascity still running. Pick one rig repo with low traffic.

1. Add an instance to jobshop.yaml (a checkout jobshop can read, not the
   gascity worktrees), with the subuid start from `/etc/subuid`:

   ```yaml
   packs:
     - path: /var/lib/jobshop/src/jobshop/packs/agent
       prefix: app-
       vars:
         agent: jcode
         checkout: /var/lib/jobshop/repos/<repo>
         repo: <owner>/<repo>
         branch: main
         primary_sessions: 1
         review_sessions: 1
         isolation:
           user: {uid: 1000, map_uid: <subuid start>, map_gid: <subgid start>, groups: [jobshop]}
           limits: {memory: 8G, pids: 4096}
         sandbox: {backend: bubblewrap, network: pasta, deny: [10.0.0.0/8]}
   ```

   `jobshop config check` then `jobshop reload`. Lower the gascity
   `routing.primary` cap for that rig by the same count so the host
   session total does not grow.
2. `jobshop run start app-pr-work --var task="<a small real task>" --key
   compare-1`. Give the same task to gascity (`gc sling
   <rig>/routing.primary`).
3. Check, for the jobshop run:
   - `jobshop session sandbox <id>` shows the mapped uid, the cgroup
     limits and pasta with the deny list.
   - `jobshop run show <run>`: deliver, review, merge close done, the PR
     merged only after an approving review on its head.
4. Compare with the gascity bead: time to PR, time to merge, tool calls
   (session journal vs gascity transcript), and holds or mails to the
   mayor. Post both on #634.

Roll back: remove the pack entry and reload; restore the gascity cap.
