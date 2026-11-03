// Formatting helpers. Everything from the controller is rendered as text
// ({...} in Svelte escapes it); nothing uses {@html}.

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
  if (ln.url && safe(ln.url)) out.push({ label: ln.identifier || 'Linear', href: ln.url });
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
