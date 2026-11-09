// Fixture apps for the Desk embed checks (JOB-151). Real HTTP servers, no mocks:
//
//   node e2e/embed-fixture.mjs            ports 7497 (localhost) and 7498 (127.0.0.1)
//
// The same process answers on two host names so one is cross-host and one is
// Desk's own host when Desk is opened at 127.0.0.1. Routes:
//
//   /app      a page that probes what it can reach and writes the result into
//             #probe, and polls /tick so the spec can see it is alive
//   /deny     X-Frame-Options: DENY
//   /ancestor Content-Security-Policy: frame-ancestors 'none'
//   /seen     JSON of the last /app request headers and the tick count
//   /reset    zeroes the tick count
import http from 'node:http';

const cross = Number(process.env.EMBED_CROSS_PORT || 7497);
const same = Number(process.env.EMBED_SAME_PORT || 7498);
let ticks = 0;
let seen = {};

const page = `<!doctype html>
<meta charset="utf-8">
<title>fixture app</title>
<h1 id="hello">fixture app</h1>
<pre id="probe">probing</pre>
<script>
const result = {};
const attempt = (name, fn) => { try { result[name] = String(fn()); } catch (e) { result[name] = 'blocked:' + e.name; } };
attempt('parentDocument', () => window.parent.document.title);
attempt('parentLocation', () => window.parent.location.href);
attempt('parentStore', () => window.parent.localStorage.length);
attempt('cookie', () => document.cookie);
attempt('localStorage', () => localStorage.length);
attempt('topNavigation', () => { window.top.location = 'about:blank'; return 'navigated'; });
attempt('origin', () => window.origin);
attempt('popup', () => window.open('about:blank') ? 'opened' : 'blocked:null');
attempt('modal', () => { alert('x'); return 'shown'; });
(async () => {
  try {
    const r = await fetch(new URL('/api/whoami', document.referrer || location.origin).href, { credentials: 'include' });
    result.deskApi = 'status:' + r.status;
  } catch (e) { result.deskApi = 'blocked:' + e.name; }
  document.getElementById('probe').textContent = JSON.stringify(result);
})();
setInterval(() => fetch('/tick').catch(() => {}), 150);
</script>`;

function handler(req, res) {
  const url = new URL(req.url || '/', 'http://x');
  const send = (status, type, body, headers = {}) => {
    res.writeHead(status, { 'content-type': type, 'cache-control': 'no-store', ...headers });
    res.end(body);
  };
  switch (url.pathname) {
    case '/app':
      seen = { headers: req.headers, at: Date.now() };
      return send(200, 'text/html', page);
    case '/deny':
      return send(200, 'text/html', '<title>denied</title><h1>should not show</h1>', { 'x-frame-options': 'DENY' });
    case '/ancestor':
      return send(200, 'text/html', '<title>ancestor</title><h1>should not show</h1>', {
        'content-security-policy': "frame-ancestors 'none'",
      });
    case '/tick':
      ticks++;
      return send(204, 'text/plain', '');
    case '/seen':
      return send(200, 'application/json', JSON.stringify({ ...seen, ticks }), { 'access-control-allow-origin': '*' });
    case '/reset':
      ticks = 0;
      return send(204, 'text/plain', '', { 'access-control-allow-origin': '*' });
    default:
      return send(404, 'text/plain', 'not found');
  }
}

for (const [host, port] of [
  ['localhost', cross],
  ['127.0.0.1', same],
]) {
  http.createServer(handler).listen(port, host, () => console.log(`embed fixture on http://${host}:${port}`));
}
