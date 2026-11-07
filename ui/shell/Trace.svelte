<script lang="ts">
  import type { GraphResult } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  import { graphLayout, linkLabel, relationLabel, safeLinks, state as ticketState, stepName } from './trace.js';
  let { graph, selected, onselect, onwatch }: {
    graph: GraphResult;
    selected: string;
    onselect: (id: string) => void;
    onwatch: (id: string) => void;
  } = $props();
  const layout = $derived(graphLayout(graph));
  let viewport: HTMLDivElement;
  let width = $state(0);
  let height = $state(0);
  let zoom = $state(1);
  let x = $state(0);
  let y = $state(0);
  let touched = $state(false);
  let dragging = $state(false);
  let root = '';
  const pointers = new Map<number, { x: number; y: number }>();

  $effect(() => {
    if (root !== graph.root) {
      root = graph.root;
      touched = false;
      pointers.clear();
      dragging = false;
    }
  });
  $effect(() => {
    if (!touched && width && height) {
      zoom = Math.min(1, (width - 16) / layout.width, (height - 16) / layout.height);
      x = 8;
      y = 8;
    }
  });

  function zoomAt(next: number, clientX?: number, clientY?: number) {
    const box = viewport.getBoundingClientRect();
    const px = clientX == null ? width / 2 : clientX - box.left;
    const py = clientY == null ? height / 2 : clientY - box.top;
    next = Math.max(0.05, Math.min(3, next));
    const ratio = next / zoom;
    x = px - (px - x) * ratio;
    y = py - (py - y) * ratio;
    zoom = next;
    touched = true;
  }

  function wheel(element: HTMLDivElement) {
    const handler = (event: WheelEvent) => {
      event.preventDefault();
      const delta = event.deltaY * (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? height : 1);
      zoomAt(zoom * Math.exp(-Math.max(-400, Math.min(400, delta)) * 0.002), event.clientX, event.clientY);
    };
    element.addEventListener('wheel', handler, { passive: false });
    return { destroy: () => element.removeEventListener('wheel', handler) };
  }

  function pointerDown(event: PointerEvent) {
    if (event.button !== 0 || (event.target as Element).closest('button, a')) return;
    viewport.setPointerCapture(event.pointerId);
    pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
    dragging = true;
    touched = true;
    viewport.focus();
  }
  function pointerMove(event: PointerEvent) {
    const previous = pointers.get(event.pointerId);
    if (!previous) return;
    const before = [...pointers.values()];
    pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (pointers.size === 2) {
      const after = [...pointers.values()];
      const distance = (points: typeof before) => Math.hypot(points[1].x - points[0].x, points[1].y - points[0].y);
      if (distance(before) > 0) zoomAt(zoom * distance(after) / distance(before), (before[0].x + before[1].x) / 2, (before[0].y + before[1].y) / 2);
      x += (after[0].x + after[1].x - before[0].x - before[1].x) / 2;
      y += (after[0].y + after[1].y - before[0].y - before[1].y) / 2;
    } else if (pointers.size === 1) {
      x += event.clientX - previous.x;
      y += event.clientY - previous.y;
    }
  }
  function pointerEnd(event: PointerEvent) {
    pointers.delete(event.pointerId);
    dragging = pointers.size > 0;
  }
  function keydown(event: KeyboardEvent) {
    if (event.target !== viewport) return;
    const delta: Record<string, [number, number]> = { ArrowLeft: [40, 0], ArrowRight: [-40, 0], ArrowUp: [0, 40], ArrowDown: [0, -40] };
    if (event.key in delta) {
      event.preventDefault(); touched = true;
      x += delta[event.key][0]; y += delta[event.key][1];
    } else if (event.key === '+' || event.key === '=') { event.preventDefault(); zoomAt(zoom * 1.25); }
    else if (event.key === '-') { event.preventDefault(); zoomAt(zoom / 1.25); }
    else if (event.key === '0') { event.preventDefault(); touched = false; }
  }
</script>

