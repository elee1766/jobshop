// API reads and the live change stream. Views call `load` and refetch
// when `live.tick` changes; events are only a signal, data comes from
// the JSON endpoints.

export const live = $state({ connected: false, tick: 0, lastSeq: 0 });

/** @type {Set<(connected: boolean) => void>} */
const liveFns = new Set();

/**
 * Called with the connection state now and on every change (api.live,
 * JOB-85). Returns an unsubscribe function.
 * @param {(connected: boolean) => void} fn
 */
export function onLive(fn) {
  liveFns.add(fn);
  fn(live.connected);
  return () => { liveFns.delete(fn); };
}

/** @param {boolean} v */
function setLive(v) {
  if (live.connected === v) return;
  live.connected = v;
  for (const fn of liveFns) {
    try { fn(v); } catch (e) { console.error('desk live handler', e); }
  }
}

/**
 * Where Desk is served, from the <base href> Desk writes into index.html
 * (#516): "/" at the root, "/desk/" under a prefix. Always ends in "/".
 * Every API and event URL is built from it; no other code builds an
 * absolute path.
 */
export const base = new URL('.', document.baseURI).pathname;

/**
 * The URL of a Desk path under base: api('api/work') -> "/desk/api/work".
 * @param {string} path relative path, with or without a leading slash
 */
export function api(path) {
  return base + path.replace(/^\/+/, '');
}

/**
 * @param {string} path
 * @returns {Promise<any>}
 */
export async function get(path) {
  const res = await fetch(api(path), { headers: { Accept: 'application/json' } });
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
  const res = await fetch(api(path), {
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

/**
 * Opens /api/events. Change bursts are coalesced to one refetch per second.
 * onEvent gets every parsed event (plugins' api.subscribe, #517). On a
 * reconnect EventSource sends Last-Event-ID, so the server resumes from
 * the last seq and no event is missed across a drop.
 * @param {(ev: any) => void} [onEvent]
 */
export function connect(onEvent) {
  const es = new EventSource(api('api/events'));
  es.addEventListener('open', () => setLive(true));
  es.addEventListener('error', () => setLive(false));
  const bump = (/** @type {MessageEvent} */ e) => {
    if (e.lastEventId) live.lastSeq = Number(e.lastEventId);
    if (onEvent) {
      try { onEvent({ kind: e.type, ...JSON.parse(e.data || '{}') }); } catch { /* not JSON */ }
    }
    if (pending) return;
    pending = window.setTimeout(() => { pending = 0; live.tick++; }, 1000);
  };
  for (const k of ['ticket', 'run', 'session', 'lagged', 'ready', 'activity']) es.addEventListener(k, bump);
  return () => es.close();
}
