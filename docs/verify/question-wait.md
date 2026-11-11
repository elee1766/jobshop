# Question wait

2026-11-11 (JOB-202). Run against an isolated controller (see the
isolated-controller skill), never the live one.

1. As a session holding a step: `jobshop report waiting "sign-off" -q "Approve?" --context - --wait`
   with two lines of context on stdin. Expect the step released and
   `jobshop wait list` showing `question` on the new question id.
2. `jobshop wait list` still shows it after a minute. Answer it:
   `jobshop answer <question id> "approved"`.
3. Within a second `jobshop wait list` no longer lists the step, the step
   has a comment `Wait (...) ended: condition met: question <id> closed (answered)`
   and it is claimed by the session again (resume). The answer is a
   comment on the step.
4. Repeat with `jobshop wait add --any 'question:id=<q>' 'github.pr:pr=<repo>#<n>,state=head_changed' --resume -r x`;
   answer only the question: the wait clears.
5. Close a parked-on question with `jobshop close <q> -r dismissed`: the
   wait clears too.
6. `jobshop wait add 'question:id=<a step id>' -r x` is refused (not a
   question), as is a question of another project.
