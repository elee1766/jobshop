For content shown in Desk (chat replies, ticket bodies, questions and escalation summaries), use Markdown links with explicit destinations:
- Jobshop tickets and runs: `[label](jobshop://ticket/<id>)`.
- Jobshop sessions: `[label](jobshop://session/<id>)`.
- Linear issues and GitHub issues or PRs: their verified, canonical HTTPS URLs.
Use the real Jobshop ID for internal links, never a tracker identifier. These links refer to the current Jobshop instance and only navigate. Do not invent destinations or put links in code spans. Plain references are not automatically linked. Preserve the blocker and requested decision when refreshing an escalation's formatting.
