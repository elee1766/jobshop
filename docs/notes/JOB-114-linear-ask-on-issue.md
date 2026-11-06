# Linear asks are opt-in (JOB-114)

2026-11-06.

`linear.ask_on_issue` (default false) controls the #293 behavior: open
worker questions and run escalations on a run with `meta.linear.issue`
are posted as a comment on the Linear issue, a question adds `Blocked`,
and a reply under the ask comment answers the question.

Before JOB-114 this was always on. An instance that wants it now opts in:

```yaml
linear:
  ask_on_issue: true
```

Off (the default), questions and escalations stay on Desk. Nothing is
posted or labelled, and a Linear reply under an earlier ask comment is
delivered as an ordinary comment instead of answering the question.
Answering a question that was posted while the switch was on still
removes its `Blocked` label, so turning it off leaves no stale labels
behind. The setting is global. A config reload applies it.
