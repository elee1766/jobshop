// Canned resolve actions per attention item kind (#535, #603), shared by
// the Attention rows and the chat panel (#607). Each is an existing RPC
// through api.call, or an ack for kinds with no other action.

/**
 * @typedef {{ label: string, run: () => Promise<any>, ask: string }} Action
 * @param {any} api @param {string} kind @param {string} id
 * @returns {Action[]}
 */
export function resolveActions(api, kind, id) {
  if (!kind) return [];
  const call = (/** @type {string} */ m, /** @type {any} */ p) => () => api.call(m, p);
  const ack = () => api.post('api/attention/ack', { kind, id });
  switch (kind) {
    case 'run': return [{ label: 'Resume run', run: call('run.resume', { id }), ask: 'Resume run ' + id + '?' }];
    case 'pool': return [{ label: 'Unquarantine pool', run: call('pool.unquarantine', { pool: id }), ask: 'Unquarantine pool ' + id + '?' }];
    case 'wait': return [{ label: 'Clear wait', run: call('wait.clear', { id, reason: 'cleared in Desk' }), ask: 'Clear the wait on ' + id + '?' }];
    // decision.resolve acts on the proposal issue, closes the ticket and
    // wakes whoever waits on it (#603).
    case 'decision': return [
      { label: 'Approve as target', run: call('decision.resolve', { id, action: 'approve' }), ask: 'Approve ' + id + ': relabel the issue proposed to target?' },
      { label: 'Reject', run: call('decision.resolve', { id, action: 'reject' }), ask: 'Reject ' + id + ': close the issue as not planned?' },
      { label: 'Dismiss', run: call('decision.resolve', { id, action: 'dismiss' }), ask: 'Dismiss ' + id + ' for now? The issue is left as is.' },
    ];
    case 'ticket': case 'question': case 'stale':
      return [{ label: 'Close ticket', run: call('ticket.close', { id }), ask: 'Close ' + id + ' as done?' }];
    default: return [{ label: 'Acknowledge', run: ack, ask: 'Acknowledge ' + kind + ':' + id + '? It is hidden until it recurs.' }];
  }
}

/**
 * Confirms and runs a, with a toast either way. True when it succeeded.
 * @param {any} api @param {Action} a
 */
export async function runAction(api, a) {
  if (!(await api.confirm(a.ask))) return false;
  try {
    await a.run();
    api.toast(a.label + ': done');
    return true;
  } catch (e) {
    api.toast(a.label + ' failed: ' + String(/** @type {any} */ (e).message || e));
    return false;
  }
}

// Pending item actions (JOB-60): after an action succeeds the item stays
// in api/attention until the next load (seconds on a big store), so the
// row shows a pending state until a refreshed list drops it. The chat
// panel and the Attention view are separate bundles; they share the
// pending set through a global keyed by a registered symbol and a window
// event. Client only: a reload forgets it.
const PENDING = Symbol.for('jobshop.desk.attention.pending');
const PENDING_EVENT = 'jobshop-attention-pending';

/** @returns {Map<string, string>} item key (kind:id) to action label */
export function pendingItems() {
  const g = /** @type {any} */ (globalThis);
  return (g[PENDING] ??= new Map());
}

/** Marks an item pending after a successful action. @param {string} key @param {string} label */
export function markPending(key, label) {
  pendingItems().set(key, label);
  globalThis.dispatchEvent?.(new CustomEvent(PENDING_EVENT, { detail: { key, label } }));
}

/** Forgets a pending item. @param {string} key */
export function clearPending(key) {
  pendingItems().delete(key);
}

/** Calls f on each markPending; returns the unsubscribe. @param {(key: string, label: string) => void} f */
export function onPending(f) {
  const h = (/** @type {any} */ e) => f(e.detail.key, e.detail.label);
  globalThis.addEventListener?.(PENDING_EVENT, h);
  return () => globalThis.removeEventListener?.(PENDING_EVENT, h);
}
