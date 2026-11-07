<script lang="ts">
  import type { GraphResult, ShowResult } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  import { linkLabel, relationshipText, safeLinks, state, statusLetter } from './trace.js';
  import TicketHistory from './TicketHistory.svelte';
  let { detail, graph, onselect, ontrace, onjournal }: {
    detail: ShowResult;
    graph: GraphResult;
    onselect: (id: string) => void;
    ontrace: (id: string) => void;
    onjournal: (id: string) => void;
  } = $props();
  const ticket = $derived(detail.ticket);
  const node = $derived(graph.nodes.find(n => n.id === ticket.id));
  const related = $derived(graph.relations.filter(edge => edge.from === ticket.id || edge.to === ticket.id));
  const fields = $derived(ticket.fields as { wait?: { reason?: string }; step?: string; state?: string; failures?: number } | undefined);
</script>

<div class="inspector" aria-label="Ticket details">
  <h2 class="mono">{ticket.id} {ticket.title}</h2>
  <p><b>{@html md(state(ticket), true)}</b> · {ticket.kind}{ticket.pool ? ' · ' + ticket.pool : ''}</p>
  <p class="muted">created {new Date(ticket.created_at).toLocaleString()} · updated {new Date(ticket.updated_at).toLocaleString()}</p>
  <div class="ticket-links">
    {#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}
    <button onclick={() => ontrace(ticket.id)}>open as graph root</button>
  </div>
  {#if ticket.labels?.length}<p class="muted">labels: {ticket.labels.join(', ')}</p>{/if}
  {#if fields?.wait?.reason}<p><b>waiting:</b> {@html md(fields.wait.reason, true)}</p>{/if}
  {#if fields?.step}<p>step: {fields.step}{fields.failures ? ' · ' + fields.failures + ' failures' : ''}</p>{/if}
  {#if ticket.parent_id}<p>parent: <button class="mono" onclick={() => onselect(ticket.parent_id!)}>{ticket.parent_id}</button></p>{/if}
  {#if ticket.holder}<p>holder: <button class="mono" onclick={() => onselect(ticket.holder!)}>{ticket.holder}</button></p>{/if}
  {#if ticket.kind === 'session'}<button onclick={() => onjournal(ticket.id)}>load journal</button>{/if}
  <h3>relationships</h3>
  <div class="table-scroll"><table aria-label="Ticket relationships">
    <thead><tr><th title="Status: O open, R running, C closed">s</th><th>ticket</th><th>relationship</th></tr></thead>
    <tbody>{#each related as edge}
      {@const id = edge.from === ticket.id ? edge.to : edge.from}
      {@const other = graph.nodes.find(node => node.id === id)!}
      <tr><td class="ticket-status mono" title={state(other)} aria-label={state(other)}>{statusLetter(other)}</td><td><button class="relation-ticket ticket-label" onclick={() => onselect(id)}><span>{other.title || id}</span>{#if other.title}<small class="mono muted">{id}</small>{/if}</button></td><td>{relationshipText(edge, ticket.id)}</td></tr>
    {:else}<tr><td colspan="3" class="muted">No recorded relationships.</td></tr>{/each}</tbody>
  </table></div>
  {#if ticket.body}<h3>body</h3><div class="markdown">{@html md(ticket.body)}</div>{/if}
  {#if detail.comments?.length}<h3>comments</h3>
    {#each detail.comments as comment}<article class="ticket-comment"><p class="muted">{comment.author} · {new Date(comment.created_at).toLocaleString()}</p><div class="markdown">{@html md(comment.body)}</div></article>{/each}
  {/if}
  {#if detail.history?.length}<TicketHistory entries={detail.history} {ticket} {graph} />{/if}
  <details><summary>raw ticket data</summary><pre>{JSON.stringify({ fields: ticket.fields, meta: ticket.meta }, null, 2)}</pre></details>
</div>
