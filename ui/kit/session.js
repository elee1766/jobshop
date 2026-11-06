// Closing a session from Desk (JOB-111), shared by the Sessions rows and
// the chat pane. A dry run lists what the session holds so the confirm
// says where each ticket goes; nothing is orphaned by the close.

/**
 * Asks, then closes session id. True when it closed.
 * @param {any} api @param {string} id
 */
export async function closeSessionFlow(api, id) {
  /** @type {{ held: { id: string, title: string, status: string, pool?: string }[] }} */
  let dry;
  try {
    dry = await api.call('session.close', { id, dry_run: true });
  } catch (e) {
    api.toast('Close failed: ' + String(/** @type {any} */ (e).message || e));
    return false;
  }
  if (!(await api.confirm(closeQuestion(id, dry.held || [])))) return false;
  try {
    await api.call('session.close', { id, reason: 'closed in Desk' });
    api.toast('Session ' + id + ' closed');
    return true;
  } catch (e) {
    api.toast('Close failed: ' + String(/** @type {any} */ (e).message || e));
    return false;
  }
}

/**
 * The confirm text: what the close does to each held ticket.
 * @param {string} id @param {{ id: string, title: string, status: string, pool?: string }[]} held
 */
export function closeQuestion(id, held) {
  let q = 'Close session ' + id + '? A running turn is cancelled and its slot is freed. The transcript stays readable.';
  if (!held.length) return q + '\nIt holds no tickets.';
  q += '\nIt holds:';
  for (const t of held) {
    const where = t.status === 'in_progress' ? (t.pool ? 'handed back to ' + t.pool : 'released') : 'unassigned';
    q += '\n- ' + t.id + ' ' + t.title + ' (' + where + ')';
  }
  return q;
}
