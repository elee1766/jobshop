Step `fail` of a `drill` run for issue #${issue}.

This is an escalation drill. It checks that a step at its retry limit
puts the run in `needs_human` and files an escalation ticket. Do nothing
else: do not read or change the repo and do not comment on the issue.

Run exactly these two commands, then end your turn:

    jobshop comment "escalation drill"
    jobshop close -r failed
