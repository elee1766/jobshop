// Inline Markdown links in ticket and run titles (JOB-135).
//
// Only `[label](url)` is parsed. Everything else, including emphasis, code
// and raw HTML, stays literal text, so a plain title renders unchanged. The
// result is data (text and link parts), never HTML: Svelte escapes it, and
// every link destination passed resolveHref.

const LINK = /\[([^[\]]+)\]\(\s*(?:<([^<>\s]+)>|([^\s()<>]+))\s*(?:"[^"]*"|'[^']*')?\s*\)/g;

/**
 * The safe destination for a link, or null. Same rules as md(): http(s)
 * opens in a new tab, jobshop://ticket|session/<id> becomes a Desk trace
 * link, anything else is not a link.
 * @param {string} href
 * @returns {{ href: string, external: boolean } | null}
 */
export function resolveHref(href) {
  const internal = /^jobshop:\/\/(ticket|session)\/([^/?#]+)$/i.exec(href);
  if (internal) {
    try {
      const id = decodeURIComponent(internal[2]);
      if (/^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(id)) return { href: '?view=trace&ref=' + encodeURIComponent(id), external: false };
    } catch { /* malformed identifier */ }
  }
  try {
    if (['http:', 'https:'].includes(new URL(href).protocol)) return { href, external: true };
  } catch { /* relative or bad */ }
  return null;
}

/**
 * @typedef {{ text: string, link?: { href: string, external: boolean } }} TitlePart
 * @param {string | undefined | null} title
 * @returns {{ text: string, parts: TitlePart[], links: { text: string, href: string, external: boolean }[] }}
 *   text is the readable plain form (link labels, no URLs) for tooltips,
 *   accessible names and matching. parts alternate plain text and links in
 *   order. links lists the safe links.
 */
export function parseTitle(title) {
  const source = title || '';
  /** @type {TitlePart[]} */
  const parts = [];
  /** @type {{ text: string, href: string, external: boolean }[]} */
  const links = [];
  let offset = 0;
  /** @param {string} text */
  const plain = (text) => {
    if (!text) return;
    const last = parts.at(-1);
    if (last && !last.link) last.text += text;
    else parts.push({ text });
  };
  for (const match of source.matchAll(LINK)) {
    plain(source.slice(offset, match.index));
    offset = match.index + match[0].length;
    const target = resolveHref(match[2] || match[3]);
    if (!target) {
      // An unsafe destination loses its link and keeps its label.
      plain(match[1]);
      continue;
    }
    parts.push({ text: match[1], link: target });
    links.push({ text: match[1], ...target });
  }
  plain(source.slice(offset));
  return { text: parts.map((part) => part.text).join(''), parts, links };
}
