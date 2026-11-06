# Turning off Linear asks (JOB-114)

2026-11-06.

`linear.ask_on_issue` (default true) controls the #293 behavior: open
worker questions and run escalations on a run with `meta.linear.issue`
are posted as a comment on the Linear issue, a question adds `Blocked`,
and a reply under the ask comment answers the question.

```yaml
linear:
  ask_on_issue: false
```

With it off, questions and escalations stay on Desk. Nothing is posted
or labelled, and a Linear reply under an earlier ask comment is
delivered as an ordinary comment instead of answering the question.
Answering a question that was posted while the switch was on still
removes its `Blocked` label, so turning the switch off leaves no stale
labels behind. The setting is global. A config reload applies it.
