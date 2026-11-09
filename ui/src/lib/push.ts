// Web Push (shell, #535): register the service worker, subscribe with
// Desk's VAPID key and hand the subscription to Desk, which pushes new
// attention items. The system page shows the opt-in button.
import { atom, getDefaultStore } from 'jotai';
import { get, post, api } from './api.ts';

/** unsupported | off | on | denied | unavailable (Desk runs without --push) */
export type PushState = 'unsupported' | 'off' | 'on' | 'denied' | 'unavailable';

export const pushStateAtom = atom<PushState>('off');
export const pushErrorAtom = atom('');

const store = () => getDefaultStore();
const setState = (state: PushState) => store().set(pushStateAtom, state);
const setError = (error: string) => store().set(pushErrorAtom, error);
const message = (e: unknown) => (e instanceof Error ? e.message : String(e));

const supported = 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window;

function key(s: string) {
  const b = atob(s.replace(/-/g, '+').replace(/_/g, '/'));
  return Uint8Array.from(b, (c) => c.charCodeAt(0));
}

let ready: Promise<void> | null = null;

/** Registers the worker once and reads the subscription; later calls wait for the first. */
export function initPush(): Promise<void> {
  ready ??= (async () => {
    if (!supported) {
      setState('unsupported');
      return;
    }
    try {
      const available = await get('api/push/key');
      if (!available.key) {
        setState('unavailable');
        return;
      }
    } catch {
      setState('unavailable');
      return;
    }
    try {
      // Base-relative, so the worker's scope is Desk's prefix (#516).
      const reg = await navigator.serviceWorker.register(api('sw.js'), { scope: api('') });
      if (Notification.permission === 'denied') {
        setState('denied');
        return;
      }
      const sub = await reg.pushManager.getSubscription();
      setState(sub ? 'on' : 'off');
      // Re-send so Desk has it even if its state file was reset.
      if (sub) await post('api/push/subscribe', sub.toJSON()).catch(() => {});
    } catch (e) {
      setError(message(e));
    }
  })();
  return ready;
}

export async function enablePush() {
  await initPush();
  setError('');
  try {
    const perm = await Notification.requestPermission();
    if (perm !== 'granted') {
      setState(perm === 'denied' ? 'denied' : 'off');
      return;
    }
    const { key: k } = await get('api/push/key');
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: key(k) });
    await post('api/push/subscribe', sub.toJSON());
    setState('on');
  } catch (e) {
    setError(message(e));
  }
}

export async function disablePush() {
  await initPush();
  setError('');
  try {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();
    if (sub) {
      await post('api/push/unsubscribe', { endpoint: sub.endpoint });
      await sub.unsubscribe();
    }
    setState('off');
  } catch (e) {
    setError(message(e));
  }
}
