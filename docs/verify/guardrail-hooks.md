# Guardrail hooks (#188)

On gasserver, after a deploy with #188:

1. Add to jobshop.yaml and `jobshop reload`:
   ```yaml
   hooks:
     - on: action
       actions: [verify-hook]
       command: echo '{"decision":"deny","reason":"verify"}'
   ```
2. `jobshop guard verify-hook; echo $?` from the admin socket prints
   `"decision": "deny"` and exits 1. `jobshop guard push` prints
   allow and exits 0 (the hook is limited to `verify-hook`).
3. `journalctl -u jobshop | grep 'hook decision'` shows the deny.
4. Remove the hook and reload.
