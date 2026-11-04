// API reads and the live change stream. Views call `load` and refetch
// when `live.tick` changes; events are only a signal, data comes from
// the JSON endpoints.

export const live = $state({ connected: false, tick: 0, lastSeq: 0 });

/**
 * @param {string} path
 * @returns {Promise<any>}
 */
export async function get(path) {
  const res = await fetch(path, { headers: { Accept: 'application/json' } });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || res.status + ' ' + res.statusText);
  return body;
}

/**
 * POST JSON with the CSRF header Desk requires on writes.
 * @param {string} path
 * @param {any} data
 * @returns {Promise<any>}
 */
export async function post(path, data) {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json', 'X-Jobshop-Desk': '1' },
    body: JSON.stringify(data),
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || res.status + ' ' + res.statusText);
  return body;
}

/**
 * A reactive resource: refetches when deps change or a change event arrives.
 * @param {() => string} path
 */
export function resource(path) {
  const r = $state({ data: /** @type {any} */ (null), error: '', loading: true });
  $effect(() => {
    const p = path();
    void live.tick;
    let stale = false;
    get(p).then(
      (d) => { if (!stale) Object.assign(r, { data: d, error: '', loading: false }); },
      (e) => { if (!stale) Object.assign(r, { error: String(e.message || e), loading: false }); },
    );
    return () => { stale = true; };
  });
  return r;
}

let pending = 0;

/** Opens /api/events. Change bursts are coalesced to one refetch per second. */
export function connect() {
  const es = new EventSource('/api/events');
  es.addEventListener('open', () => { live.connected = true; });
  es.addEventListener('error', () => { live.connected = false; });
  const bump = (/** @type {MessageEvent} */ e) => {
    if (e.lastEventId) live.lastSeq = Number(e.lastEventId);
    if (pending) return;
    pending = window.setTimeout(() => { pending = 0; live.tick++; }, 1000);
  };
  for (const k of ['ticket', 'run', 'session', 'lagged', 'ready', 'activity']) es.addEventListener(k, bump);
  return () => es.close();
}
