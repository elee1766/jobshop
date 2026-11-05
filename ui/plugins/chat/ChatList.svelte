<script>
  // Chat view (#535): the chat pool's sessions, a form for a new one, and
  // the controller's escalation chats. A row opens the session panel.
  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();

  /** @type {any} */
  let chats = $state(null);
  /** @type {any[] | null} */
  let escalations = $state(null);
  let error = $state('');
  let first = $state('');
  let starting = $state(false);
  let sel = $state('');

  async function load() {
    try {
      [chats, escalations] = await Promise.all([api.get('api/chat'), api.get('api/escalations')]);
      error = '';
    } catch (e) {
      error = String(/** @type {any} */ (e).message || e);
    }
  }
  $effect(() => {
    sel = api.params().session || '';
    load();
    let t = 0;
    const off = api.subscribe(() => {
      sel = api.params().session || '';
      if (!t) t = window.setTimeout(() => { t = 0; load(); }, 1000);
    });
    return () => { off(); clearTimeout(t); };
  });

  const others = $derived((chats?.sessions ?? []).filter((/** @type {any} */ s) => !s.key.startsWith('escalation:')));
  /** @type {any[]} */
  const open = $derived((escalations ?? []).filter((/** @type {any} */ e) => e.open));
  /** @type {any[]} */
  const closed = $derived((escalations ?? []).filter((/** @type {any} */ e) => !e.open));

  /** @param {string} k */
  function label(k) {
    if (k.startsWith('attention:')) return 'Resolve ' + k.slice(10);
    if (k.startsWith('desk:')) return 'Desk chat';
    return k.split(':')[0] || 'Session';
  }
  /** @type {Record<string, string>} */
  const causes = { run: 'run failed', pool: 'pool quarantined', wait: 'wait timed out', guard: 'needs approval', label: 'needs human' };

  /** @param {string} t */
  function ago(t) {
    if (!t || t.startsWith('0001')) return '';
    const s = Math.round((Date.now() - new Date(t).getTime()) / 1000);
    if (s < 60) return s + 's ago';
    if (s < 3600) return Math.round(s / 60) + 'm ago';
    if (s < 86400) return Math.round(s / 3600) + 'h ago';
    return Math.round(s / 86400) + 'd ago';
  }

  /** @param {string} id @param {string} [item] */
  function pick(id, item) {
    sel = id;
    api.openSession(id, item ? { item } : undefined);
  }

  async function start() {
    if (!first.trim()) return;
    starting = true;
    error = '';
    try {
      const r = await api.post('api/chat/new', { text: first });
      first = '';
      pick(r.session);
    } catch (e) {
      error = String(/** @type {any} */ (e).message || e);
    } finally {
      starting = false;
    }
  }
</script>

<header class="vh">
  <h1>Chat</h1>
  <span class="sub">pool <span class="mono">{chats?.pool ?? '…'}</span></span>
</header>
<div class="list">
  <form class="new" onsubmit={(e) => { e.preventDefault(); start(); }}>
    <textarea bind:value={first} rows="3" placeholder="What should the operator look at?" data-testid="chat-new-input"></textarea>
    <div class="row"><button type="submit" class="primary" disabled={starting || !first.trim()} data-testid="chat-new">New session</button></div>
  </form>
  {#if error}<div class="err">{error}</div>{/if}
  <div data-testid="chat-list">
    {#each others as s (s.id)}
      <a href={'?view=chat&session=' + s.id} class="item" class:active={sel === s.id} onclick={(e) => { e.preventDefault(); pick(s.id, s.key.startsWith('attention:') ? s.key.slice(10) : ''); }}>
        <span class="dot" class:accent={s.alive}></span>
        <span class="name">{label(s.key)}</span>
        <span class="mono faint id">{s.id}</span>
      </a>
    {:else}
      <div class="faint pad0">No open operator sessions.</div>
    {/each}
  </div>

  <h2 class="st">Escalations</h2>
  {#if escalations && open.length + closed.length === 0}
    <div class="card empty" data-testid="escalations-empty"><div class="muted">No escalation sessions.</div></div>
  {:else if escalations}
    <div class="card scroll">
      <table class="stack" data-testid="escalations">
        <thead><tr><th>Ticket</th><th>Cause</th><th>Why</th><th>State</th><th>Age</th></tr></thead>
        <tbody>
          {#each [...open, ...closed] as e (e.ticket)}
            <tr class="click" class:sel={sel === e.session} onclick={() => pick(e.session, 'ticket:' + e.ticket)} data-testid="escalation-row">
              <td class="c-title">{e.title}<div class="faint mono">{e.ticket}{#if e.held && e.held !== e.ticket} · {e.held}{/if}</div></td>
              <td class="c-st"><span class="pill {e.open ? 'warn' : ''}">{causes[e.cause] ?? e.cause ?? 'escalation'}</span></td>
              <td class="muted full">{e.why}</td>
              <td class="c-st"><span class="pill {e.open ? 'accent' : 'ok'}">{e.open ? 'open' : 'closed' + (e.resolution ? ' · ' + e.resolution : '')}</span></td>
              <td class="muted c-time">{ago(e.open ? e.since : e.closed)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .list { flex: 1; min-height: 0; overflow-y: auto; padding: 12px 20px 24px; max-width: 960px; width: 100%; box-sizing: border-box; }
  .item { display: flex; align-items: center; gap: 8px; height: 34px; padding: 0 8px; border-radius: var(--desk-radius); color: var(--desk-muted); min-width: 0; text-decoration: none; }
  .item:hover { background: var(--desk-hover); color: var(--desk-text); }
  .item.active { background: var(--desk-raised); color: var(--desk-text); }
  .name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .id { flex: none; font-size: 11px; }
  .pad0 { padding: 8px; }
  .new { display: flex; flex-direction: column; gap: 6px; padding: 0 0 12px; }
  .new textarea { resize: vertical; }
  .new .row { display: flex; justify-content: flex-end; }
  .err { color: var(--desk-bad); padding: 4px 2px; font-size: 12px; }
  @media (max-width: 768px) {
    .list { padding: 8px 8px 16px; }
    .item { height: 48px; padding: 0 10px; font-size: 14px; border-bottom: 1px solid var(--desk-border); border-radius: 0; }
    .vh .sub { display: none; }
  }
</style>
