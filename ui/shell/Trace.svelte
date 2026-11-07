<script lang="ts">
  import type { GraphResult } from '../../packages/jobshop-client/rpc.js';
  import { graphLayout, linkLabel, relationLabel, safeLinks, state } from './trace.js';
  let { graph, selected, onselect }: {
    graph: GraphResult;
    selected: string;
    onselect: (id: string) => void;
  } = $props();
  const layout = $derived(graphLayout(graph));
</script>

<section aria-label="Ticket dependency graph" class="ticket-graph">
  <div class="graph-heading">
    <h2>Ticket graph</h2>
    <span class="muted">{graph.nodes.length} tickets · {graph.relations.length} relationships</span>
  </div>
  {#if graph.truncated}<p role="status">This graph exceeds the display limit. Showing a bounded portion. Open a related ticket as a new graph to explore further.</p>{/if}
  <p class="muted graph-help">Select a ticket to inspect it. Arrows show parent → child, or dependent → prerequisite. Scroll to explore.</p>
  <div class="graph-scroll" role="region" aria-label="Scrollable ticket graph">
    <div class="graph-canvas" style:width={layout.width + 'px'} style:height={layout.height + 'px'}>
      <svg width={layout.width} height={layout.height} aria-hidden="true">
        <defs><marker id="graph-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" /></marker></defs>
        {#each graph.relations as edge}
          {@const from = layout.positions.get(edge.from)!}
          {@const to = layout.positions.get(edge.to)!}
          {@const right = to.x > from.x}
          {@const x1 = from.x + (right || to.x === from.x ? 248 : 0)}
          {@const x2 = to.x + (right ? 0 : 248)}
          {@const y1 = from.y + 72}
          {@const y2 = to.y + 72}
          {@const bend = to.x === from.x ? x1 + 36 : (x1 + x2) / 2}
          <path class:graph-gate={edge.kind === 'blocks' || edge.kind === 'waits_for'} d={`M ${x1} ${y1} C ${bend} ${y1}, ${bend} ${y2}, ${x2} ${y2}`} marker-end="url(#graph-arrow)"><title>{edge.from} {relationLabel(edge.kind)} {edge.to}</title></path>
        {/each}
      </svg>
      {#each graph.nodes as node (node.id)}
        {@const position = layout.positions.get(node.id)!}
        <article class="graph-card" class:graph-selected={selected === node.id} style:left={position.x + 'px'} style:top={position.y + 'px'} data-ticket={node.id}>
          <button class="graph-select" aria-label={'Inspect ' + node.id} aria-pressed={selected === node.id} onclick={() => onselect(node.id)}>
            <span class="mono">{node.id}</span>
            <span class="graph-state" data-status={node.status}>{state(node)}</span>
            <strong title={node.title}>{node.title || node.kind}</strong>
            <span class="muted">{node.kind}{node.pool ? ' · ' + node.pool : ''}</span>
          </button>
          <div class="graph-links">{#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}</div>
        </article>
      {/each}
    </div>
  </div>
</section>
