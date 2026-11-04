# Intake hook (#288)

2026-11-04. Hook point `on: intake` (part of #258, untrusted text). Linear
issue text is checked by the configured hooks before it reaches a prompt:

- Run start for a `tracker: linear` workflow (run start, triggers,
  batches): the controller reads the issue with the team's identity and
  sends title, description and non-self comments.
- Subscription delivery of a Linear comment: each comment is checked on
  its own. A refused one is dropped and the read position still moves on.

Event on stdin: `{"on":"intake","target":"ENG-5","project":...,"extra":
{"source":"run_start"|"comment","issue":"ENG-5","author":...,"text":...}}`.

Any answer but allow (deny, ask_human, hook error, timeout, an unreadable
issue) refuses: the run start fails with forbidden, the Linear issue gets
the `needs-human` label, and one `needs-human` ticket per issue (or
comment) is filed with meta.hook.key, so it shows on `jobshop attention`.
ask_human also labels it `question`. Closing that ticket `done` approves:
the next start for the issue skips the hook. No intake hooks means allow,
and the issue is not read.
