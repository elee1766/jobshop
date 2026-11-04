<script>
  import { url, go, closeSession } from '../lib/url.svelte.js';
  import { ago } from '../lib/format.js';

  /** @type {{ items: { data: any, error: string, loading: boolean } }} */
  let { items } = $props();

  /** Cause label for the pill. @type {Record<string, string>} */
  const causes = { run: 'run failed', pool: 'pool quarantined', wait: 'wait timed out', guard: 'needs approval', label: 'needs human' };

  const open = $derived((items.data ?? []).filter((/** @type {any} */ e) => e.open));
  const closed = $derived((items.data ?? []).filter((/** @type {any} */ e) => !e.open));

  /** @param {any} e */
  function pick(e) {
    if (url.session === e.session) closeSession();
    else go({ session: e.session });
  }
</script>

{#snippet row(/** @type {any} */ e)}
  <tr class="click" class:sel={url.session === e.session} onclick={() => pick(e)} data-testid="escalation-row">
    <td class="c-title">
      {e.title}
      <div class="faint mono">{e.ticket}{#if e.held && e.held !== e.ticket} · {e.held}{/if}</div>
    </td>
    <td class="c-st"><span class="pill {e.open ? 'warn' : ''}">{causes[e.cause] ?? e.cause ?? 'escalation'}</span></td>
    <td class="muted full">{e.why}</td>
    <td class="c-st"><span class="pill {e.open ? 'accent' : 'ok'}">{e.open ? 'open' : 'closed' + (e.resolution ? ' · ' + e.resolution : '')}</span></td>
    <td class="muted c-time">{ago(e.open ? e.since : e.closed)}</td>
  </tr>
{/snippet}

<h2 class="st">Escalations</h2>
{#if items.error}
  <div class="card err">{items.error}</div>
{:else if items.data}
  {#if open.length === 0 && closed.length === 0}
    <div class="card empty" data-testid="escalations-empty"><div class="muted">No escalation sessions.</div></div>
  {:else}
    <div class="card scroll">
      <table class="stack" data-testid="escalations">
        <thead><tr><th>Ticket</th><th>Cause</th><th>Why</th><th>State</th><th>Age</th></tr></thead>
        <tbody>
          {#each open as e (e.ticket)}{@render row(e)}{/each}
          {#each closed as e (e.ticket)}{@render row(e)}{/each}
        </tbody>
      </table>
    </div>
  {/if}
{/if}
