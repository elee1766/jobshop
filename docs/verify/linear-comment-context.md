# Linear comment context live check (#595)

Needs a Linear identity for the team and subscriptions on (waits on
#258). Two Linear users A and B; give B a second account C with the
same full name as A.

1. Set `subscriptions.authors: [<A's displayName>, <B's user id>]`.
2. Start a run on a test issue so a session follows it.
3. As A, comment. As B, reply under A's comment. As C, comment.
4. Within an interval the ticket gets one comment by `linear:several`
   starting `New on Linear issue <KEY> "<title>" (<state>)`, with
   `A (Full Name)` and B lines, B's line followed by `in reply to A
   ...: "<first line>"`. C's comment is absent.
5. As A, edit an older comment. The next delivery has it once as
   `comment (edited)`; a later pass does not repeat it.
6. Record the GetIssue response for that issue (`jobshop linear api`)
   and replace `pkg/orchestration/control/testdata/linear_comments.json` if its shape
   differs.
