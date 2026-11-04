<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource, post, live } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';
  import Journal from './Journal.svelte';

  const sessions = resource(() => '/api/sessions');

  let actErr = $state('');

  /** @param {Event} e @param {any} body */
  async function cordon(e, body) {
    e.stopPropagation();
    actErr = '';
    try { await post('/api/session/cordon', body); live.tick++; } catch (err) { actErr = String(/** @type {any} */ (err).message || err); }
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
            <tr class="click" class:sel={url.session === s.id} onclick={() => go({ session: url.session === s.id ? '' : s.id })}>
              <td class="mono c-id">{s.id}{#if s.crashes}<span class="pill bad gap">{s.crashes} crashes</span>{/if}</td>
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
              </td>
            </tr>
            {#if url.session === s.id}
              <tr class="jrow"><td colspan="7" class="journal full"><Journal id={s.id} /></td></tr>
            {/if}
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
  .journal { padding: 0; background: var(--bg); }
  .gap { margin-left: 6px; }
  .tk { display: inline-block; padding: 2px 0; }
  .c-act button { height: 24px; }
  @media (max-width: 720px) {
    .tk { padding: 8px 0; }
    tr.sel td { background: transparent; }
    tr.sel { background: var(--accent-soft); }
    .stack tr.jrow { padding: 0; }
    .journal { flex: 1 0 100%; }
    .c-act button { height: 40px; }
  }
</style>
