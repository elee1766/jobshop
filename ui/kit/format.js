// Formatting helpers. Everything from the controller is rendered as text
// ({...} in Svelte escapes it). The one exception is markdown(), used by
// the Artifacts view.

import { marked } from 'marked';
import DOMPurify from 'dompurify';

/**
 * Renders an artifact body (#508) for {@html}. The body is written by an
 * agent, so it is untrusted: marked turns it into HTML and DOMPurify
 * strips scripts, event handlers, iframes, forms and style, keeping only
 * plain document markup. Links open in a new tab without a referrer.
 * The page CSP (script-src 'self') is a second guard.
 * @param {string} md
 */
export function markdown(md) {
  const html = /** @type {string} */ (marked.parse(md || '', { async: false, gfm: true }));
  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS: ['style', 'form', 'input', 'button', 'iframe', 'object', 'embed', 'svg', 'math'],
    FORBID_ATTR: ['style'],
  });
  const doc = new DOMParser().parseFromString(clean, 'text/html');
  for (const a of doc.querySelectorAll('a')) {
    if (!safe(a.href)) a.removeAttribute('href');
    a.setAttribute('target', '_blank');
    a.setAttribute('rel', 'noopener noreferrer');
  }
  return doc.body.innerHTML;
}

/** @param {number} n */
export function bytes(n) {
  if (n < 1024) return n + ' B';
  if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
  return (n / 1024 / 1024).toFixed(1) + ' MB';
}

/** @param {string | undefined | null} t */
export function ago(t) {
  if (!t || t.startsWith('0001')) return '';
  const s = Math.round((Date.now() - new Date(t).getTime()) / 1000);
  const abs = Math.abs(s);
  const sign = s < 0 ? 'in ' : '';
  const tail = s < 0 ? '' : ' ago';
  if (abs < 60) return sign + abs + 's' + tail;
  if (abs < 3600) return sign + Math.round(abs / 60) + 'm' + tail;
  if (abs < 86400) return sign + Math.round(abs / 3600) + 'h' + tail;
  return sign + Math.round(abs / 86400) + 'd' + tail;
}

/** @param {any} t */
export function fields(t) {
  if (!t || !t.fields) return {};
  return typeof t.fields === 'string' ? JSON.parse(t.fields) : t.fields;
}

/** @param {any} t */
export function meta(t) {
  if (!t || !t.meta) return {};
  return typeof t.meta === 'string' ? JSON.parse(t.meta) : t.meta;
}

/** @param {string} s */
export function tone(s) {
  switch (s) {
    case 'done': case 'running': case 'closed': return 'ok';
    case 'needs_human': case 'in_progress': return 'warn';
    case 'failed': case 'cancelled': case 'crashed': case 'stuck': return 'bad';
    default: return '';
  }
}

/**
 * Links to GitHub and Linear from a ticket's meta.
 * @param {any} t
 * @returns {{ label: string, href: string }[]}
 */
export function links(t) {
  const m = meta(t);
  /** @type {{ label: string, href: string }[]} */
  const out = [];
  const gh = m.github || {};
  const repo = gh.repo || 'elee1766/jobshop';
  if (gh.issue) out.push({ label: 'issue #' + gh.issue, href: `https://github.com/${repo}/issues/${gh.issue}` });
  if (gh.pr) out.push({ label: 'PR #' + gh.pr, href: `https://github.com/${repo}/pull/${gh.pr}` });
  const ln = m.linear || {};
  // Runs store workspace and issue key (control/runs.go), not a url.
  const key = ln.issue || ln.identifier;
  const href = ln.url || (ln.workspace && key ? `https://linear.app/${encodeURIComponent(ln.workspace)}/issue/${encodeURIComponent(key)}` : '');
  if (href && safe(href)) out.push({ label: key || 'Linear', href });
  return out;
}

/** Only http(s) links are rendered as anchors. @param {string} u */
export function safe(u) {
  try {
    const p = new URL(u);
    return p.protocol === 'https:' || p.protocol === 'http:';
  } catch {
    return false;
  }
}

/** Session ids are ticket ids (project-suffix); actors such as admin or controller are not. @param {string | undefined | null} s */
export function isSessionId(s) {
  return !!s && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(s);
}

/** Name for a chat session key: escalation:<ticket> and desk:<id> read plainly, anything else by its prefix. @param {string} k */
export function chatLabel(k) {
  if (k.startsWith('escalation:')) return 'Escalation ' + k.slice('escalation:'.length);
  if (k.startsWith('desk:')) return 'Desk chat';
  return k.split(':')[0] || 'Session';
}
