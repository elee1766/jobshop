<script lang="ts">
  import type { GraphResult, ShowResult, Ticket } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  import { linkLabel, relationshipText, safeLinks, state, stateDetail, statusLetter } from './trace.js';
  import TicketHistory from './TicketHistory.svelte';
  let { detail, graph, onselect, ontrace, onwatch }: {
    detail: ShowResult;
    graph: GraphResult;
    onselect: (id: string) => void;
    ontrace: (id: string) => void;
    onwatch: (id: string) => void;
  } = $props();
  const ticket = $derived(detail.ticket);
  const node = $derived(graph.nodes.find(n => n.id === ticket.id));
  const parent = $derived(graph.nodes.find(n => n.id === ticket.parent_id));
  const groups = $derived.by(() => {
    const groups = new Map<string, Ticket[]>();
    for (const edge of graph.relations) {
      if (edge.from !== ticket.id && edge.to !== ticket.id) continue;
      const id = edge.from === ticket.id ? edge.to : edge.from;
      if (edge.kind === 'parent' && id === ticket.parent_id) continue;
      const other = graph.nodes.find(node => node.id === id);
      if (!other) continue;
      const kind = edge.kind === 'parent' && edge.from === ticket.id ? 'children' : relationshipText(edge, ticket.id);
      const list = groups.get(kind) || [];
      if (!list.some(node => node.id === id)) list.push(other);
      groups.set(kind, list);
    }
    const priority = ['blocked by', 'blocks', 'waits for', 'waited on by', 'children'];
    return [...groups].sort(([a], [b]) => (priority.indexOf(a) < 0 ? 99 : priority.indexOf(a)) - (priority.indexOf(b) < 0 ? 99 : priority.indexOf(b)));
  });
  const fields = $derived(ticket.fields as { wait?: { reason?: string }; step?: string; state?: string; failures?: number } | undefined);
</script>

{#snippet ticketLink(other: Ticket)}
  <span class="ticket-status mono" data-status={other.status} role="img" title={state(other)} aria-label={state(other)}>{statusLetter(other)}</span>
  <button class="relation-ticket ticket-label" title={other.id} onclick={() => onselect(other.id)}><span>{other.title || other.id}</span>{#if other.title}<small class="mono muted">{other.id}</small>{/if}</button>
{/snippet}

<div class="inspector" aria-label="Ticket details">
  <header class="ticket-overview">
    {#if ticket.parent_id}<div class="ticket-parent"><span class="muted">child of</span>{#if parent}{@render ticketLink(parent)}{:else}<button class="mono" onclick={() => onselect(ticket.parent_id!)}>{ticket.parent_id}</button>{/if}</div>{/if}
    <h2>{ticket.title || ticket.id}{#if ticket.title}<small class="mono muted">{ticket.id}</small>{/if}</h2>
    <p><span class="ticket-status mono" data-status={ticket.status} role="img" title={state(ticket)} aria-label={state(ticket)}>{statusLetter(ticket)}</span> {#if stateDetail(ticket)}<b>{@html md(stateDetail(ticket), true)}</b> · {/if}{ticket.kind}{ticket.pool ? ' · ' + ticket.pool : ''}</p>
    <p class="muted">created {new Date(ticket.created_at).toLocaleString()} · updated {new Date(ticket.updated_at).toLocaleString()}</p>
    <div class="ticket-links">
      {#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}
      <button onclick={() => ontrace(ticket.id)}>open as graph root</button>
      {#if ticket.holder}<button class="watch-session" title={ticket.holder} onclick={() => onwatch(ticket.holder!)}>view running session ↗</button>{/if}
      {#if ticket.kind === 'session'}<button class="watch-session" onclick={() => onwatch(ticket.id)}>view session ↗</button>{/if}
    </div>
    {#if ticket.labels?.length}<p class="muted">labels: {ticket.labels.join(', ')}</p>{/if}
    {#if fields?.wait?.reason}<p><b>waiting:</b> {@html md(fields.wait.reason, true)}</p>{/if}
    {#if fields?.step}<p>step: {fields.step}{fields.failures ? ' · ' + fields.failures + ' failures' : ''}</p>{/if}
  </header>
  {#if groups.length}<div class="relationship-groups" aria-label="Ticket relationships">
    {#each groups as [kind, others]}<section class="relationship-group" aria-label={kind}>
      <h3>{kind} <span class="muted">{others.length}</span></h3>
      <ul>{#each others as other (other.id)}<li>{@render ticketLink(other)}</li>{/each}</ul>
    </section>{/each}
  </div>{/if}
  {#if detail.history?.length}<TicketHistory entries={detail.history} {ticket} {graph} {onwatch} />{/if}
  {#if ticket.body}<details class="ticket-body" open><summary>body</summary><div class="markdown">{@html md(ticket.body)}</div></details>{/if}
  {#if detail.comments?.length}<h3>comments</h3>
    {#each detail.comments as comment}<article class="ticket-comment"><p class="muted">{comment.author} · {new Date(comment.created_at).toLocaleString()}</p><div class="markdown">{@html md(comment.body)}</div></article>{/each}
  {/if}
  <details><summary>raw ticket data</summary><pre>{JSON.stringify({ fields: ticket.fields, meta: ticket.meta }, null, 2)}</pre></details>
</div>
