<script lang="ts">
  import type { GraphResult, ShowResult } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  import { linkLabel, relationshipText, safeLinks, state } from './trace.js';
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
  <p><b>{state(ticket)}</b> · {ticket.kind}{ticket.pool ? ' · ' + ticket.pool : ''}</p>
  <p class="muted">Created {new Date(ticket.created_at).toLocaleString()} · Updated {new Date(ticket.updated_at).toLocaleString()}</p>
  <div class="ticket-links">
    {#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}
    <button onclick={() => ontrace(ticket.id)}>Open as graph root</button>
  </div>
  {#if ticket.labels?.length}<p class="muted">Labels: {ticket.labels.join(', ')}</p>{/if}
  {#if fields?.wait?.reason}<p><b>Waiting:</b> {fields.wait.reason}</p>{/if}
  {#if fields?.step}<p>Step: {fields.step}{fields.failures ? ' · ' + fields.failures + ' failures' : ''}</p>{/if}
  {#if ticket.parent_id}<p>Parent: <button class="mono" onclick={() => onselect(ticket.parent_id!)}>{ticket.parent_id}</button></p>{/if}
  {#if ticket.holder}<p>Holder: <button class="mono" onclick={() => onselect(ticket.holder!)}>{ticket.holder}</button></p>{/if}
  {#if ticket.kind === 'session'}<button onclick={() => onjournal(ticket.id)}>Load journal</button>{/if}
  <h3>Relationships</h3>
  <div class="table-scroll"><table aria-label="Ticket relationships">
    <thead><tr><th>Relationship</th><th>Ticket</th><th>Status</th></tr></thead>
    <tbody>{#each related as edge}
      {@const id = edge.from === ticket.id ? edge.to : edge.from}
      {@const other = graph.nodes.find(node => node.id === id)!}
      <tr><td>{relationshipText(edge, ticket.id)}</td><td><button class="relation-ticket" onclick={() => onselect(id)}><span class="mono">{id}</span> {other.title}</button></td><td>{state(other)}</td></tr>
    {:else}<tr><td colspan="3" class="muted">No recorded relationships.</td></tr>{/each}</tbody>
  </table></div>
  {#if ticket.body}<h3>Body</h3><div class="markdown">{@html md(ticket.body)}</div>{/if}
  {#if detail.comments?.length}<h3>Comments</h3>
    {#each detail.comments as comment}<article class="ticket-comment"><p class="muted">{comment.author} · {new Date(comment.created_at).toLocaleString()}</p><div class="markdown">{@html md(comment.body)}</div></article>{/each}
  {/if}
  {#if detail.history?.length}<details><summary>History ({detail.history.length})</summary>
    <ol>{#each detail.history as entry}<li>{entry.op} · {entry.actor} · {new Date(entry.at).toLocaleString()}</li>{/each}</ol>
  </details>{/if}
  <details><summary>Raw ticket data</summary><pre>{JSON.stringify({ fields: ticket.fields, meta: ticket.meta }, null, 2)}</pre></details>
</div>
