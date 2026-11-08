<script lang="ts">
  import { onDestroy } from 'svelte';
  import type { GraphResult } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  import { parseTitle } from '../kit/titles.js';
  import TitleText from './TitleText.svelte';
  import { axisTicks, card, exactLabel, fullLabel, graphLayout, linkLabel, relationLabel, relativeLabel, safeLinks, shortLabel, state as ticketState, stateDetail, statusLetter, stepName, traceEnd } from './trace.js';
  let { graph, selected, onselect, onwatch }: {
    graph: GraphResult;
    selected: string;
    onselect: (id: string) => void;
    onwatch: (id: string) => void;
  } = $props();
  // Height of the time axis pinned to the top of the viewport.
  const axis = 48;
  const layout = $derived(graphLayout(graph));
  let viewport: HTMLDivElement;
  let width = $state(0);
  let height = $state(0);
  let zoom = $state(1);
  let x = $state(0);
  let y = $state(0);
  let touched = $state(false);
  let dragging = $state(false);
  let now = $state(Date.now());
  let root = '';
  const pointers = new Map<number, { x: number; y: number }>();
  // An active trace ends at "now", which moves.
  const clock = setInterval(() => { now = Date.now(); }, 30_000);
  onDestroy(() => clearInterval(clock));

  // Canvas x of a time, then its viewport x through the camera: the axis and
  // the cards share this one transform, so ticks stay on their cards while
  // panning and zooming.
  const screenX = (t: number) => x + (card.left + (t - layout.start) * layout.scale) * zoom;
  const timeAt = (px: number) => layout.start + ((px - x) / zoom - card.left) / layout.scale;
  const end = $derived(traceEnd(graph, now));
  const startX = $derived(screenX(layout.start));
  const endX = $derived(screenX(end.at));
  // Endpoint labels stay in view (clamped to the edges) while their marks
  // sit at the exact position. Relative tick labels take the second row;
  // when the endpoint labels would collide the end label moves there too.
  let startW = $state(0);
  let endW = $state(0);
  let tipW = $state(0);
  const pad = 4;
  const startLeft = $derived(Math.max(pad, Math.min(width - pad - startW, startX)));
  const endLeft = $derived(Math.max(pad, Math.min(width - pad - endW, endX - endW)));
  const stacked = $derived(endLeft < startLeft + startW + 8);
  // Relative ticks about 88px apart on screen, labels kept inside the view.
  const ticks = $derived(axisTicks(layout.start, end.at, layout.scale * zoom, timeAt(0), timeAt(width)).filter(t => {
    const px = screenX(t);
    return px > 32 && px < width - 32 && px - startX > 40 && Math.abs(endX - px) > 40 && (!stacked || px + 32 < endLeft - 4 || px - 32 > endLeft + endW + 4);
  }));
  let tip = $state<{ text: string; left: number } | null>(null);
  let pinned = $state(false);

  $effect(() => {
    if (root !== graph.root) {
      root = graph.root;
      touched = false;
      pointers.clear();
      dragging = false;
      tip = null;
    }
  });
  $effect(() => {
    if (!touched && width && height) {
      zoom = Math.min(1, (width - 16) / layout.width, (height - axis - 16) / layout.height);
      x = 8;
      y = axis + 8;
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
    tip = null;
    pinned = false;
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
    if (event.key === 'Escape' && tip) { tip = null; pinned = false; return; }
    if (event.target !== viewport) return;
    const delta: Record<string, [number, number]> = { ArrowLeft: [40, 0], ArrowRight: [-40, 0], ArrowUp: [0, 40], ArrowDown: [0, -40] };
    if (event.key in delta) {
      event.preventDefault(); touched = true;
      x += delta[event.key][0]; y += delta[event.key][1];
    } else if (event.key === '+' || event.key === '=') { event.preventDefault(); zoomAt(zoom * 1.25); }
    else if (event.key === '-') { event.preventDefault(); zoomAt(zoom / 1.25); }
    else if (event.key === '0') { event.preventDefault(); touched = false; }
  }

  // Exact time on hover, focus or tap (a tap pins it until the next tap).
  function show(t: number, left: number) {
    tip = { text: exactLabel(t), left };
  }
  function hide() {
    if (!pinned) tip = null;
  }
  function toggle(t: number, left: number) {
    pinned = !(pinned && tip?.text === exactLabel(t));
    if (pinned) show(t, left); else tip = null;
  }
</script>

{#snippet stamp(t: number, label: string, kind: string, left: number, center: number)}
  <button class={'graph-stamp graph-stamp-' + kind} style:left={left + 'px'} aria-label={label + ', ' + exactLabel(t)}
    onpointerenter={() => show(t, center)} onpointerleave={hide} onfocus={() => show(t, center)} onblur={() => { pinned = false; tip = null; }}
    onclick={() => toggle(t, center)}><time datetime={new Date(t).toISOString()}>{label}</time></button>
{/snippet}

<section aria-label="Ticket dependency graph" class="ticket-graph">
  <div class="tag graph-heading">
    <h2>graph</h2>
    <span class="muted">{graph.nodes.length} tickets · {graph.relations.length} relationships</span>
    <div class="graph-controls push-right" aria-label="Graph zoom">
      <button aria-label="Zoom out" disabled={zoom <= 0.05} onclick={() => zoomAt(zoom / 1.25)}>−</button>
      <output aria-label="Graph zoom level">{Math.round(zoom * 100)}%</output>
      <button aria-label="Zoom in" disabled={zoom >= 3} onclick={() => zoomAt(zoom * 1.25)}>+</button>
      <button onclick={() => touched = false}>reset</button>
    </div>
  </div>
  {#if graph.truncated}<p role="status">This graph exceeds the display limit. Showing a bounded portion. Open a related ticket as a new graph to explore further.</p>{/if}
  <p class="muted graph-help">Creation time → left to right: a card's left edge is when it was created. Arrows show relationships, not execution order. Drag to pan; wheel or pinch to zoom. Keyboard: arrows, +/−, 0 to reset. Focus or tap a time for the exact date.</p>
  <!-- Graph camera has application keyboard controls; ticket buttons remain native. -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
  <div class="graph-scroll" class:dragging bind:this={viewport} bind:clientWidth={width} bind:clientHeight={height} style:--graph-axis={axis + 'px'} use:wheel role="application" aria-label="Ticket graph navigation" tabindex="0" onpointerdown={pointerDown} onpointermove={pointerMove} onpointerup={pointerEnd} onpointercancel={pointerEnd} onlostpointercapture={pointerEnd} onkeydown={keydown}>
    <div class="graph-grid" aria-hidden="true">
      {#each ticks as t (t)}<span style:left={screenX(t) + 'px'}></span>{/each}
      <span class="graph-grid-start" style:left={startX + 'px'}></span>
      <span class="graph-grid-end" class:graph-now={end.active} style:left={endX + 'px'}></span>
    </div>
    <div class="graph-canvas" style:width={layout.width + 'px'} style:height={layout.height + 'px'} style:transform={`translate(${x}px, ${y}px) scale(${zoom})`}>
      <svg width={layout.width} height={layout.height} aria-hidden="true">
        <defs><marker id="graph-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" /></marker></defs>
        {#each graph.relations as edge}
          {@const from = layout.positions.get(edge.from)!}
          {@const to = layout.positions.get(edge.to)!}
          {@const right = to.x > from.x}
          {@const x1 = from.x + (right || to.x === from.x ? card.width : 0)}
          {@const x2 = to.x + (right ? 0 : card.width)}
          {@const y1 = from.y + card.height / 2}
          {@const y2 = to.y + card.height / 2}
          {@const bend = to.x === from.x ? x1 + 36 : (x1 + x2) / 2}
          <path class:graph-gate={edge.kind === 'blocks' || edge.kind === 'waits_for'} d={`M ${x1} ${y1} C ${bend} ${y1}, ${bend} ${y2}, ${x2} ${y2}`} marker-end="url(#graph-arrow)"><title>{edge.from} {relationLabel(edge.kind)} {edge.to}</title></path>
        {/each}
      </svg>
      {#each graph.nodes as node (node.id)}
        {@const position = layout.positions.get(node.id)!}
        {@const wait = (node.fields as { wait?: { reason?: string } } | undefined)?.wait}
        {@const detail = stateDetail(node)}
        {@const step = stepName(node.fields)}
        <article class="graph-card" class:graph-selected={selected === node.id} style:left={position.x + 'px'} style:top={position.y + 'px'} style:width={card.width + 'px'} style:height={card.height + 'px'} data-ticket={node.id} data-status={node.status}>
          <div class="graph-select hosted">
            <button class="hit" aria-label={'Inspect ' + node.id} aria-describedby={'graph-status-' + node.id} aria-pressed={selected === node.id} onclick={() => onselect(node.id)}></button>
            <strong class="graph-title" title={parseTitle(node.title).text}>{#if node.title}<TitleText title={node.title} />{:else}{node.kind}{/if}</strong>
            <span class="graph-meta"><span id={'graph-status-' + node.id} class="ticket-status mono" data-status={node.status} role="img" title={ticketState(node)} aria-label={ticketState(node)}>{statusLetter(node)}</span><span class="mono">{node.id}</span><span>{node.kind}{node.pool ? ' · ' + node.pool : ''}{step ? ' · ' + step : ''}</span></span>
            <span class="graph-meta muted"><time datetime={node.created_at} title={exactLabel(node.created_at)}>{shortLabel(node.created_at)}</time>{#if node.closed_at}<span>→ <time datetime={node.closed_at} title={exactLabel(node.closed_at)}>{shortLabel(node.closed_at)}</time></span>{/if}</span>
          </div>
          {#if detail || wait?.reason}<div class="graph-state" data-status={node.status}>{#if detail}<span>{@html md(detail, true)}</span>{/if}{#if wait?.reason}<div>{@html md(wait.reason, true)}</div>{/if}</div>{/if}
          <div class="graph-links">{#if node.holder || node.kind === 'session'}<button title={node.holder || node.id} onclick={() => onwatch(node.holder || node.id)}>view session ↗</button>{/if}{#each safeLinks(node) as href}<a {href} target="_blank" rel="noopener noreferrer">{linkLabel(href)}</a>{/each}</div>
        </article>
      {/each}
    </div>
    <div class="graph-axis" class:graph-axis-stacked={stacked} aria-label="Creation time axis" role="group">
      {#each ticks as t (t)}<span class="graph-mark" style:left={screenX(t) + 'px'}></span>{/each}
      <span class="graph-mark graph-mark-end" class:graph-now={end.active} style:left={endX + 'px'}></span>
      <span class="graph-mark graph-mark-start" style:left={startX + 'px'}></span>
      {#each ticks as t (t)}{@render stamp(t, relativeLabel(t - layout.start), 'tick', screenX(t), screenX(t))}{/each}
      <span class="graph-stamp-box" class:graph-off-left={startX < startLeft - 1} bind:clientWidth={startW} style:left={startLeft + 'px'}>{@render stamp(layout.start, 'start ' + fullLabel(layout.start), 'start', 0, startLeft + startW / 2)}</span>
      <span class="graph-stamp-box graph-stamp-box-end" class:graph-off-right={endX > endLeft + endW + 1} class:graph-off-left={endX < endLeft - 1} bind:clientWidth={endW} style:left={endLeft + 'px'}>{@render stamp(end.at, (end.active ? 'now ' : 'ended ') + fullLabel(end.at), end.active ? 'now' : 'end', 0, endLeft + endW / 2)}</span>
    </div>
    {#if tip}<div role="tooltip" class="graph-tip" bind:clientWidth={tipW} style:left={Math.max(pad, Math.min(width - pad - tipW, tip.left - tipW / 2)) + 'px'}>{tip.text}</div>{/if}
  </div>
</section>
