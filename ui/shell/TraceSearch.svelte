<script lang="ts">
  import { onDestroy } from 'svelte';
  import type { Ticket } from '../../packages/jobshop-client/rpc.js';
  import { client } from './lib/api.svelte.js';
  import { state as ticketState, statusLetter } from './trace.js';
  import TitleAction from './TitleAction.svelte';

  let { reference, onopen }: { reference: string; onopen: (id: string) => void } = $props();
  let query = $state('');
  let status = $state<'open' | 'in_progress' | 'closed' | ''>('');
  let results: Ticket[] = $state([]);
  let searched = $state<string | null>(null);
  let searchedStatus: typeof status = '';
  let more = $state(false);
  let busy = $state(false);
  let error = $state('');
  let request: AbortController | undefined;
  const pageSize = 50;

  $effect(() => { query = reference; });
  onDestroy(() => request?.abort());

  function clear() {
    request?.abort();
    request = undefined;
    busy = false;
    searched = null;
    results = [];
    error = '';
    more = false;
  }
  async function search(append = false) {
    request?.abort();
    const controller = new AbortController();
    request = controller;
    if (!append) {
      searched = query.trim();
      searchedStatus = status;
      results = [];
      more = false;
    }
    busy = true;
    error = '';
    try {
      const tickets = await client.call('ticket.list', {
        query: searched || '', status: searchedStatus, order: 'newest', limit: pageSize + 1,
        after: append ? results.at(-1)?.id : undefined,
      }, { signal: controller.signal });
      if (request !== controller) return;
      more = tickets.length > pageSize;
      results = append ? [...results, ...tickets.slice(0, pageSize)] : tickets.slice(0, pageSize);
    } catch (e) {
      if (request === controller && !controller.signal.aborted) error = String(e);
    } finally {
      if (request === controller) busy = false;
    }
  }
</script>

<div class="tag trace-search-bar">
  <span class="mono">/trace</span>
  <form role="search" aria-label="Search tickets" onsubmit={(event) => { event.preventDefault(); void search(); }}>
    <input aria-label="Trace reference" bind:value={query} maxlength="512" placeholder="title, ticket, session, repo#PR" />
    <select aria-label="Ticket status" bind:value={status}>
      <option value="">all</option><option value="open">open</option><option value="in_progress">running</option><option value="closed">closed</option>
    </select>
    <button type="submit">search</button>
    <button type="button" disabled={!query.trim()} title="Open an exact ticket, session or external reference" onclick={() => onopen(query.trim())}>open</button>
  </form>
</div>
{#if searched !== null}
  <section class="trace-search-results" aria-label="Ticket search results" aria-busy={busy}>
    <div class="tag">
      <span role="status">{busy ? 'searching…' : `${results.length}${more ? '+' : ''} matches`} · {searched || 'recent tickets'}</span>
      <button class="push-right" onclick={clear}>dismiss</button>
    </div>
    {#if error}<p class="error" role="alert">{error}</p>{:else if !busy && !results.length}<p class="muted">no matching tickets. try a title, ticket id, or related reference.</p>{/if}
    {#if results.length}<div class="table-scroll">
      <table class="mono">
        <thead><tr><th title="Status: O open, R running, C closed">s</th><th>ticket</th><th>kind</th><th class="wide">pool</th></tr></thead>
        <tbody>{#each results as ticket (ticket.id)}<tr>
          <td class="ticket-status" data-status={ticket.status} title={ticketState(ticket)} aria-label={ticketState(ticket)}>{statusLetter(ticket)}</td>
          <td><TitleAction class="ticket-label" title={ticket.title} fallback={ticket.id} tip={ticket.id} href={"?view=trace&ref=" + encodeURIComponent(ticket.id)} /></td>
          <td>{ticket.kind}</td><td class="wide">{ticket.pool || '–'}</td>
        </tr>{/each}</tbody>
      </table>
    </div>{/if}
    {#if more}<button disabled={busy} onclick={() => search(true)}>more</button>{/if}
  </section>
{:else if !reference}
  <p class="muted trace-search-hint">search all tickets, including completed work, or open an exact reference.</p>
{/if}
