# Compact GitHub PR and Linear issue links in Desk

2026-11-08. JOB-125.

Desk's shared Markdown renderer (`ui/kit/markdown.js`, used by chat,
session transcripts, ticket status, bodies, comments and history) shows
bare URLs and `<url>` autolinks for these two forms with short labels:

- `https://github.com/<owner>/<repo>/pull/<n>` as `PR:<owner>/<repo>#<n>`
- `https://linear.app/<workspace>/issue/<KEY-n>[/<slug>]` as `<workspace>@<KEY-n>`

The href stays as written, query and fragment included. Authored
`[label](url)` links keep their label. Code spans and blocks are never
linked. Only `https` on the exact hosts `github.com` and `linear.app`
counts: other hosts, ports, credentials and other paths keep the full URL.
`ui/kit/links.test.js` covers the matching (`node --test`, run in CI).
