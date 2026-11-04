# Model config live check (#304)

Added 2026-11-04. Needs gasserver with jcode >= v0.85 (ACP
`session/set_config_option`).

1. Note a live `acphost-dev` session and its model:

   ```sh
   JOBSHOP_SOCKET= jobshop session list | jq -r '.[] | select(.ticket.pool == "acphost-dev") | "\(.id) \(.model) \(.effort) \(.host_pid)"'
   ```

2. Set `model: <route>` on the `acphost-dev` pool and
   `JOBSHOP_SOCKET= jobshop config reload`. After the session's current
   turn ends, `session list` shows the new model with the same
   `host_pid` (no respawn), and the journal has the change:

   ```sh
   JOBSHOP_SOCKET= jobshop session attach <id> --json --no-follow | grep config_option_update
   ```

   The next turn still knows the earlier conversation (ask it in
   `session attach --rw` on a chat pool, or read the next turn's reply).

3. Move one session to another provider:

   ```sh
   JOBSHOP_SOCKET= jobshop session config <id> model=openai-oauth:<model>
   ```

   Pass: `session list` shows the route and `override`, the next turn
   runs (journal shows a reply), and `session config <id> --reset` goes
   back to the pool's model.

Post the session id, models and journal seqs on #304.
