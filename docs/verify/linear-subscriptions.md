# Linear comments reach the working session (#233)

Needs `subscriptions:` configured and a Linear identity covering the team.

1. Start a `tracker: linear` run for a test issue, e.g. ENG-1, and wait
   until its implement step is claimed.
2. `jobshop ticket show <step>`: `meta.subscribe.threads` has
   `{"linear":"ENG-1", ...}`.
3. As a human, comment on ENG-1 in Linear. Within one interval the step
   has a comment by `linear:<your name>` and the session got an
   interrupt message "Linear ENG-1 has new comments".
4. Have the session run `jobshop linear comment ENG-1 hi`. It is not
   delivered back (own comment), and `seen` advances.
