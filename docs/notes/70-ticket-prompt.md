# Ticket prompt (#70)

The claim prompt is the pool prompt, the ticket id, title and body, then
`# Recent comments` when the ticket has any: the last 3 as
`[time] author:` and the body, each body capped at 2000 bytes with
`[truncated]`. A line tells the agent to address a refinery finding, push
and hand back to refinery. A ticket without comments gets the same
prompt as before.
