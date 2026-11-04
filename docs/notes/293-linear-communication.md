# Linear issue communication (#293)

2026-11-04. Part of #258.

For a run whose tickets carry `meta.linear.issue`:

- Each tick the controller posts open needs-human questions (`jobshop
  report waiting --question`) and `Escalate run` tasks to the Linear
  issue with the `linear ask` mutation (#280). The body starts with the
  profile URL of `meta.linear.human`, which run start copies from pack
  var `linear_human`; Linear turns that into a notifying mention.
- The write key is `linear-ask/<ticket id>` in `write_keys`, and the
  comment id is derived from it, so a retry after a crash replays the
  stored result or finds the comment. The posted comment is stored in
  `meta.linear_ask {comment, url, issue}` on the question or escalation.
- Questions also add the `Blocked` label. The pack's `linear_skip_labels`
  now defaults to `Blocked`, so intake skips the issue while it waits.
  The issue state is not changed.
- Linear subscriptions (#233) read each comment's parent. A human reply
  whose parent is an open question's ask comment passes the intake gate
  and then goes to `answerQuestion`, from `linear:<name>`. Nothing else
  goes there: other comments, a second reply, and replies to escalations
  are delivered as before. Own and bot comments are dropped as before.
- Any answer (Linear, gateway, `jobshop answer`) removes `Blocked` once
  no other open question on that issue has been posted.
- Progress is not posted to Linear.
