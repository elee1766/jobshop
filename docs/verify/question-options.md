# Question with context and options (JOB-101)

On gasserver after deploying JOB-101:

1. In a session holding a scratch ticket, run
   `jobshop report waiting "base?" -q "Which base branch?" --context
   $'Rebased onto both.\nmain has the API, release does not.' --option
   "main|ship next minor" --option "release|backport the API"
   --recommend 1`. `-q` without `--context` fails before any call.
2. `jobshop attention` lists the question with `context` (your text, then
   the ticket, the asking session and recent comments) and both options.
3. In Desk Attention the row has a Context disclosure, two option
   buttons (main highlighted) and Discuss. Discuss opens the chat.
4. Click `release` and confirm. The question closes `answered`; the
   session's next turn shows `Chose option 2: release` and the
   description.

Record the ticket ids on JOB-101.
