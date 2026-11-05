<script>
  import { go } from '../lib/url.svelte.js';
  import { ago } from '../lib/format.js';
  import { resource } from '../lib/api.svelte.js';
  import Progress from './Progress.svelte';

  /** @type {{ items: { data: any, error: string, loading: boolean } }} */
  let { items } = $props();

  // An escalation ticket with a chat opens that chat; any other ticket
  // opens its drawer.
  const escalations = resource(() => 'api/escalations');

  /** @param {any} it */
  function open(it) {
    const chat = it.kind === 'ticket' && escalations.data?.find((/** @type {any} */ e) => e.ticket === it.id);
    if (chat) go({ session: chat.session });
    else if (it.kind === 'run' || it.kind === 'ticket') go({ ticket: it.id });
    else if (it.kind === 'pool') go({ view: 'pools', ticket: '' });
    else if (it.kind === 'deploy') go({ view: 'deploys', ticket: '' });
  }
</script>

<header class="vh">
  <h1>Attention</h1>
  <span class="sub">What waits on a human, oldest first.</span>
</header>
<div class="pad">
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
      <table class="stack" data-testid="attention-list">
        <thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead>
        <tbody>
          {#each items.data as it (it.kind + it.id)}
            <tr class="click" onclick={() => open(it)}>
              <td class="c-id"><span class="pill warn">{it.kind}</span></td>
              <td class="c-title">{it.title || it.id}<div class="faint mono">{it.id}</div></td>
              <td class="muted full">{it.why}</td>
              <td class="muted c-time">{ago(it.since)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <h2 class="st">In progress</h2>
  <Progress />
</div>
