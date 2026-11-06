# Stall trigger max_open (JOB-93)

2026-11-06. A stall trigger starts at most `max_open` runs that are open
at once, across all its keys (default 3, was a hard-coded constant).
Top-level next to `max_runs` and `cooldown`, stall triggers only:

```yaml
triggers:
  review-requests:
    when: {stall: [{signal: review_owed, check: "..."}]}
    run: {workflow: pr-review, vars: {pr: "{{.stall_subject}}"}}
    max_open: 8
```

Raise it when a few slow runs hold back a backlog of independent stalls
(jobcity had 7 owed reviews behind 3 stuck runs). `trigger list` shows
`N stall runs open (max_open M)` while the cap holds.
