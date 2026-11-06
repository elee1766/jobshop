<script>
  import { resource, params, errText } from '../../kit/plugin.svelte.js';
  import { ago } from '../../kit/format.js';
  import { closeSessionFlow } from '../../kit/session.js';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api } = $props();
  const url = params(() => api);
  const go = (/** @type {Record<string, string>} */ p) => api.go(p);
  const closeSession = () => api.closeSession();
  const sessions = resource(() => api, () => 'api/sessions');

  let actErr = $state('');

  /** @param {Event} e @param {any} body */
  async function cordon(e, body) {
    e.stopPropagation();
    actErr = '';
    try { await api.post('api/session/cordon', body); sessions.refresh(); } catch (err) { actErr = errText(err); }
  }

  /** @param {Event} e @param {string} id */
  async function close(e, id) {
    e.stopPropagation();
    actErr = '';
    if (await closeSessionFlow(api, id)) {
      if (url.session === id) closeSession();
      sessions.refresh();
    }
  }

  /** @param {any} s */
  function stateOf(s) {
    if (!s.alive) return ['dead', 'bad'];
    if (s.backoff_until) return ['backoff', 'warn'];
    if (s.drained) return ['drained', 'warn'];
    if (s.current) return ['working', 'ok'];
    return ['idle', ''];
  }
</script>

<header class="vh">
  <h1>Sessions</h1>
  {#if sessions.data}<span class="sub">{sessions.data.length} live</span>{/if}
</header>
<div class="pad">
  {#if sessions.error}
    <div class="card err">{sessions.error}</div>
  {:else if sessions.data}
    {#if actErr}<div class="card err">{actErr}</div>{/if}
    <div class="card scroll">
      <table class="stack" data-testid="sessions">
        <thead><tr><th>Session</th><th>State</th><th>Ticket</th><th>Pool</th><th>Model</th><th>Last event</th><th></th></tr></thead>
        <tbody>
          {#each sessions.data as s (s.id)}
            {@const st = stateOf(s)}
            <tr class="click" class:sel={url.session === s.id} onclick={() => (url.session === s.id ? closeSession() : go({ session: s.id }))}>
              <td class="c-id"><a href={'?session=' + s.id} class="mono sl" onclick={(e) => { e.preventDefault(); e.stopPropagation(); go({ session: s.id, ticket: '' }); }}>{s.id}</a>{#if s.display_name}<div class="muted">{s.display_name}</div>{/if}{#if s.crashes}<span class="pill bad gap">{s.crashes} crashes</span>{/if}</td>
              <td class="c-st"><span class="pill {st[1]}">{st[0]}</span>{#if s.cordon}<span class="pill warn gap" title="no new claims">cordoned: {s.cordon}</span>{/if}</td>
              <td>
                {#if s.current}
                  <a href={'?ticket=' + s.current} onclick={(e) => { e.preventDefault(); e.stopPropagation(); go({ ticket: s.current }); }} class="mono tk">{s.current}</a>
                {/if}
              </td>
              <td class="muted">{s.ticket?.pool ?? ''}</td>
              <td class="muted mono">{s.model ?? ''}{s.effort ? ' · ' + s.effort : ''}{#if s.override}<span class="pill warn gap">override</span>{/if}</td>
              <td class="muted c-time">{ago(s.seq_time)}</td>
              <td class="c-act">
                {#if s.alive}
                  {#if s.cordon}
                    <button onclick={(e) => cordon(e, { id: s.id, uncordon: true })}>Uncordon</button>
                  {:else}
                    <button onclick={(e) => cordon(e, { id: s.id })}>Cordon</button>
                  {/if}
                {/if}
                {#if s.ticket && s.ticket.status !== 'closed'}
                  <button onclick={(e) => close(e, s.id)} title="End the session now and free its slot" data-testid="session-close">Close</button>
                {/if}
              </td>
            </tr>
          {:else}
            <tr><td colspan="7" class="faint">No live sessions.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  tr.sel td { background: var(--accent-soft); }
  .gap { margin-left: 6px; }
  .tk { display: inline-block; padding: 2px 0; }
  .c-act button { height: 24px; }
  .c-act button + button { margin-left: 4px; }
  @media (max-width: 768px) {
    .tk { padding: 8px 0; }
    tr.sel td { background: transparent; }
    tr.sel { background: var(--accent-soft); }
    .c-act button { height: 40px; }
  }
</style>
