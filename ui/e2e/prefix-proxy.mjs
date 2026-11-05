// A small reverse proxy for the prefix e2e run (#516): serves DESK_URL
// under /desk/ on PREFIX_PORT. It strips the prefix and sends
// X-Forwarded-Prefix, so Desk needs --trusted-proxy 127.0.0.1 (or
// --base-path /desk/, which also accepts the stripped path).
import http from 'node:http';

const upstream = new URL(process.env.DESK_URL || 'http://127.0.0.1:7480');
const port = Number(process.env.PREFIX_PORT || 7499);
const prefix = '/desk/';

http
  .createServer((req, res) => {
    const path = req.url || '/';
    if (path === '/desk') {
      res.writeHead(301, { Location: prefix }).end();
      return;
    }
    if (!path.startsWith(prefix)) {
      res.writeHead(404).end('outside ' + prefix);
      return;
    }
    const up = http.request(
      {
        host: upstream.hostname,
        port: upstream.port,
        method: req.method,
        path: '/' + path.slice(prefix.length),
        // Host stays the browser's (loopback), so Desk's Host and Origin
        // checks see what the browser sent.
        headers: { ...req.headers, 'x-forwarded-prefix': prefix },
      },
      (r) => {
        res.writeHead(r.statusCode || 502, r.headers);
        r.pipe(res);
      },
    );
    up.on('error', (e) => { res.writeHead(502).end(String(e)); });
    req.pipe(up);
  })
  .listen(port, '127.0.0.1', () => console.log('prefix proxy on http://127.0.0.1:' + port + prefix));
