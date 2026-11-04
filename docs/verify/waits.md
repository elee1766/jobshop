# Waits (#381)

On gasserver after deploying #381:

1. In a session holding a scratch step, run `jobshop wait add
   github.review:pr=elee1766/jobshop#<open PR>,login=<reviewer>
   github.checks:pr=elee1766/jobshop#<same>,state=success,required=true
   -r "scratch approval" --owner <reviewer> --resume --timeout 2h`.
   The step is open, unclaimed, `fields.failures` unchanged, and absent
   from `jobshop ready`. `jobshop wait list` shows it with the head.
2. Approve the PR on that head with checks green. Within a minute (or
   one webhook) the step has a "Wait (...) ended: condition met"
   comment, no `fields.wait`, and is pre-assigned to the session.
3. Repeat step 1 on a PR, then push a new commit: the wait clears with
   "invalidated: ... head moved".
4. Add a wait with `--timeout 1m` on a condition that does not hold.
   After a minute `jobshop attention` lists it as kind `wait` with the
   reason and owner. `jobshop wait clear <id>` removes it and the label.

Record the ticket ids on #381.