<section aria-label="Ticket dependency graph" class="ticket-graph">
  <div class="graph-heading">
    <h2>ticket graph</h2>
    <span class="muted">{graph.nodes.length} tickets · {graph.relations.length} relationships</span>
    <div class="graph-controls" aria-label="Graph zoom">
      <button aria-label="Zoom out" disabled={zoom <= 0.05} onclick={() => zoomAt(zoom / 1.25)}>−</button>
      <output aria-label="Graph zoom level">{Math.round(zoom * 100)}%</output>
      <button aria-label="Zoom in" disabled={zoom >= 3} onclick={() => zoomAt(zoom * 1.25)}>+</button>
      <button onclick={() => touched = false}>reset</button>
    </div>
  </div>
  {#if graph.truncated}<p role="status">This graph exceeds the display limit. Showing a bounded portion. Open a related ticket as a new graph to explore further.</p>{/if}
  <p class="muted graph-help">Creation time → left to right, spaced by elapsed time. Arrows show relationships, not execution order. Drag to pan; wheel or pinch to zoom. Keyboard: arrows, +/−, 0 to reset.</p>
  <!-- Graph camera has application keyboard controls; ticket buttons remain native. -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
  <div class="graph-scroll" class:dragging bind:this={viewport} bind:clientWidth={width} bind:clientHeight={height} use:wheel role="application" aria-label="Ticket graph navigation" tabindex="0" onpointerdown={pointerDown} onpointermove={pointerMove} onpointerup={pointerEnd} onpointercancel={pointerEnd} onlostpointercapture={pointerEnd} onkeydown={keydown}>
    <div class="graph-canvas" style:width={layout.width + 'px'} style:height={layout.height + 'px'} style:transform={`translate(${x}px, ${y}px) scale(${zoom})`}>
      <div class="graph-time-axis" aria-label="Creation time axis">{#each layout.ticks as tick}<time style:left={tick.x + 'px'} datetime={tick.at}>{new Date(tick.at).toLocaleString()}</time>{/each}</div>
      <svg width={layout.width} height={layout.height} aria-hidden="true">
        <defs><marker id="graph-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" /></marker></defs>
        {#each graph.relations as edge}
          {@const from = layout.positions.get(edge.from)!}
          {@const to = layout.positions.get(edge.to)!}
          {@const right = to.x > from.x}
          {@const x1 = from.x + (right || to.x === from.x ? 248 : 0)}
          {@const x2 = to.x + (right ? 0 : 248)}
          {@const y1 = from.y + 88}
          {@const y2 = to.y + 88}
          {@const bend = to.x === from.x ? x1 + 36 : (x1 + x2) / 2}
          <path class:graph-gate={edge.kind === 'blocks' || edge.kind === 'waits_for'} d={`M ${x1} ${y1} C ${bend} ${y1}, ${bend} ${y2}, ${x2} ${y2}`} marker-end="url(#graph-arrow)"><title>{edge.from} {relationLabel(edge.kind)} {edge.to}</title></path>
        {/each}
      </svg>
      {#each graph.nodes as node (node.id)}
        {@const position = layout.positions.get(node.id)!}
        {@const wait = (node.fields as { wait?: { reason?: string } } | undefined)?.wait}
        <article class="graph-card" class:graph-selected={selected === node.id} style:left={position.x + 'px'} style:top={position.y + 'px'} data-ticket={node.id}>
          <div class="graph-state" data-status={node.status}>{@html md(ticketState(node), true)}{#if wait?.reason}<div>{@html md(wait.reason, true)}</div>{/if}</div>
          <button class="graph-select" aria-label={'Inspect ' + node.id} aria-pressed={selected === node.id} onclick={() => onselect(node.id)}>
            <span class="mono">{node.id}</span>
            <strong title={node.title}>{node.title || node.kind}</strong>
            <span class="muted">{node.kind}{node.pool ? ' · ' + node.pool : ''}</span>
            {#if stepName(node.fields)}<span>step: {stepName(node.fields)}</span>{/if}
            <time class="muted" datetime={node.created_at}>created {new Date(node.created_at).toLocaleString()}</time>
            {#if node.closed_at}<time class="muted" datetime={node.closed_at}>closed {new Date(node.closed_at).toLocaleString()}</time>{/if}
          </button>
          <div class="graph-links">{#if node.holder || node.kind === 'session'}<button title={node.holder || node.id} onclick={() => onwatch(node.holder || node.id)}>view session ↗</button>{/if}{#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}</div>
        </article>
      {/each}
    </div>
  </div>
</section>
