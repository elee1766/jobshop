// URL holds all UI state: view, open ticket, open session, filters.
// Back and forward work, and a link reproduces what you see. Any query
// key is kept, so a plugin view stores its own state here (#541).

/** @typedef {{ view: string, ticket: string, session: string, item: string, [k: string]: string }} UrlState */

// Every view is a plugin (#541); `view` names a plugin view id and App
// falls back to attention.
const viewRe = /^[a-z0-9][a-z0-9-]{0,62}$/;
const keyRe = /^[a-z][a-z0-9_]{0,31}$/;

/** @returns {UrlState} */
function read() {
  const q = new URLSearchParams(location.search);
  /** @type {UrlState} */
  const out = { view: 'attention', ticket: '', session: '', item: '' };
  for (const [k, v] of q) if (keyRe.test(k)) out[k] = v;
  if (!viewRe.test(out.view)) out.view = 'attention';
  return out;
}

export const url = $state(read());

/** @type {Set<() => void>} */
const changed = new Set();

/** Called after every URL change. @param {() => void} fn */
export function onUrl(fn) {
  changed.add(fn);
  return () => { changed.delete(fn); };
}

/** @param {UrlState} next */
function replace(next) {
  for (const k of Object.keys(url)) if (!(k in next)) delete url[k];
  Object.assign(url, next);
  for (const fn of changed) fn();
}

window.addEventListener('popstate', () => replace(read()));

/**
 * Change URL state and push a history entry. Empty values are dropped.
 * @param {Record<string, string>} patch
 */
export function go(patch) {
  // The entry that opens the session panel is marked, so closing it can
  // step back instead of stacking another entry (see closeSession).
  const opening = !!patch.session && !url.session;
  const next = /** @type {UrlState} */ ({ ...url });
  // Read-only (ro=1, JOB-54) belongs to the session it was opened for.
  if ('session' in patch && !('ro' in patch)) next.ro = '';
  for (const [k, v] of Object.entries(patch)) if (keyRe.test(k)) next[k] = v ?? '';
  if (!viewRe.test(next.view)) next.view = 'attention';
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(next)) {
    if (v && !(k === 'view' && v === 'attention')) q.set(k, v);
  }
  const s = q.toString();
  history.pushState(opening ? { panel: 1 } : null, '', s ? '?' + s : location.pathname);
  replace(next);
}

/**
 * Open a view with only the given params: another view's filters do not
 * carry over.
 * @param {string} view @param {Record<string, string>} [params]
 */
export function navigate(view, params) {
  const keep = /** @type {Record<string, string>} */ ({});
  for (const k of Object.keys(url)) keep[k] = '';
  go({ ...keep, ...(params || {}), view });
}

/** Close the session panel; back when it was opened from this page, so back and close agree. */
export function closeSession() {
  if (history.state?.panel) history.back();
  else go({ session: '', item: '' });
}
