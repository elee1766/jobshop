// Markdown to sanitized HTML, after jcode-web's format.ts (#535). In
// ui/kit so the chat and attention plugins each bundle their own copy
// (#541): a plugin may not import the shell (#517).
import { marked } from 'marked';
import DOMPurify from 'dompurify';

/** Agent markdown to sanitized HTML; links open in a new tab. @param {string} text */
export function md(text) {
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
  return doc.body.innerHTML;
}
