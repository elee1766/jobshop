// URL holds all UI state: view, open ticket, open session, filters.
// Back and forward work, and a link reproduces what you see.

/** @typedef {{ view: string, ticket: string, session: string, state: string }} UrlState */

const VIEWS = ['attention', 'chat', 'work', 'sessions', 'pools', 'deploys'];

/** @returns {UrlState} */
function read() {
  const q = new URLSearchParams(location.search);
  const view = q.get('view') ?? 'attention';
  return {
    view: VIEWS.includes(view) ? view : 'attention',
    ticket: q.get('ticket') ?? '',
    session: q.get('session') ?? '',
    state: q.get('state') ?? '',
  };
}

export const url = $state(read());

window.addEventListener('popstate', () => Object.assign(url, read()));

/**
 * Change URL state and push a history entry. Empty values are dropped.
 * @param {Partial<UrlState>} patch
 */
export function go(patch) {
  Object.assign(url, patch);
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(url)) {
    if (v && !(k === 'view' && v === 'attention')) q.set(k, v);
  }
  const s = q.toString();
  history.pushState(null, '', s ? '?' + s : location.pathname);
}

export { VIEWS };
