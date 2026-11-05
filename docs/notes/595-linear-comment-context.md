# Linear comment context (#595)

2026-11-05. Linear comments reached tickets with the full name as the
author, one author for a mixed batch, no reply context, a bare header,
and no edits.

- Comment queries (`GetIssue`, `GetContext`, `Changes`) read
  `user { id name displayName email isMe }` and `botActor { id name }`.
  Lines show `displayName (name)`. A single-author delivery's ticket
  comment author is `linear:<displayName>`.
- `subscriptions.authors`, `bots`, `ignore` and a reply wait's `from`
  match a Linear comment by user (or bot actor) id, displayName or
  email, case-insensitively. The full name still matches as a fallback
  and logs a WARN naming the entry, since two people can share it and
  a rename drops a person. GitHub matching is unchanged.
- One delivery per thread pass, as before. When it has more than one
  author (by Linear id), the ticket comment author is `linear:several`
  and each line keeps its own name. Chosen over one comment per author
  because the session gets one interrupt message per delivery and a
  thread's news reads in order.
- A reply has `in reply to <author>: "<first line, 120 chars>"` under
  its header line when the parent is in the issue's fetched comments,
  else `in reply to <parent url>`.
- Header: `New on Linear issue ENG-5 "<title>" (<state>)`. GitHub is
  unchanged: `ghSource` does not read the title, so it would cost a
  call per pass.
- Edits: a comment created before the read position and edited after
  it is delivered as `comment (edited)` at its `editedAt`. The inbox key
  is the comment id plus that time, so one edit is delivered once and
  each later edit again. A new comment edited before the pass takes the
  edit time, so it is not sent twice. Edits by self or bots are not
  delivered.
- Poller rows written before this change lack the new fields and
  deliver as before.
- Tests: `TestLinearCommentContext` (fixture
  `pkg/control/testdata/linear_comments.json`, Linear's response
  shape; no Linear workspace was reachable to record from, see
  `docs/verify/linear-comment-context.md`), `TestLinearDeliveryAuthor`.
