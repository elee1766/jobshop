# README for humans, llm/ for agents (#551)

2026-11-05.

- README.md is short: what Jobshop is, goals, concepts, example flows,
  quick start, a docs table. Its second line points agents to
  `llm/README.md`.
- The setup reference moved with light edits: concepts, run, own-user
  tokens and `*_file` forms to `docs/running.md`; isolation to
  `docs/isolation.md`; webhooks and generic ingest to `docs/ingest.md`;
  NixOS to `docs/nixos.md`; the text-file rule into `docs/packs.md`.
  The Develop bullets are already in AGENTS.md.
- `llm/cli.md` is the help of every visible command, generated from the
  kong tree. `TestCLIDoc` fails when it is stale; regenerate with
  `go generate ./pkg/cli`.
- `TestDocLinks` checks relative links in README.md, `llm/*.md` and
  `docs/*.md` resolve to files (not anchors).
- `jobshop docs llm` prints the link to `llm/README.md`.
