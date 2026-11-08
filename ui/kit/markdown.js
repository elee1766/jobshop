// Shared sanitized Markdown for chat, waiting items and ticket bodies.
import DOMPurify from 'dompurify';
import { marked } from './links.js';

/** @param {string} text */
export function md(text, inline = false) {
  const html = /** @type {string} */ ((inline ? marked.parseInline : marked.parse)(text || ''));
  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe', 'object', 'embed', 'svg', 'math', 'img'],
    FORBID_ATTR: ['style'],
    ALLOWED_URI_REGEXP: /^(?:https?:|jobshop:)/i,
  });
  const doc = new DOMParser().parseFromString(clean, 'text/html');
  // Only link prose, never authored links or code examples.
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
  const nodes = [];
  while (walker.nextNode()) nodes.push(walker.currentNode);
  for (const node of nodes) {
    if (node.parentElement?.closest('a, code, pre')) continue;
    const text = node.textContent || '';
    const matches = [...text.matchAll(/(?<![\w./@-])([A-Za-z0-9][A-Za-z0-9-]*)\/([A-Za-z0-9_.-]+)#([1-9][0-9]*)(?![\w#])/g)];
    if (!matches.length) continue;
    const fragment = doc.createDocumentFragment();
    let offset = 0;
    for (const match of matches) {
      fragment.append(text.slice(offset, match.index));
      const link = doc.createElement('a');
      link.textContent = match[0];
      link.href = `https://github.com/${match[1]}/${match[2]}/pull/${match[3]}`;
      fragment.append(link);
      offset = match.index + match[0].length;
    }
    fragment.append(text.slice(offset));
    node.parentNode?.replaceChild(fragment, node);
  }
  for (const a of doc.querySelectorAll('a')) {
    const href = a.getAttribute('href') || '';
    const internal = /^jobshop:\/\/(ticket|session)\/([^/?#]+)$/i.exec(href);
    if (internal) {
      try {
        const id = decodeURIComponent(internal[2]);
        if (/^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(id)) {
          a.setAttribute('href', '?view=trace&ref=' + encodeURIComponent(id));
          a.removeAttribute('target');
          a.removeAttribute('rel');
          continue;
        }
      } catch { /* malformed identifier */ }
    }
    let ok = false;
    try { ok = ['http:', 'https:'].includes(new URL(href).protocol); } catch { /* relative or bad */ }
    if (ok) {
      a.target = '_blank';
      a.rel = 'noopener noreferrer';
    } else {
      a.removeAttribute('href');
      a.removeAttribute('target');
      a.removeAttribute('rel');
    }
  }
  return doc.body.innerHTML;
}
