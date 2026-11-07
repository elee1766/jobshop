<script lang="ts">
  import Trace from "./Trace.svelte";
  import { ago } from "./model.js";
  import type { TraceNode } from "../../packages/jobshop-client/rpc.js";
  let {
    node,
    onselect,
    origin = 0,
    end = Date.now(),
  }: {
    node: TraceNode;
    onselect: (id: string) => void;
    origin?: number;
    end?: number;
  } = $props();
  const start = $derived(origin || Date.parse(node.created_at));
  const left = $derived(
    Math.max(
      0,
      ((Date.parse(node.created_at) - start) / Math.max(1, end - start)) * 100,
    ),
  );
  const width = $derived(
    Math.max(
      1,
      ((Date.parse(node.closed_at || node.updated_at) -
        Date.parse(node.created_at)) /
        Math.max(1, end - start)) *
        100,
    ),
  );
</script>

<div class="trace-node">
  <button class="mono" onclick={() => onselect(node.id)}>{node.id}</button>
  {node.title}<span class="muted">
    {node.status} {node.resolution || ""} · {ago(node.created_at)}</span
  >
  <div class="waterfall" aria-label="Ticket duration">
    <span
      style:left={left + "%"}
      style:width={Math.min(width, 100 - left) + "%"}
    ></span>
  </div>
  {#if node.holder}<button
      class="muted mono"
      onclick={() => onselect(node.holder!)}
      >session {node.session?.display_name || node.holder}</button
    >{/if}
  {#if node.children?.length}<div class="trace-children">
      {#each node.children as child (child.id)}<Trace
          node={child}
          {onselect}
          origin={start}
          {end}
        />{/each}
    </div>{/if}
</div>
