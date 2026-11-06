<script>
  // Built-in Attention view (#535): what waits on a human, a push opt-in,
  // and runs in progress. A row shows a summary and its canned resolve
  // buttons (#607); Open chat (or a click on the row) opens the session
  // chat that resolves the item (POST api/attention/chat). Uses api only
  // (#517).
  import { resolveActions, runAction, pendingItems, markPending, clearPending, onPending } from '../../kit/attention.js';
  import { safe } from '../../kit/format.js';
  import Holder from '../../kit/Holder.svelte';
  import { untrack } from 'svelte';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();

  /** @type {any[] | null} */
  let items = $state(null);
  let error = $state('');
  /** @type {{ run: any, steps: any[], comment: any }[]} */
  let progress = $state([]);
  /** Holder sessions of in-progress steps, by id (JOB-54). @type {Record<string, any>} */
  let holders = $state({});
  let progressLoaded = $state(false);
  let opening = $state('');
  let push = $state({ state: 'off', error: '' });

  /** Rows whose action succeeded, waiting for a load without them
   * (JOB-60). loads counts loads started after the action that still
   * listed the item. @type {Record<string, { label: string, loads: number }>} */
  let pending = $state({});
  /** Rows a refresh still listed after their action. @type {Record<string, string>} */
  let stuck = $state({});
  // Loads that may still list the item before it counts as a no-op: the
  // first can race the action's commit.
  const pendingLoads = 2;
  let retry = 0;

  async function load() {
    const asked = Object.keys(pending);
    try {
      const got = await api.get('api/attention');
      const listed = new Set(got.map((/** @type {any} */ x) => x.kind + ':' + x.id));
      let again = false;
      for (const key of asked) {
        const p = pending[key];
        if (!p) continue;
        if (!listed.has(key)) {
          delete pending[key];
          clearPending(key);
        } else if (++p.loads >= pendingLoads) {
          stuck[key] = p.label;
          delete pending[key];
          clearPending(key);
        } else again = true;
      }
      items = got;
      // A pending row gets another refresh soon instead of the next event.
      if (again && !retry) retry = window.setTimeout(() => { retry = 0; load(); }, 1500);
      error = '';
    } catch (e) {
      error = String(/** @type {any} */ (e).message || e);
    }
    try {
      // One request: steps, last comment and holder sessions come with
      // the runs (JOB-54).
      const w = await api.get('api/work?state=running&limit=6&detail=1');
      const d = w.detail ?? {};
      holders = w.sessions ?? {};
      progress = (w.runs ?? []).slice(0, 6).map((/** @type {any} */ run) => ({
        run, steps: d[run.id]?.steps ?? [], comment: d[run.id]?.comment ?? null,
      }));
    } catch { /* progress is a side list */ }
    progressLoaded = true;
  }

  /** @param {string} key @param {string} label */
  function setPending(key, label) {
    delete stuck[key];
    pending[key] = { label, loads: 0 };
  }

  $effect(() => untrack(() => {
    // Actions taken in the chat panel mark the row too.
    for (const [key, label] of pendingItems()) setPending(key, label);
    const offPending = onPending((key, label) => { setPending(key, label); load(); });
    load();
    api.push.status().then((/** @type {any} */ s) => (push = s));
    let t = 0;
    const off = api.subscribe(() => {
      if (!t) t = window.setTimeout(() => { t = 0; load(); }, 1000);
    });
    // A push notification opens Desk at ?item=<kind:id> (#535).
    const p = api.params();
    if (p.item && !p.session) {
      const [kind, ...rest] = p.item.split(':');
      open({ kind, id: rest.join(':') });
    }
    return () => { off(); offPending(); clearTimeout(t); clearTimeout(retry); retry = 0; };
  }));

  /** @param {{ kind: string, id: string, title?: string, why?: string }} it */
  async function open(it) {
    const item = it.kind + ':' + it.id;
    opening = item;
    try {
      const c = await api.post('api/attention/chat', { kind: it.kind, id: it.id, title: it.title || '', why: it.why || '' });
      api.openSession(c.session, { item });
    } catch (e) {
      api.toast('Could not open a chat for ' + item + ': ' + String(/** @type {any} */ (e).message || e));
      if (['run', 'ticket', 'question', 'decision', 'stale', 'wait'].includes(it.kind)) api.openTicket(it.id);
    } finally {
      opening = '';
    }
  }

  let acting = $state('');
  /** A canned button: confirm, resolve, drop the row; no session opens.
   * @param {any} it @param {import('../../kit/attention.js').Action} a */
  async function resolve(it, a) {
    const key = it.kind + ':' + it.id;
    acting = key;
    try {
      // markPending notifies onPending above, which refreshes at once.
      if (await runAction(api, a)) markPending(key, a.label);
    } finally {
      acting = '';
    }
  }

  /** @type {Record<string, string>} */
  const pushTitle = {
    unsupported: 'This browser has no Web Push (on iOS, add Desk to the home screen first).',
    unavailable: 'Desk runs without --push.',
    denied: 'Notifications are blocked in the browser settings.',
    off: 'Get a notification for each new attention item.',
    on: 'Notifications are on for this browser.',
  };
  async function togglePush() {
    push = push.state === 'on' ? await api.push.unsubscribe() : await api.push.subscribe();
  }
  async function testPush() {
    push = await api.push.test();
    if (!push.error) api.toast('Test notification sent');
  }

  /** @param {string} t */
  function ago(t) {
    if (!t || t.startsWith('0001')) return '';
    const s = Math.round((Date.now() - new Date(t).getTime()) / 1000);
    if (s < 60) return s + 's ago';
    if (s < 3600) return Math.round(s / 60) + 'm ago';
    if (s < 86400) return Math.round(s / 3600) + 'h ago';
    return Math.round(s / 86400) + 'd ago';
  }
  /** @param {any} t @param {string} k */
  const parse = (t, k) => (typeof t?.[k] === 'string' ? JSON.parse(t[k] || '{}') : t?.[k] || {});
  /** @param {any[]} steps */
  function reported(steps) {
    return steps
      .map((s) => ({ s, r: parse(s, 'meta').report }))
      .filter((x) => x.r && (x.r.note || x.r.state))
      .sort((a, b) => String(b.r.at).localeCompare(String(a.r.at)));
  }
  /** @param {string} s */
  const rtone = (s) => (s === 'failed' || s === 'blocked' ? 'bad' : s === 'waiting' || s === 'dropped' ? 'warn' : s === 'ready' ? 'ok' : 'accent');
  const isSession = (/** @type {string} */ s) => !!s && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(s);
  /** The holder session of an in-progress step. @param {any} s */
  const holderOf = (s) => (s?.status === 'in_progress' && s.holder ? holders[s.holder] ?? { id: s.holder } : null);
  /** @param {string} id */
  const watch = (id) => api.openSession(id, { readonly: true });
