import { test, expect } from "@playwright/test";

// This creates real attention items and shuts down their chats.
test.skip(!process.env.DESK_MUTATE, "set DESK_MUTATE=1 against an isolated controller with escalation and attention chats");

test("picker distinguishes managed chats and close differs from resolve", async ({ page, request, baseURL, isMobile }) => {
  const call = async (method, data) => {
    const response = await request.post(new URL("api/rpc/" + method, baseURL).href, {
      headers: { "X-Jobshop-Desk": "1" }, data,
    });
    expect(response.ok(), await response.text()).toBeTruthy();
    return response.json();
  };
  const sources = [];
  try {
    const title = "Escalation dismissal " + crypto.randomUUID();
    const source = await call("ticket.create", { title, labels: ["needs-human"] });
    sources.push(source.id);
    const question = await call("ticket.create", { title: "Picker question " + crypto.randomUUID(), labels: ["question"] });
    sources.push(question.id);
    await page.goto("./");
    if (isMobile) await page.getByRole("button", { name: "chats", exact: true }).click();
    const picker = page.getByRole("dialog", { name: "Choose chat" });
    await page.getByRole("button", { name: "Choose chat", exact: true }).click();
    const row = picker.getByRole("row").filter({ hasText: source.id });
    await expect(row).toContainText("escalation");
    await expect(picker.getByRole("row").filter({ hasText: question.id })).toContainText("attention");
    await row.getByRole("button").click();
    const actions = page.getByRole("toolbar", { name: "Chat actions" });
    await actions.getByRole("button", { name: "close", exact: true }).click();
    await expect(actions).toHaveCount(0);
    expect((await call("ticket.show", { id: source.id })).ticket.status).toBe("open");
    await page.getByRole("button", { name: "Choose chat", exact: true }).click();
    await expect(picker.getByRole("row").filter({ hasText: source.id })).toHaveCount(0);
    await picker.getByRole("button", { name: /^archived/ }).click();
    await picker.getByRole("row").filter({ hasText: source.id }).getByRole("button").click();
    await page.getByRole("button", { name: "restore", exact: true }).click();
    await expect(actions.getByRole("button", { name: "resolve", exact: true })).toBeEnabled();
    page.once("dialog", dialog => dialog.dismiss());
    await actions.getByRole("button", { name: "resolve", exact: true }).click();
    await expect(actions).toBeVisible();
    expect((await call("ticket.show", { id: source.id })).ticket.status).toBe("open");
    page.once("dialog", dialog => dialog.accept());
    await actions.getByRole("button", { name: "resolve", exact: true }).click();
    await expect(actions).toHaveCount(0);
    expect((await call("ticket.show", { id: source.id })).ticket.status).toBe("closed");
    await page.reload();
    if (isMobile) await page.getByRole("button", { name: "chats", exact: true }).click();
    await page.getByRole("button", { name: "Choose chat", exact: true }).click();
    await expect(picker.getByRole("table", { name: "Recent chats" })).toBeVisible();
    await expect(picker.getByRole("row").filter({ hasText: source.id })).toHaveCount(0);
  } finally {
    for (const id of sources) {
      const current = await call("ticket.show", { id });
      if (current.ticket.status !== "closed") await call("ticket.close", { id });
    }
  }
});
