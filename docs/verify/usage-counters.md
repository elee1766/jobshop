# Verify: usage counters in shipped journals

After a deploy containing JOB-160 (`jobshop version --server`), run any
prompt in a session. Query VictoriaLogs for the newest row of that
session whose message has `"usage"`: `inputTokens`, `outputTokens`,
`totalTokens`, `cachedReadTokens` and `cachedWriteTokens` must be
numbers, and the pool, project, session, ticket, run, workflow and step
labels must be present. Rows shipped before the deploy stay redacted.
