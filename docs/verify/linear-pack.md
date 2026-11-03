# Linear pack prompts (#232)

Needs the Linear identity from `docs/verify/linear-tracker.md` on a
throwaway team (key `<KEY>`) and a throwaway GitHub repo.

1. Add a packs/github entry with `prefix: lin-` and vars `tracker:
   linear`, `linear_workspace`, `linear_team: <KEY>`, `linear_target:
   <KEY>-1`, plus the usual repo vars; `jobshop reload`.
2. In Linear, file `<KEY>-2` (a one-line README change) and name it in
   the description of `<KEY>-1`.
3. Wait for a `lin-plan` ticket (or create one in `lin-plan`). Expect a
   `lin-pr-work` run with key `issue-<KEY>-2` in `jobshop run list`.
4. Expect on `<KEY>-2`: state In Progress, then a `PR <url>` comment;
   the PR title ends `(<KEY>-2)`.
5. After merge and post-merge: a `Shipped in <sha>` comment and state
   Done.
