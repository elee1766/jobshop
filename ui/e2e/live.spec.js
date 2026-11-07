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
      /\/api\/(overview|system|improvements|chat)(\?|$)|\/api\/rpc\/attention$/.test(
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
    await context.setOffline(true);
    await expect(page.locator(".status")).not.toContainText("sys ok");
    await call("ticket.update", {
      id: ticket.id,
      title: title + " during disconnect",
    });
    await context.setOffline(false);
    await expect(waiting).toContainText(title + " during disconnect");
    await call("ticket.close", { id: ticket.id });
    ticket = null;
    await expect(waiting).not.toContainText(title);
    expect(reads).toEqual([]);
  } finally {
    await context.setOffline(false);
    if (ticket) await call("ticket.close", { id: ticket.id });
  }
});
