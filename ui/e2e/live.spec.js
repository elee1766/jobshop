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
    await page
      .getByRole("navigation", { name: "Main", exact: true })
      .getByRole("button", { name: "Trace", exact: true })
      .click();
    await page
      .getByRole("textbox", { name: "Trace reference" })
      .fill(ticket.id);
    await page.getByRole("button", { name: "Open", exact: true }).click();
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
    await expect(page.locator(".trace-tree")).toContainText(
      title + " watched trace",
    );
    await page
      .getByRole("navigation", { name: "Main", exact: true })
      .getByRole("button", { name: "Home", exact: true })
      .click();
    await call("ticket.close", { id: ticket.id });
    ticket = null;
    await expect(waiting).not.toContainText(title);
    expect(reads).toEqual([]);
  } finally {
    await context.setOffline(false);
    if (ticket) await call("ticket.close", { id: ticket.id });
  }
});
