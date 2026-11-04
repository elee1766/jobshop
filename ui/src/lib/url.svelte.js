// URL holds all UI state: view, open ticket, open session, filters.
// Back and forward work, and a link reproduces what you see.

/** @typedef {{ view: string, ticket: string, session: string, state: string, source: string, repo: string, kind: string, who: string }} UrlState */

const VIEWS = ['attention', 'chat', 'work', 'activity', 'sessions', 'pools', 'deploys'];

/** @returns {UrlState} */
function read() {
  const q = new URLSearchParams(location.search);
  const view = q.get('view') ?? 'attention';
  return {
    view: VIEWS.includes(view) ? view : 'attention',
    ticket: q.get('ticket') ?? '',
    session: q.get('session') ?? '',
    state: q.get('state') ?? '',
    // Activity filters (#255).
    source: q.get('source') ?? '',
    repo: q.get('repo') ?? '',
    kind: q.get('kind') ?? '',
    who: q.get('who') ?? '',
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
  else go({ session: '' });
}

export { VIEWS };
