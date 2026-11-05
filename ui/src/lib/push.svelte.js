// Web Push (shell, #535): register the service worker, subscribe with
// Desk's VAPID key and hand the subscription to Desk, which pushes new
// attention items. Plugins reach this through api.push; the built-in
// Attention plugin shows the opt-in button.

import { get, post, api } from './api.svelte.js';

export const push = $state({
  /** unsupported | off | on | denied | unavailable (Desk runs without --push) */
  state: 'off',
  error: '',
});

const supported = 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;

/** @param {string} s */
function key(s) {
  const b = atob(s.replace(/-/g, '+').replace(/_/g, '/'));
  return Uint8Array.from(b, (c) => c.charCodeAt(0));
}

/** @type {Promise<void> | null} */
let ready = null;

/** Registers the worker once and reads the subscription; later calls wait for the first. */
export function initPush() {
  ready ??= (async () => {
    if (!supported) { push.state = 'unsupported'; return; }
    try {
      await get('api/push/key');
    } catch {
      push.state = 'unavailable';
      return;
    }
    try {
      // Base-relative, so the worker's scope is Desk's prefix (#516).
      const reg = await navigator.serviceWorker.register(api('sw.js'), { scope: api('') });
      if (Notification.permission === 'denied') { push.state = 'denied'; return; }
      const sub = await reg.pushManager.getSubscription();
      push.state = sub ? 'on' : 'off';
      // Re-send so Desk has it even if its state file was reset.
      if (sub) await post('api/push/subscribe', sub.toJSON()).catch(() => {});
    } catch (e) {
      push.error = String(/** @type {any} */ (e).message || e);
    }
  })();
  return ready;
}

/** @returns {Promise<{ state: string, error: string }>} */
export async function pushStatus() {
  await initPush();
  return { state: push.state, error: push.error };
}

export async function enablePush() {
  await initPush();
  push.error = '';
  try {
    const perm = await Notification.requestPermission();
    if (perm !== 'granted') { push.state = perm === 'denied' ? 'denied' : 'off'; return; }
    const { key: k } = await get('api/push/key');
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: key(k) });
    await post('api/push/subscribe', sub.toJSON());
    push.state = 'on';
  } catch (e) {
    push.error = String(/** @type {any} */ (e).message || e);
  }
}

export async function disablePush() {
  await initPush();
  push.error = '';
  try {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();
    if (sub) {
      await post('api/push/unsubscribe', { endpoint: sub.endpoint });
      await sub.unsubscribe();
    }
    push.state = 'off';
  } catch (e) {
    push.error = String(/** @type {any} */ (e).message || e);
  }
}

export async function testPush() {
  push.error = '';
  try { await post('api/push/test', {}); } catch (e) { push.error = String(/** @type {any} */ (e).message || e); }
}
