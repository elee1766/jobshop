# Prompt text from files (#382)

2026-11-04. One loader in pkg/config (`textfile.go`) reads every
prompt-like `*_file` field into its inline field at load:

| File field | Fills | Pack var substitution |
| --- | --- | --- |
| `pools.<n>.prompt_file` | `prompt` | yes |
| `workflows.<n>.steps[].prompt_file` | `prompt` | no (run vars at start) |
| `triggers.<n>.ticket.body_file` | `body` | yes |
| `prompt_rules[].file` | `text` | no (as before) |

- Paths are relative to the declaring file (pack.yaml or jobshop.yaml).
  Host files are always literal.
- Setting both the inline field and the file field is a load error.
- `jobshop reload` re-reads the files.
- `config show` sources list the file read under `<entry>.<field>`, for
  example `triggers.replan.ticket.body` or `prompt_rules.recall.text`.
- packs/github keeps its `recall` and `learn-nudge` rules in
  `prompts/rules/*.md`.
- The #379 escalation chat prompt, when it lands, should add one
  `t.load(..., "prompt", "prompt_file", ...)` call here.
