<script>
  import { url, go } from '../lib/url.svelte.js';
  import { resource } from '../lib/api.svelte.js';
  import { ago } from '../lib/format.js';
  import Journal from './Journal.svelte';

  const sessions = resource(() => '/api/sessions');

  /** @param {any} s */
  function state(s) {
    if (!s.alive) return ['dead', 'bad'];
    if (s.backoff_until) return ['backoff', 'warn'];
    if (s.drained) return ['drained', 'warn'];
    if (s.current) return ['working', 'ok'];
    return ['idle', ''];
  }
</script>

<h1>Sessions</h1>
{#if sessions.error}
  <div class="card err">{sessions.error}</div>
{:else if sessions.data}
  <div class="card scroll">
    <table data-testid="sessions">
      <thead><tr><th>Session</th><th>State</th><th>Ticket</th><th class="hide-sm">Pool</th><th class="hide-sm">Last event</th></tr></thead>
      <tbody>
        {#each sessions.data as s (s.id)}
          {@const st = state(s)}
          <tr class="click" class:sel={url.session === s.id} onclick={() => go({ session: url.session === s.id ? '' : s.id })}>
            <td class="mono">{s.id}{#if s.crashes}<span class="pill bad">{s.crashes} crashes</span>{/if}</td>
            <td><span class="pill {st[1]}">{st[0]}</span></td>
            <td>
              {#if s.current}
                <a href={'?ticket=' + s.current} onclick={(e) => { e.preventDefault(); e.stopPropagation(); go({ ticket: s.current }); }} class="mono">{s.current}</a>
              {/if}
            </td>
            <td class="hide-sm muted">{s.ticket?.pool ?? ''}</td>
            <td class="hide-sm muted">{ago(s.seq_time)}</td>
          </tr>
          {#if url.session === s.id}
            <tr><td colspan="5" class="journal"><Journal id={s.id} /></td></tr>
          {/if}
        {:else}
          <tr><td colspan="5" class="faint">No live sessions.</td></tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0 0 12px; }
  .err { padding: 12px; color: var(--bad); }
  tr.sel td { background: var(--accent-soft); }
  .journal { padding: 0; background: var(--bg); }
  .pill.bad { margin-left: 6px; }
</style>
