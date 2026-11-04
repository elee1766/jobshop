<script>
  import { go } from '../lib/url.svelte.js';
  import { tone, safe } from '../lib/format.js';
  import Trace from './Trace.svelte';

  /** @type {{ node: any }} */
  let { node } = $props();
</script>

<div class="node">
  <a href={'?ticket=' + node.id} class="mono" onclick={(e) => { e.preventDefault(); go({ ticket: node.id }); }}>{node.id}</a>
  <span class="pill {tone(node.status)}">{node.status}</span>
  <span class="t">{node.title}</span>
  {#if node.trigger}<span class="faint">via {node.trigger}</span>{/if}
  {#each node.links ?? [] as l (l)}
    {#if safe(l)}<a href={l} target="_blank" rel="noopener noreferrer" class="faint">{l.replace(/^https:\/\/github\.com\//, '')}</a>{/if}
  {/each}
  {#if node.children?.length}
    <div class="kids">
      {#each node.children as c (c.id)}<Trace node={c} />{/each}
    </div>
  {/if}
</div>

<style>
  .node { margin: 4px 0; display: flex; flex-wrap: wrap; gap: 4px 8px; align-items: baseline; min-width: 0; }
  .node a, .t { overflow-wrap: anywhere; min-width: 0; }
  .kids { flex-basis: 100%; padding-left: 12px; margin-left: 2px; border-left: 1px solid var(--border); min-width: 0; }
  @media (max-width: 720px) { .kids { padding-left: 8px; } .node > a:first-child { padding: 6px 0; } }
</style>
