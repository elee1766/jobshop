<script>
  import { go } from '../lib/url.svelte.js';

  /** A session id that opens its chat in the side panel. */
  /** @type {{ id: string }} */
  let { id } = $props();
</script>

<a href={'?session=' + id} class="mono sl" onclick={(e) => { e.preventDefault(); e.stopPropagation(); go({ session: id, ticket: '' }); }}>{id}</a>

<style>
  .sl { display: inline-block; }
  @media (max-width: 768px) { .sl { padding: 8px 0; } }
</style>
