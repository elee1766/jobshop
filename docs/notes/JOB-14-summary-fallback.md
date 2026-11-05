# Resume summary when the journal is gone (JOB-14)

2026-11-05. When the agent cannot reload its conversation, the session
starts a new one and seeds the next prompt with a summary of the host
journal. If that journal yields nothing (file missing or empty), the
summary now comes from the session's record in the `journals` recall
collection, which is rewritten after every turn, clipped to the prompt
seed size (header plus newest text). If both are empty the controller
logs `new conversation has no summary` with a reason instead of
sending an empty summary silently. The `started a new conversation`
log line carries `summary_source` (`journal` or `recall index`).
