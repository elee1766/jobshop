# Answering questions without options (JOB-115)

2026-11-06. `pkg/control/report.go`, `ui/plugins/attention/Attention.svelte`.

- Config `questions: {require_options: true}` (default false): the
  controller refuses a `report -q` with fewer than two options
  (`--option` or `--options-json`), with an error naming the setting.
  The check is in the controller because the CLI does not read the
  config. Reports without a question are unaffected.
- Desk: a question row with no options shows a one-line answer box.
  Enter or Answer confirms, then calls `question.answer {id, text}`,
  so a short answer does not need the chat. Discuss still opens it.
- Tests: `TestReportRequireOptions`, `TestParseQuestions`, e2e
  `a question without options answers in one line (JOB-115)`.
