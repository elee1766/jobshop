// URL holds all UI state: view, open ticket, open session, filters.
// Back and forward work, and a link reproduces what you see.

/** @typedef {{ view: string, ticket: string, session: string, state: string, source: string, repo: string, kind: string, who: string, artifact: string, trigger: string, pool: string, read: string }} UrlState */

const VIEWS = ['attention', 'artifacts', 'chat', 'work', 'activity', 'sessions', 'pools', 'deploys'];

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
  Object.assign(url, patch);
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(url)) {
    if (v && !(k === 'view' && v === 'attention')) q.set(k, v);
  }
  const s = q.toString();
  history.pushState(null, '', s ? '?' + s : location.pathname);
}

export { VIEWS };
