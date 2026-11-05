// URL holds all UI state: view, open ticket, open session, filters.
// Back and forward work, and a link reproduces what you see.

/** @typedef {{ view: string, ticket: string, session: string, item: string, state: string, source: string, repo: string, kind: string, who: string, artifact: string, trigger: string, pool: string, read: string }} UrlState */

// Built-in views. A plugin view (#517) is any other id; App falls back
// to attention when no plugin claims it.
const VIEWS = ['attention', 'artifacts', 'chat', 'work', 'activity', 'sessions', 'pools'];
const viewRe = /^[a-z0-9][a-z0-9-]{0,62}$/;

/** @returns {UrlState} */
function read() {
  const q = new URLSearchParams(location.search);
  const view = q.get('view') ?? 'attention';
  return {
    view: viewRe.test(view) ? view : 'attention',
    ticket: q.get('ticket') ?? '',
    session: q.get('session') ?? '',
    // The attention item (kind:id) an open session resolves (#535).
    item: q.get('item') ?? '',
    state: q.get('state') ?? '',
    // Activity filters (#255).
    source: q.get('source') ?? '',
    repo: q.get('repo') ?? '',
    kind: q.get('kind') ?? '',
    who: q.get('who') ?? '',
    // Artifacts (#508): open artifact, filters, show read ones.
    artifact: q.get('artifact') ?? '',
    trigger: q.get('trigger') ?? '',
    pool: q.get('pool') ?? '',
    read: q.get('read') ?? '',
  };
}

export const url = $state(read());

window.addEventListener('popstate', () => Object.assign(url, read()));

/**
 * Change URL state and push a history entry. Empty values are dropped.
 * @param {Partial<UrlState>} patch
 */
export function go(patch) {
  // The entry that opens the session panel is marked, so closing it can
  // step back instead of stacking another entry (see closeSession).
  const opening = !!patch.session && !url.session;
  Object.assign(url, patch);
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(url)) {
    if (v && !(k === 'view' && v === 'attention')) q.set(k, v);
  }
  const s = q.toString();
  history.pushState(opening ? { panel: 1 } : null, '', s ? '?' + s : location.pathname);
}

/** Close the session panel; back when it was opened from this page, so back and close agree. */
export function closeSession() {
  if (history.state?.panel) history.back();
  else go({ session: '', item: '' });
}

export { VIEWS };
