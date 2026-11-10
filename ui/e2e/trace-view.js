// The TL page keeps the timeline on top and the selection's panel under it
// (#1183). On a phone that panel is a full-screen popup over the chart. Specs
// that read the panel call showDetails and specs that read the chart call
// showTimeline, so each works at any width.
export async function showDetails(page) {
  const lower = page.getByLabel('Ticket details', { exact: true });
  const picked = page.locator('.tl-row.tl-picked .tl-name').first();
  await lower.or(picked).first().waitFor();
  if (!(await lower.isVisible())) await picked.click();
  await lower.waitFor();
}

export async function showTimeline(page) {
  const close = page.getByRole('button', { name: 'Close and return to the timeline', exact: true });
  const chart = page.getByRole('region', { name: 'Ticket timeline', exact: true });
  await close.or(chart).first().waitFor();
  if (await close.isVisible()) await close.click();
  await chart.waitFor({ state: 'visible' });
}

// Opens one tab of the panel under the chart.
export async function openTab(page, name) {
  await showDetails(page);
  await page.getByRole('tab', { name, exact: true }).click();
}
