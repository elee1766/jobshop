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
