// Formatting helpers. Everything from the controller is rendered as text
// ({...} in Svelte escapes it).

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
    case 'needs_human': case 'in_progress': case 'request_changes': return 'warn';
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

/**
 * A tool call's label and secondary text. Agents title tools their own
 * way: jcode titles by tool type ("Running shell command") and sends a
 * short intent with each call, so for jcode the intent leads; any other
 * agent keeps its title, with the command, path or query as detail.
 * @param {{ name?: string } | null | undefined} agent
 * @param {{ title?: string, target?: string, intent?: string }} t
 * @returns {{ label: string, detail: string }}
 */
export function toolLabel(agent, t) {
  const title = t.title || '', target = t.target || '', intent = t.intent || '';
  if (agent?.name === 'jcode' && intent) return { label: intent, detail: target || title };
  if (title) return { label: title, detail: target || intent };
  return { label: target || intent || 'tool', detail: target && intent ? intent : '' };
}
