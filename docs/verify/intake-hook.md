# Intake hook (#288)

On gasserver, after a deploy with #288, with a Linear identity and a
`tracker: linear` workflow (here `eng-pr-work`) configured:

1. Add to jobshop.yaml and `jobshop reload`:
   ```yaml
   hooks:
     - on: intake
       command: echo '{"decision":"deny","reason":"verify"}'
   ```
2. `jobshop run start eng-pr-work --var issue=<KEY>` fails with
   `intake hook deny for <KEY>: verify`, and `jobshop run list` shows no new run.
3. The Linear issue has the `needs-human` label and `jobshop attention`
   lists `Intake refused for <KEY>`. Close that ticket with resolution
   `rejected` and remove the label on Linear.
4. Gas City classifier: set `command` to a wrapper that pipes stdin (the
   event JSON, text in `.extra.text`) to the classifier and prints
   `{"decision": "allow"|"deny"|"ask_human", "reason": "..."}`. Empty
   output with exit 0 is allow; a non-zero exit or bad output is deny.
   Start a run for a harmless issue and check it starts.
5. Remove the verify hook and reload.
