// Desk service worker: shows Web Push notifications and opens Desk on
// click. No offline cache: Desk is live data over the admin socket.

self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (e) => e.waitUntil(self.clients.claim()));

self.addEventListener('push', (e) => {
  let m = { title: 'Jobshop Desk', body: '', tag: '', url: '/' };
  try { m = { ...m, ...e.data.json() }; } catch { /* keep defaults */ }
  e.waitUntil(self.registration.showNotification(m.title, {
    body: m.body,
    tag: m.tag || undefined,
    icon: '/icon-192.png',
    badge: '/badge.png',
    data: { url: m.url },
  }));
});

self.addEventListener('notificationclick', (e) => {
  e.notification.close();
  const url = new URL(e.notification.data?.url || '/', self.location.origin).href;
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
