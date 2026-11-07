// Shared sanitized Markdown for chat, waiting items and ticket bodies.
import { marked } from 'marked';
import DOMPurify from 'dompurify';

/** @param {string} text @param {Record<string, string>} refs Known references and their ticket IDs. */
export function md(text, refs = {}) {
  const html = /** @type {string} */ (marked.parse(text || '', { async: false, gfm: true, breaks: false }));
  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe', 'object', 'embed', 'svg', 'math', 'img'],
    FORBID_ATTR: ['style'],
  });
  const doc = new DOMParser().parseFromString(clean, 'text/html');
  for (const a of doc.querySelectorAll('a')) {
    let ok = false;
    try { ok = ['http:', 'https:'].includes(new URL(a.href).protocol); } catch { /* relative or bad */ }
    if (!ok) a.removeAttribute('href');
    a.target = '_blank';
    a.rel = 'noopener noreferrer';
  }
  const names = Object.keys(refs).filter(Boolean).sort((a, b) => b.length - a.length);
  if (names.length) {
    const pattern = new RegExp('(?<![\\w-])(?:' + names.map(name => name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|') + ')(?![\\w-])', 'g');
    const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
    const nodes = [];
    while (walker.nextNode()) nodes.push(walker.currentNode);
    for (const node of nodes) {
      if (node.parentElement?.closest('a, pre, code')) continue;
      const text = node.textContent || '';
      const matches = [...text.matchAll(pattern)];
      if (!matches.length) continue;
      const fragment = doc.createDocumentFragment();
      let offset = 0;
      for (const match of matches) {
        fragment.append(text.slice(offset, match.index));
        const link = doc.createElement('a');
        link.textContent = match[0];
        link.href = '?view=trace&ref=' + encodeURIComponent(refs[match[0]]);
        fragment.append(link);
        offset = match.index + match[0].length;
      }
      fragment.append(text.slice(offset));
      node.parentNode?.replaceChild(fragment, node);
    }
  }
  return doc.body.innerHTML;
}
