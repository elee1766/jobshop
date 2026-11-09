import { useEffect } from 'react';
import { client } from '../lib/api.ts';
import { initPush } from '../lib/push.ts';
import { useDeskRoute } from '../lib/route.ts';
import { syncTrace } from '../state/actions.ts';
import { connectedAtom, lastFocusAtom, openStream, type StreamParams } from '../state/live.ts';
import { store } from '../state/store.ts';
import { keyboardAtom } from '../state/ui.ts';

// LiveOwner owns the tab's one api/live connection. It sits above the routed
// screens and renders nothing: a navigation that does not change the stream's
// scope (view, ref, inspect, audit) keeps the same connection, and one that
// does reopens it once. Offline closes it and online reopens it.
export function LiveOwner() {
  const route = useDeskRoute();
  // The scope is a string, so the stream reopens only when it changes.
  const scope = JSON.stringify(
    route.view === 'trace'
      ? { trace: route.ref, inspect: route.inspect }
      : route.view === 'system' && route.audit
        ? { audit: route.audit }
        : {},
  );

  useEffect(() => {
    syncTrace(route.ref, route.inspect);
  }, [route.ref, route.inspect]);

  useEffect(() => {
    const params: StreamParams = JSON.parse(scope);
    if (!navigator.onLine) {
      store.set(connectedAtom, false);
    }
    let close = () => {};
    const open = () => {
      close();
      if (!navigator.onLine) return;
      close = openStream(params);
    };
    const offline = () => {
      close();
      close = () => {};
      store.set(connectedAtom, false);
    };
    window.addEventListener('online', open);
    window.addEventListener('offline', offline);
    open();
    return () => {
      window.removeEventListener('online', open);
      window.removeEventListener('offline', offline);
      close();
    };
  }, [scope]);

  return null;
}

// Once per tab: the last-focus time for catch-up, push registration, the
// visual viewport (mobile keyboard) and the focus ping on hide.
export function useTabLifecycle() {
  useEffect(() => {
    void client
      .get('api/focus')
      .then((d) => store.set(lastFocusAtom, d.last_focus || ''))
      .catch(() => {});
    void initPush();
    const vv = window.visualViewport;
    let fullHeight = vv?.height || window.innerHeight;
    const typing = () => !!document.activeElement?.matches('textarea,input');
    const fit = () => {
      const height = vv?.height || window.innerHeight;
      document.documentElement.style.setProperty('--vvh', height + 'px');
      document.documentElement.style.setProperty('--vvtop', (vv?.offsetTop || 0) + 'px');
      if (!typing()) fullHeight = Math.max(fullHeight, height);
      store.set(keyboardAtom, height < fullHeight * 0.78 && typing());
    };
    const visibility = () => {
      if (document.visibilityState === 'hidden') void client.post('api/focus', {}).catch(() => {});
    };
    fit();
    vv?.addEventListener('resize', fit);
    vv?.addEventListener('scroll', fit);
    document.addEventListener('visibilitychange', visibility);
    return () => {
      vv?.removeEventListener('resize', fit);
      vv?.removeEventListener('scroll', fit);
      document.removeEventListener('visibilitychange', visibility);
    };
  }, []);
}
