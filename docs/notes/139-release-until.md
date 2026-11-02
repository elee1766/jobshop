# Deferred release (#139)

2026-11-02. `jobshop release --until <dur|RFC3339> -r "<why>"` hands a
held ticket back with `defer_until` set, so `ready` and `claim next` skip
it until then. A deferred release is not a step failure: `fields.failures`
is unchanged and the controller comment reads "Released until <time>:
<why>". `until` must be in the future. A plain release still counts as a
failed attempt (#90).

Why: #56 had a step that could only run after a fixed time. Each release
reopened it at once, a new session claimed it within seconds, and the
three retries were used up in under a minute.
