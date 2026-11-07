import { test } from "node:test";
import assert from "node:assert/strict";
import { createServer } from "node:http";
import { once } from "node:events";
import { createClient, ApiError } from "./index.js";

test("controller transport preserves proxy prefix, CSRF, errors and cancellation", async (t) => {
  const received = [];
  const server = createServer(async (req, res) => {
    if (req.url === "/desk/api/slow") return;
    let body = "";
    for await (const part of req) body += part;
    received.push({ url: req.url, csrf: req.headers["x-jobshop-desk"], body });
    res.setHeader("Content-Type", "application/json");
    if (req.url === "/desk/api/rpc/ticket.show")
      res.end(JSON.stringify({ ticket: { id: JSON.parse(body).id } }));
    else {
      res.statusCode = 409;
      res.end(JSON.stringify({ error: "Already closed", code: "conflict" }));
    }
  });
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  t.after(() => {
    server.closeAllConnections();
    server.close();
  });
  const client = createClient({
    baseURL: `http://127.0.0.1:${server.address().port}/desk/`,
  });
  assert.deepEqual(await client.call("ticket.show", { id: "js-1" }), {
    ticket: { id: "js-1" },
  });
  assert.deepEqual(received[0], {
    url: "/desk/api/rpc/ticket.show",
    csrf: "1",
    body: '{"id":"js-1"}',
  });
  await assert.rejects(
    client.get("api/missing"),
    (error) =>
      error instanceof ApiError &&
      error.status === 409 &&
      error.code === "conflict",
  );
  const controller = new AbortController();
  const pending = client.get("api/slow", { signal: controller.signal });
  controller.abort();
  await assert.rejects(pending, { name: "AbortError" });
  for (const path of ["../api", "https://example.com/", "//example.com/../.."])
    assert.throws(() => client.url(path), TypeError);
});
