<script lang="ts">
  import type { GraphResult, HistoryEntry, Ticket } from '../../packages/jobshop-client/rpc.js';
  import { stepName } from './trace.js';
  import { md } from '../kit/markdown.js';
  let { entries, ticket, graph, onwatch }: { entries: HistoryEntry[]; ticket: Ticket; graph: GraphResult; onwatch: (id: string) => void } = $props();
  function label(id: string) {
    const node = graph.nodes.find(node => node.id === id);
    return node?.title || id;
  }
  function isSession(id: string) {
    return graph.nodes.some(node => node.id === id && node.kind === 'session') || /^[a-zA-Z0-9][a-zA-Z0-9._-]*-s-[a-zA-Z0-9]+$/.test(id);
  }
  function describe(entry: HistoryEntry) {
    const diff = (entry.diff || {}) as Record<string, any>;
    const fields = diff.fields || {};
    const title = ({ create: 'created', claim: diff.resumed ? 'resumed work' : 'claimed', release: 'released', close: 'closed', reopen: 'reopened', update: 'updated', edge_add: 'relationship added', edge_remove: 'relationship removed', label_add: 'label added', label_remove: 'label removed', comment: 'commented', expire: 'claim expired' } as Record<string, string>)[entry.op] || entry.op.replaceAll('_', ' ');
    const step = stepName(fields) || stepName(ticket.fields);
    const summary = [diff.reason, diff.note, diff.resolution, diff.title, fields.wait?.reason, fields.state, diff.kind, diff.label].filter(value => typeof value === 'string' && value).join(' · ');
    const refs = [...new Set([diff.holder, diff.parent_id, diff.to, diff.from, diff.session, diff.session_id, diff.ticket_id, diff.request_changes].filter(value => typeof value === 'string' && value !== ticket.id))] as string[];
    return { title, step, summary, refs };
  }
</script>

<details class="ticket-history" open>
  <summary>history ({entries.length})</summary>
  <ol>
    {#each [...entries].reverse() as entry (entry.seq)}
      {@const event = describe(entry)}
      <li>
        <div class="history-heading"><b>{event.title}</b>{#if event.step}<span class="history-step">step: {event.step}</span>{/if}<time class="muted" datetime={entry.at}>{new Date(entry.at).toLocaleString()}</time></div>
        {#if event.summary}<div class="history-summary markdown">{@html md(event.summary, true)}</div>{/if}
        {#if event.refs.length}<div class="history-references">{#each event.refs as id}{#if isSession(id)}<button title={id} onclick={() => onwatch(id)}>session: {label(id)} ↗</button>{:else}<a title={id} href={'?view=trace&ref=' + encodeURIComponent(id)}>{label(id)} <small class="mono">{id}</small></a>{/if}{/each}</div>{/if}
        <div class="history-footer"><span class="muted">by {entry.actor}</span>{#if isSession(entry.actor)}<button title={entry.actor} onclick={() => onwatch(entry.actor)}>view session</button>{/if}
          {#if entry.diff && Object.keys(entry.diff).length}<details class="history-raw"><summary>details</summary><pre>{JSON.stringify(entry.diff, null, 2)}</pre></details>{/if}
        </div>
      </li>
    {/each}
  </ol>
</details>
