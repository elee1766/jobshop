// Desk service worker: shows Web Push notifications and opens Desk on
// click. No offline cache: Desk is live data over the admin socket.
// Paths are relative to the worker's scope, Desk's base (#516); a push
// message's url is relative too ("./?ticket=js-1").
const at = (/** @type {string} */ p) => new URL(p, self.registration.scope).href;

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));

self.addEventListener('push', (e) => {
  let m = { title: 'Jobshop Desk', body: '', tag: '', url: './' };
  try { m = { ...m, ...e.data.json() }; } catch { /* keep defaults */ }
  e.waitUntil(self.registration.showNotification(m.title, {
    body: m.body,
    tag: m.tag || undefined,
    icon: at('icon-192.png'),
    badge: at('badge.png'),
    data: { url: m.url },
  }));
});

self.addEventListener('notificationclick', (e) => {
  e.notification.close();
  const url = at(e.notification.data?.url || './');
  e.waitUntil((async () => {
    const wins = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
    for (const w of wins) {
      if (new URL(w.url).origin === self.location.origin) {
        await w.focus();
        return w.navigate(url);
      }
    }
    return self.clients.openWindow(url);
  })());
});
