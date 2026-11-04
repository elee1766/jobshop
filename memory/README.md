# Memory and skills

This directory is injected into Jobshop agent prompts (#266) when a
host sets the selfdev or operator pack var `memory_dir` to this
directory in a checkout of elee1766/jobshop. Layout:

- `memory/global.md`: every session.
- `memory/repo/<owner>__<name>.md`: turns on a ticket for that repo.
- `memory/pool/<name>.md`: sessions of that pool.
- `skills/<name>/SKILL.md`: agentskills.io skills (frontmatter `name`,
  `description`); prompts get a one-line index and read the file on demand.

Each memory file is capped at 2200 characters; longer files are cut at
injection. Run `jobshop memory check <dir>` before committing.

This repo is public. No secrets, tokens, hostnames that are not already
public, or personal data here.
