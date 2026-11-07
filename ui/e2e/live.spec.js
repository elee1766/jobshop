import { test, expect } from "@playwright/test";

// Opt in only against an isolated controller: this creates and closes a task.
test.skip(
  !process.env.DESK_MUTATE,
  "set DESK_MUTATE=1 against an isolated controller",
);

test("live state changes and reconnect snapshots need no browser refresh requests", async ({
  page,
  context,
  request,
  baseURL,
  isMobile,
}) => {
  const reads = [];
  page.on("request", (req) => {
    if (
      /\/api\/(overview|system|improvements|chat|trace)(\/|\?|$)|\/api\/rpc\/attention$/.test(
        new URL(req.url()).pathname,
      )
    )
      reads.push(req.url());
  });
  const call = async (method, data) => {
    const response = await request.post(
      new URL("api/rpc/" + method, baseURL).href,
      {
        headers: { "X-Jobshop-Desk": "1" },
        data,
      },
    );
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  let ticket;
  const related = [];
  try {
    await page.goto("./");
    await expect(page.locator(".status")).toContainText("sys ok");
    const title = "Live delivery check " + crypto.randomUUID();
    ticket = await call("ticket.create", { title, labels: ["needs-human"] });
    const waiting = page.getByRole("region", { name: "Waiting", exact: true });
    await expect(waiting).toContainText(title);
    await call("ticket.update", { id: ticket.id, title: title + " updated" });
    await expect(waiting).toContainText(title + " updated");
    const cursor = await page.evaluate(
      () =>
        new Promise((resolve, reject) => {
          const source = new EventSource(
            new URL("api/events", document.baseURI),
          );
          source.addEventListener("ready", (event) => {
            source.close();
            resolve(Number(event.lastEventId));
          });
          source.onerror = () => {
            source.close();
            reject(new Error("event stream failed"));
          };
        }),
    );
    expect(cursor).toBeGreaterThan(0);
    await context.setOffline(true);
    await expect(page.locator(".status")).not.toContainText("sys ok");
    await call("ticket.update", {
      id: ticket.id,
      title: title + " during disconnect",
    });
    await context.setOffline(false);
    await expect(waiting).toContainText(title + " during disconnect");
    const replay = await page.evaluate(
      ({ cursor, id }) =>
        new Promise((resolve, reject) => {
          const events = [];
          const source = new EventSource(
            new URL("api/events?since=" + cursor, document.baseURI),
          );
          source.addEventListener("ticket", (event) => {
            const data = JSON.parse(event.data);
            if (data.id === id)
              events.push({
                seq: data.seq,
                id: event.lastEventId,
                op: data.op,
              });
          });
          source.addEventListener("ready", () => {
            source.close();
            resolve(events);
          });
          source.onerror = () => {
            source.close();
            reject(new Error("replay failed"));
          };
        }),
      { cursor, id: ticket.id },
    );
    expect(replay).not.toHaveLength(0);
    expect(
      replay.every(
        (event) => event.seq > cursor && event.id === String(event.seq),
      ),
    ).toBeTruthy();
    expect(new Set(replay.map((event) => event.seq)).size).toBe(replay.length);
    await waiting.getByRole('link', { name: 'Trace ' + ticket.id, exact: true }).click();
    await expect(page.locator(".inspector h2")).toContainText(
      title + " during disconnect",
    );
    await call("ticket.update", {
      id: ticket.id,
      title: title + " watched trace",
    });
    await expect(page.locator(".inspector h2")).toContainText(
      title + " watched trace",
    );
    await expect(page.locator('.ticket-graph')).toContainText(title + ' watched trace');
    const child = await call('ticket.create', { title: 'Implementation child', parent_id: ticket.id });
    related.push(child.id);
    const dependency = await call('ticket.create', {
      title: 'Required evidence',
      meta: { github: { repo: 'acme/widget', pr: 42 }, linear: { workspace: 'acme', issue: 'ENG-2' } },
    });
    related.push(dependency.id);
    await call('edge.add', { from: child.id, to: dependency.id, kind: 'waits_for' });
    const graph = page.getByRole('region', { name: 'Ticket dependency graph', exact: true });
    await expect(graph.locator('.graph-card')).toHaveCount(3);
    await expect(graph.locator('svg > path')).toHaveCount(2);
    const dependencyCard = graph.locator(`[data-ticket="${dependency.id}"]`);
    await expect(dependencyCard.getByRole('link', { name: 'ENG-2', exact: true })).toHaveAttribute('href', 'https://linear.app/acme/issue/ENG-2');
    await expect(dependencyCard.getByRole('link', { name: 'acme/widget PR #42', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/42');
    await graph.getByRole('button', { name: 'Inspect ' + child.id, exact: true }).click();
    await expect(page.locator('.inspector h2')).toContainText(child.id);
    const relationships = page.getByRole('table', { name: 'Ticket relationships', exact: true });
    await expect(relationships).toContainText('child of');
    await expect(relationships).toContainText('waits for');
    await expect(page.getByRole('textbox', { name: 'Trace reference', exact: true })).toHaveValue(ticket.id);
    await relationships.getByRole('button', { name: new RegExp(dependency.id) }).click();
    await expect(page.locator('.inspector h2')).toContainText(dependency.id);
    await expect(page.locator('.inspector details').filter({ hasText: 'Raw ticket data' })).not.toHaveAttribute('open');
    await call('comment.add', { id: dependency.id, body: `Plain ENG-2 and ${child.id}. [Linear issue](https://linear.app/acme/issue/ENG-2), [GitHub evidence](https://github.com/acme/widget/issues/9), [Implementation](jobshop://ticket/${child.id}). [Unsafe](javascript:alert%281%29) [Unknown](jobshop://command/delete) [Invalid](jobshop://ticket/%2Fetc).` });
    const comment = page.locator('.ticket-comment .markdown');
    await expect(comment.getByRole('link', { name: 'Linear issue', exact: true })).toHaveAttribute('href', 'https://linear.app/acme/issue/ENG-2');
    await expect(comment.getByRole('link', { name: 'GitHub evidence', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/issues/9');
    await expect(comment.getByRole('link', { name: 'Implementation', exact: true })).toHaveAttribute('href', '?view=trace&ref=' + child.id);
    await expect(comment.getByRole('link')).toHaveCount(3);
    await comment.getByRole('link', { name: 'Implementation', exact: true }).click();
    await expect(page.getByRole('textbox', { name: 'Trace reference', exact: true })).toHaveValue(child.id);
    await expect(page.locator('.inspector h2')).toContainText(child.id);
    await page.getByRole('table', { name: 'Ticket relationships', exact: true }).getByRole('button', { name: new RegExp(dependency.id) }).click();
    await call('ticket.close', { id: dependency.id });
    related.splice(related.indexOf(dependency.id), 1);
    await expect(dependencyCard.locator('.graph-state')).toContainText('closed');
    await expect(page.locator('.inspector')).toContainText('closed');
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
    if (isMobile) {
      await page.getByRole('navigation', { name: 'Mobile', exact: true }).getByRole('button', { name: /^Waiting/ }).click();
    } else {
      await page.getByRole('navigation', { name: 'Main', exact: true }).getByRole('button', { name: 'Home', exact: true }).click();
    }
    await call("ticket.close", { id: ticket.id });
    ticket = null;
    await expect(waiting).not.toContainText(title);
    expect(reads).toEqual([]);
  } finally {
    await context.setOffline(false);
    if (ticket) await call("ticket.close", { id: ticket.id });
    for (const id of related) await call('ticket.close', { id });
  }
});
