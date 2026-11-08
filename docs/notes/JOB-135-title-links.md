# Markdown links in Desk titles (JOB-135)

2026-11-08. Ticket and run titles render inline Markdown links
(`[label](url)`) as real links. Nothing else in a title is parsed: emphasis,
code, headings and raw HTML stay literal, so plain titles look as before.

- `ui/kit/titles.js` parses a title into text and link parts (no HTML is
  built). Destinations follow `md()`: http(s) opens in a new tab with
  `noopener noreferrer`, `jobshop://ticket|session/<id>` becomes a Desk
  trace link, any other scheme (`javascript:`, `data:`, unknown
  `jobshop://`) keeps the label and loses the link.
- `parseTitle(title).text` is the readable form (labels, no URLs). It feeds
  tooltips, accessible names and the In flight filter. Trace search is
  server side and still matches the stored title.
- Used in the waiting list, In flight table, trace search results, graph
  nodes, the ticket heading and relationship lists.
- A link cannot sit in a button or another link. A title with links keeps
  its action as an empty `.hit` control stretched behind the text
  (`TitleAction.svelte`, `.hosted` in app.css): one tab stop named by the
  plain text, a click on text runs the action, a click on a title link opens
  only the link. A title without links renders the same button or link as
  before.
- Titles stay as stored: ids and workflow dedup keys are untouched.
- Checked on an isolated controller (desktop and Pixel 7, light and dark):
  GitHub and Linear links, two adjacent links, plain, unsafe and long
  titles, link vs text clicks, keyboard Enter on the action, no nested
  interactive elements, no overflow. `ui/e2e/title-links.spec.js` (needs
  `DESK_MUTATE=1`) covers it, `ui/kit/titles.test.js` the parser.
