<script lang="ts">
  // Renders a title's text with its Markdown links as real anchors (JOB-135).
  // No wrapper element and no row behavior: callers that are not themselves a
  // button or link (headings, graph cards) use this directly.
  import { parseTitle } from '../kit/titles.js';

  let { title, fallback = '' }: { title?: string; fallback?: string } = $props();
  const parsed = $derived(parseTitle(title || fallback));
</script>

{#each parsed.parts as part}{#if part.link}<a href={part.link.href} target={part.link.external ? '_blank' : undefined} rel={part.link.external ? 'noopener noreferrer' : undefined}>{part.text}</a>{:else}{part.text}{/if}{/each}
