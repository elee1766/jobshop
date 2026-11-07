<script lang="ts">
  import type { GraphResult, HistoryEntry, Ticket } from '../../packages/jobshop-client/rpc.js';
  import { stepName } from './trace.js';
  import { md } from '../kit/markdown.js';
  let { entries, ticket, graph }: { entries: HistoryEntry[]; ticket: Ticket; graph: GraphResult } = $props();

  function changes(value: unknown, path = ''): { path: string; value: string; target: string }[] {
    if (value && typeof value === 'object' && Object.keys(value).length) {
      return Object.entries(value).flatMap(([key, child]) => changes(child, path ? path + '.' + key : key));
    }
    const key = path.split('.').filter(part => !/^\d+$/.test(part)).at(-1);
    const reference = ['parent_id', 'holder', 'run', 'ticket', 'ticket_id', 'session', 'session_id', 'to', 'needs'].includes(key || '');
    return [{ path, value: value == null ? 'cleared' : typeof value === 'object' ? 'empty' : String(value), target: reference && typeof value === 'string' && /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(value) ? value : '' }];
  }

  function label(id: string) {
    const node = graph.nodes.find(node => node.id === id);
    const step = stepName(node?.fields);
    return [step, node?.title, id].filter(Boolean).join(' · ');
  }

  function actorIsTicket(actor: string) {
    return graph.nodes.some(node => node.id === actor) || /^[a-zA-Z0-9][a-zA-Z0-9._-]*-s-[a-zA-Z0-9]+$/.test(actor);
  }
</script>

<details class="ticket-history">
  <summary>history ({entries.length})</summary>
  <ol>
    {#each entries as entry (entry.seq)}
      {@const diff = entry.diff as { fields?: unknown; step?: unknown } | undefined}
      {@const step = stepName(diff?.fields) || (typeof diff?.step === 'string' ? diff.step : '') || stepName(ticket.fields)}
      <li>
        <div class="history-heading">
          <b>{entry.op.replaceAll('_', ' ')}</b>
          {#if step}<span>step: {step}</span>{/if}
          {#if actorIsTicket(entry.actor)}<a href={'?view=trace&ref=' + encodeURIComponent(entry.actor)}>{label(entry.actor)}</a>{:else}<span>{entry.actor}</span>{/if}
          <time class="muted" datetime={entry.at}>{new Date(entry.at).toLocaleString()}</time>
        </div>
        {#if entry.diff && Object.keys(entry.diff).length}
          <dl class="history-changes">
            {#each changes(entry.diff) as change}
              <dt>{change.path.replace(/^fields\./, '').replaceAll('_', ' ')}</dt>
              <dd>{#if change.target}<a href={'?view=trace&ref=' + encodeURIComponent(change.target)}>{label(change.target)}</a>{:else}{@html md(change.value, true)}{/if}</dd>
            {/each}
          </dl>
        {/if}
      </li>
    {/each}
  </ol>
</details>
