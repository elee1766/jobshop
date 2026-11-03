<script>
  import { go } from '../lib/url.svelte.js';
  import { ago } from '../lib/format.js';

  /** @type {{ items: { data: any, error: string, loading: boolean } }} */
  let { items } = $props();

  /** @param {any} it */
  function open(it) {
    if (it.kind === 'run' || it.kind === 'ticket') go({ ticket: it.id });
    else if (it.kind === 'pool') go({ view: 'pools', ticket: '' });
    else if (it.kind === 'deploy') go({ view: 'deploys', ticket: '' });
  }
</script>

<h1>Attention</h1>
<p class="muted sub">What waits on a human, oldest first.</p>

{#if items.error}
  <div class="card err">{items.error}</div>
{:else if items.loading && !items.data}
  <p class="faint">Loading…</p>
{:else if items.data.length === 0}
  <div class="card empty" data-testid="attention-empty">
    <div class="big">Nothing needs you.</div>
    <div class="muted">No escalations, needs-human runs, quarantined pools or rolled-back deploys.</div>
  </div>
{:else}
  <div class="card scroll">
    <table data-testid="attention-list">
      <thead><tr><th>Kind</th><th>Item</th><th class="hide-sm">Why</th><th>Since</th></tr></thead>
      <tbody>
        {#each items.data as it (it.kind + it.id)}
          <tr class="click" onclick={() => open(it)}>
            <td><span class="pill warn">{it.kind}</span></td>
            <td>{it.title || it.id}<div class="faint mono">{it.id}</div></td>
            <td class="hide-sm muted">{it.why}</td>
            <td class="muted">{ago(it.since)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  h1 { font-size: 18px; font-weight: 600; margin: 0; }
  .sub { margin: 2px 0 16px; }
  .empty { padding: 40px 20px; text-align: center; }
  .big { font-size: 15px; font-weight: 500; margin-bottom: 4px; color: var(--ok); }
  .err { padding: 12px; color: var(--bad); }
</style>