</script>

<header class="vh">
  <h1>Attention</h1>
  <span class="sub">What waits on a human, oldest first.</span>
  <span class="actions">
    {#if push.state === 'on' || push.state === 'off'}
      {#if push.state === 'on'}<button class="ghost" onclick={testPush} data-testid="push-test">Test</button>{/if}
      <button class:on={push.state === 'on'} onclick={togglePush} title={pushTitle[push.state]} data-testid="push-toggle">{push.state === 'on' ? 'Notifications on' : 'Notify me'}</button>
    {:else}
      <span class="faint pushna" title={pushTitle[push.state]} data-testid="push-state">{push.state === 'unavailable' ? '' : 'no notifications'}</span>
    {/if}
  </span>
</header>
<div class="pad">
  {#if push.error}<div class="card err">{push.error}</div>{/if}
  {#if error}
    <div class="card err">{error}</div>
  {:else if !items}
    <p class="faint">Loading…</p>
  {:else if items.length === 0}
    <div class="card empty" data-testid="attention-empty">
      <div class="big">Nothing needs you.</div>
      <div class="muted">No escalations, questions, decisions or other items waiting on you.</div>
    </div>
  {:else}
    <div class="card scroll">
      <table class="stack" data-testid="attention-list">
        <thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead>
        <tbody>
          {#each items as it (it.kind + it.id)}
            {@const key = it.kind + ':' + it.id}
            {@const pend = pending[key]}
            <tr class="click" class:busy={opening === key || !!pend} class:pending={!!pend} onclick={() => open(it)} data-testid="attention-row" data-item={it.kind + ':' + it.id}>
              <td class="c-id"><span class="pill warn">{it.kind}</span></td>
              <td class="c-title">
                {it.title || it.id}<div class="faint mono">{it.id}{#if it.ref} · {#if it.url && safe(it.url)}<a href={it.url} target="_blank" rel="noopener noreferrer" onclick={(e) => e.stopPropagation()} data-testid="attention-ref">{it.ref}</a>{:else}{it.ref}{/if}{/if}</div>
                {#if it.summary}<div class="summary muted" data-testid="attention-summary">{it.summary}</div>{/if}
                <!-- Buttons do not open the chat: the row's click is for that. -->
                <div class="acts" role="presentation" onclick={(e) => e.stopPropagation()}>
                  {#if pend}
                    <span class="muted" data-testid="row-pending">{pend.label}…</span>
                  {:else}
                    {#if stuck[key]}<span class="warn-text" data-testid="row-stuck">still listed after {stuck[key]}</span>{/if}
                    {#each resolveActions(api, it.kind, it.id) as a (a.label)}
                      <button type="button" disabled={acting === key} onclick={() => resolve(it, a)} data-testid="row-resolve">{a.label}</button>
                    {/each}
                    <button type="button" class="ghost" disabled={opening === key} onclick={() => open(it)} data-testid="open-chat">Open chat</button>
                  {/if}
                </div>
              </td>
              <td class="muted full">{it.why}</td>
              <td class="muted c-time">{ago(it.since)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <h2 class="st">In progress</h2>
  {#if progress.length}
    <div class="progress" data-testid="progress">
      {#each progress as { run, steps, comment } (run.id)}
        {@const rep = reported(steps)}
        {@const active = steps.find((s) => s.status === 'in_progress')}
        {@const ah = holderOf(active)}
        <div class="item card" data-testid="progress-card" data-run={run.id}>
          <button class="head" onclick={() => api.openTicket(run.id)}>
            <span class="dot accent"></span>
            <span class="title">{run.title}</span>
            <span class="mono faint id">{run.id}</span>
            <span class="faint when">{ago(run.updated_at)}</span>
          </button>
          <div class="who">
            {#if ah}
              <span class="step muted">{parse(active, 'fields').step || active.title}</span>
              <Holder h={ah} onopen={watch} />
            {:else}
              <span class="faint" data-testid="no-holder">waiting for a session</span>
            {/if}
          </div>
          <ul>
            {#each rep.slice(0, 3) as { s, r } (s.id)}
              {@const sh = holderOf(s)}
              <li>
                <span class="pill {rtone(r.state)}">{r.state}</span>
                <span class="step muted">{parse(s, 'fields').step || s.title}</span>
                {#if sh}<Holder h={sh} onopen={watch} />{:else if isSession(r.by)}<a href={'?session=' + r.by} class="mono" onclick={(e) => { e.preventDefault(); watch(r.by); }}>{r.by}</a>{/if}
                {#if r.note}<span class="note">{r.note}</span>{/if}
                <span class="faint when">{ago(r.at)}</span>
              </li>
            {:else}
              {#if active}
                <li><span class="pill">{active.status}</span><span class="step muted">{parse(active, 'fields').step || active.title}</span><span class="faint">no report yet</span></li>
              {:else}
                <li><span class="faint">waiting for a step to start</span></li>
              {/if}
            {/each}
            {#if comment}
              <li><span class="pill">comment</span><span class="step muted mono">{#if isSession(comment.author)}<a href={'?session=' + comment.author} onclick={(e) => { e.preventDefault(); api.openSession(comment.author); }}>{comment.author}</a>{:else}{comment.author}{/if}</span><span class="note">{comment.body}</span><span class="faint when">{ago(comment.created_at)}</span></li>
            {/if}
          </ul>
        </div>
      {/each}
    </div>
  {:else if !progressLoaded}
    <div class="faint">Loading…</div>
  {:else}
    <div class="faint none" data-testid="progress-empty">No runs in progress.</div>
  {/if}
</div>

<style>
  tr.busy td { opacity: 0.6; }
  tr.pending { pointer-events: none; }
  .warn-text { color: var(--desk-warn, #b7791f); font-size: 12px; align-self: center; }
  .summary { margin-top: 4px; overflow-wrap: anywhere; display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
  .acts { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 6px; cursor: default; }
  .pushna { font-size: 12px; }
  .progress { display: flex; flex-direction: column; gap: 8px; }
  .item { overflow: hidden; }
  .who { display: flex; flex-wrap: wrap; gap: 6px 8px; align-items: baseline; padding: 0 12px 6px 28px; min-width: 0; }
  .head {
    width: 100%; height: auto; min-height: 32px; justify-content: flex-start; gap: 8px; padding: 6px 12px; border: 0; border-radius: 0;
    background: transparent; text-align: left; white-space: normal;
  }
  .head:hover:not(:disabled) { background: var(--desk-hover); }
  .title { flex: 1; min-width: 0; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .id { flex: none; }
  ul { list-style: none; margin: 0; padding: 0 12px 8px 28px; display: flex; flex-direction: column; gap: 4px; }
  li { display: flex; gap: 8px; align-items: baseline; min-width: 0; flex-wrap: wrap; }
  .step { flex: none; }
  .note { flex: 1 1 200px; min-width: 0; overflow-wrap: anywhere; white-space: pre-wrap; display: -webkit-box; -webkit-line-clamp: 3; line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
  .when { flex: none; margin-left: auto; font-size: 11px; }
  .none { padding: 4px 0; }
  @media (max-width: 768px) {
    .head { min-height: 44px; }
    .acts button { min-height: 40px; }
    .id { display: none; }
    li a { padding: 8px 0; }
  }
</style>
