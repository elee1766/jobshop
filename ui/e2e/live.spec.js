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
    await expect(page.locator(".ticket-overview > h2")).toContainText(
      title + " during disconnect",
    );
    await call("ticket.update", {
      id: ticket.id,
      title: title + " watched trace",
    });
    await expect(page.locator(".ticket-overview > h2")).toContainText(
      title + " watched trace",
    );
    await expect(page.locator('.ticket-graph')).toContainText(title + ' watched trace');
    const dependency = await call('ticket.create', {
      title: 'Required evidence',
      meta: { github: { repo: 'acme/widget', pr: 42 }, linear: { workspace: 'acme', issue: 'ENG-2' } },
    });
    related.push(dependency.id);
    const child = await call('ticket.create', { title: 'Implementation child', parent_id: ticket.id });
    related.push(child.id);
    await call('ticket.update', { id: child.id, fields: { wait: { reason: 'Investigating https://github.com/acme/widget/issues/9. acme/widget#42 ENG-2 [Unsafe](javascript:alert%281%29)', conds: [{ type: 'run', run: ticket.id, state: 'closed' }] } } });
    await call('edge.add', { from: child.id, to: dependency.id, kind: 'waits_for' });
    const graph = page.getByRole('region', { name: 'Ticket dependency graph', exact: true });
    await expect(graph.locator('.graph-card')).toHaveCount(3);
    await expect(graph.locator('svg > path')).toHaveCount(2);
    const temporalNodes = await graph.locator('.graph-card').evaluateAll(nodes => nodes.map(node => ({ id: node.dataset.ticket, x: parseFloat(node.style.left), created: Date.parse(node.querySelector('time').dateTime) })).sort((a, b) => a.created - b.created));
    expect(temporalNodes.map(node => node.id)).toEqual([ticket.id, dependency.id, child.id]);
    expect(temporalNodes.every((node, i) => !i || temporalNodes[i - 1].x <= node.x)).toBeTruthy();
    await expect(graph.locator('.graph-axis .graph-stamp-start')).toHaveCount(1);
    await expect(graph.locator('.graph-axis .graph-stamp-now')).toHaveCount(1);
    const viewport = graph.getByRole('application', { name: 'Ticket graph navigation', exact: true });
    const canvas = graph.locator('.graph-canvas');
    const camera = () => canvas.evaluate(node => node.style.transform);
    const initialCamera = await camera();
    const zoom = graph.getByLabel('Graph zoom level', { exact: true });
    const initialZoom = await zoom.textContent();
    await graph.getByRole('button', { name: 'Zoom in', exact: true }).click();
    await expect(zoom).not.toHaveText(initialZoom);
    await graph.getByRole('button', { name: 'reset', exact: true }).click();
    await expect.poll(camera).toBe(initialCamera);
    await viewport.focus();
    await viewport.press('ArrowRight');
    await expect.poll(camera).not.toBe(initialCamera);
    await viewport.press('0');
    await expect.poll(camera).toBe(initialCamera);
    const box = await viewport.boundingBox();
    await page.mouse.move(box.x + box.width / 2, box.y + box.height - 12);
    await page.mouse.down();
    await page.mouse.move(box.x + box.width / 2 + 30, box.y + box.height - 32, { steps: 4 });
    await expect(viewport).toHaveCSS('cursor', 'grabbing');
    await page.mouse.up();
    await expect.poll(camera).not.toBe(initialCamera);
    await expect(viewport).toHaveCSS('cursor', 'grab');
    await page.mouse.wheel(0, -150);
    await expect(zoom).not.toHaveText(initialZoom);
    await graph.getByRole('button', { name: 'reset', exact: true }).click();
    await expect.poll(camera).toBe(initialCamera);
    if (isMobile) {
      const touch = await context.newCDPSession(page);
      const left = box.x + box.width / 3;
      const right = box.x + box.width * 2 / 3;
      const y = box.y + box.height - 12;
      await touch.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: left, y, id: 1 }, { x: right, y, id: 2 }] });
      await touch.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: left - 20, y, id: 1 }, { x: right + 20, y, id: 2 }] });
      await touch.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
      await touch.detach();
      await expect(zoom).not.toHaveText(initialZoom);
      await graph.getByRole('button', { name: 'reset', exact: true }).click();
      await expect.poll(camera).toBe(initialCamera);
    }
    const dependencyCard = graph.locator(`[data-ticket="${dependency.id}"]`);
    await expect(dependencyCard.getByRole('link', { name: 'ENG-2', exact: true })).toHaveAttribute('href', 'https://linear.app/acme/issue/ENG-2');
    await expect(dependencyCard.getByRole('link', { name: 'acme/widget PR #42', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/42');
    await graph.getByRole('button', { name: 'Inspect ' + child.id, exact: true }).click();
    await expect(page.locator('.ticket-overview > h2')).toContainText(child.id);
    const statusLink = graph.locator(`[data-ticket="${child.id}"] .graph-state`).getByRole('link', { name: 'https://github.com/acme/widget/issues/9', exact: true });
    await expect(statusLink).toHaveAttribute('href', 'https://github.com/acme/widget/issues/9');
    await expect(statusLink).toHaveCSS('text-decoration-line', 'underline');
    const linkBackground = await statusLink.evaluate(node => getComputedStyle(node).backgroundColor);
    await statusLink.hover();
    await expect(statusLink).not.toHaveCSS('background-color', linkBackground);
    await expect(graph.locator(`[data-ticket="${child.id}"] .graph-state`).getByRole('link', { name: 'acme/widget#42', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/42');
    await expect(graph.locator(`[data-ticket="${child.id}"] .graph-state`).getByRole('link')).toHaveCount(2);
    const relationships = page.getByLabel('Ticket relationships', { exact: true });
    const dependencyRow = relationships.getByRole('listitem').filter({ hasText: 'Required evidence' });
    await expect(dependencyRow.locator('.ticket-status')).toHaveText('O');
    await expect(dependencyRow.locator('.ticket-status')).toHaveAttribute('title', 'open');
    await expect(dependencyRow.locator('.ticket-label > span')).toHaveText('Required evidence');
    await expect(dependencyRow.locator('.ticket-label > small')).toHaveText(dependency.id);
    await expect(page.locator('.ticket-parent')).toContainText('child of');
    await expect(relationships.getByRole('region', { name: 'waits for', exact: true })).toContainText('Required evidence');
    await expect(page.getByRole('textbox', { name: 'Trace reference', exact: true })).toHaveValue(ticket.id);
    const history = page.locator('.ticket-history');
    await expect(history.locator(':scope > summary')).toHaveCSS('cursor', 'pointer');
    await expect(page.locator('.inspector summary').filter({ hasText: 'raw ticket data' })).toHaveCSS('cursor', 'pointer');
    await expect(history).toHaveAttribute('open');
    await expect(history.locator('.history-raw').first()).not.toHaveAttribute('open');
    await expect(history.getByRole('link', { name: 'https://github.com/acme/widget/issues/9', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/issues/9');
    const historyLink = history.getByRole('link', { name: new RegExp(dependency.id) });
    await expect(historyLink).toHaveCSS('cursor', 'pointer');
    await historyLink.click();
    await expect(page.locator('.ticket-overview > h2')).toContainText(dependency.id);
    await graph.getByRole('button', { name: 'Inspect ' + child.id, exact: true }).click();
    await relationships.getByRole('button', { name: new RegExp(dependency.id) }).click();
    await expect(page.locator('.ticket-overview > h2')).toContainText(dependency.id);
    await expect(page.locator('.inspector details').filter({ hasText: 'raw ticket data' })).not.toHaveAttribute('open');
    await call('comment.add', { id: dependency.id, body: `Fallback (acme/widget#42). Plain widget#42 #42 ENG-2 and ${child.id}. \`acme/widget#43\` <code>acme/widget#44</code>\n\n\`\`\`\nacme/widget#45\n\`\`\`\n\n[acme/widget#46](https://github.com/acme/widget/issues/46), [Linear issue](https://linear.app/acme/issue/ENG-2), [GitHub evidence](https://github.com/acme/widget/issues/9), [Implementation](jobshop://ticket/${child.id}). [Unsafe](javascript:alert%281%29) [Unknown](jobshop://command/delete) [Invalid](jobshop://ticket/%2Fetc).\n\nCompact https://github.com/elee1766/jobcity/pull/21?w=1#top, <https://linear.app/gfx-labs/issue/JOB-93/stall-triggers>. \`https://github.com/acme/widget/pull/47\` https://evilgithub.com/acme/widget/pull/48 [Authored PR](https://github.com/acme/widget/pull/49)` });
    const comment = page.locator('.ticket-comment .markdown');
    await expect(comment.getByRole('link', { name: 'Linear issue', exact: true })).toHaveAttribute('href', 'https://linear.app/acme/issue/ENG-2');
    await expect(comment.getByRole('link', { name: 'GitHub evidence', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/issues/9');
    await expect(comment.getByRole('link', { name: 'Implementation', exact: true })).toHaveAttribute('href', '?view=trace&ref=' + child.id);
    await expect(comment.getByRole('link', { name: 'acme/widget#42', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/42');
    await expect(comment.getByRole('link', { name: 'acme/widget#42', exact: true })).toHaveAttribute('rel', 'noopener noreferrer');
    await expect(comment.getByRole('link', { name: 'acme/widget#46', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/issues/46');
    await expect(comment.locator('code a')).toHaveCount(0);
    const compactPR = comment.getByRole('link', { name: 'PR:elee1766/jobcity#21', exact: true });
    await expect(compactPR).toHaveAttribute('href', 'https://github.com/elee1766/jobcity/pull/21?w=1#top');
    await expect(compactPR).toHaveCSS('text-decoration-line', 'underline');
    await expect(comment.getByRole('link', { name: 'gfx-labs@JOB-93', exact: true })).toHaveAttribute('href', 'https://linear.app/gfx-labs/issue/JOB-93/stall-triggers');
    await expect(comment.getByRole('link', { name: 'https://evilgithub.com/acme/widget/pull/48', exact: true })).toBeVisible();
    await expect(comment.getByRole('link', { name: 'Authored PR', exact: true })).toHaveAttribute('href', 'https://github.com/acme/widget/pull/49');
    await expect(comment.getByRole('link')).toHaveCount(9);
    await comment.getByRole('link', { name: 'Implementation', exact: true }).click();
    await expect(page.getByRole('textbox', { name: 'Trace reference', exact: true })).toHaveValue(child.id);
    await expect(page.locator('.ticket-overview > h2')).toContainText(child.id);
    await page.getByLabel('Ticket relationships', { exact: true }).getByRole('button', { name: new RegExp(dependency.id) }).click();
    await call('ticket.close', { id: dependency.id });
    related.splice(related.indexOf(dependency.id), 1);
    await expect(dependencyCard.locator('.ticket-status')).toHaveAttribute('aria-label', 'closed · done');
    await expect(dependencyCard.locator('.graph-state')).toContainText('done');
    await expect(page.locator('.ticket-overview .ticket-status')).toHaveAttribute('aria-label', 'closed · done');
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
    if (isMobile) {
      await page.getByRole('navigation', { name: 'Mobile', exact: true }).getByRole('button', { name: /^waiting/ }).click();
    } else {
      await page.getByRole('navigation', { name: 'Main', exact: true }).getByRole('button', { name: 'home', exact: true }).click();
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
