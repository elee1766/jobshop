// Shared sanitized Markdown for chat, waiting items and ticket bodies.
import { marked } from 'marked';
import DOMPurify from 'dompurify';

/** @param {string} text */
export function md(text) {
  const html = /** @type {string} */ (marked.parse(text || '', { async: false, gfm: true, breaks: false }));
  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe', 'object', 'embed', 'svg', 'math', 'img'],
    FORBID_ATTR: ['style'],
    ALLOWED_URI_REGEXP: /^(?:https?:|jobshop:)/i,
  });
  const doc = new DOMParser().parseFromString(clean, 'text/html');
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
