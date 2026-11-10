// The trace page shows the timeline or the details, not both (#1181). Specs
// that read the details switch to them with the page's own button.
export async function showDetails(page) {
  const button = page.getByRole('button', { name: /^Show details of/ });
  const view = page.getByRole('region', { name: 'Details view', exact: true });
  await button.or(view).first().waitFor();
  if (await button.isVisible()) await button.click();
  await view.waitFor();
}

export async function showTimeline(page) {
  const button = page.getByRole('button', { name: 'Show the timeline', exact: true });
  const chart = page.getByRole('region', { name: 'Ticket timeline', exact: true });
  await button.or(chart).first().waitFor();
  if (await button.isVisible()) await button.click();
  await chart.waitFor({ state: 'visible' });
}
