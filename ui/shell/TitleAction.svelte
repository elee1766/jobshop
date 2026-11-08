<script lang="ts">
  // A ticket or run title that is also the row's action (JOB-135).
  //
  // Plain title: the same button or link as before. Title with Markdown
  // links: a link cannot sit inside a button or another link, so the action
  // becomes an empty overlay (one tab stop, named by the plain text) behind
  // the title text, and each title link is its own control above it. A click
  // on text falls through to the overlay, a click on a link goes only to the
  // link.
  import type { Snippet } from 'svelte';
  import { parseTitle } from '../kit/titles.js';
  import TitleText from './TitleText.svelte';

  let { title, fallback = '', class: cls = '', tip, href, onactivate, children }: {
    title?: string;
    fallback?: string;
    class?: string;
    tip?: string;
    // Set for a link action, omitted for a button action.
    href?: string;
    onactivate?: (event: MouseEvent) => void;
    // Trailing content inside the action: ids, ages.
    children?: Snippet;
  } = $props();

  const parsed = $derived(parseTitle(title || fallback));
</script>

{#if !parsed.links.length}
  {#if href !== undefined}<a class={cls} title={tip} {href}><span>{parsed.text}</span>{@render children?.()}</a>
  {:else}<button class={cls} title={tip} onclick={onactivate}><span>{parsed.text}</span>{@render children?.()}</button>{/if}
{:else}
  <span class="{cls} hosted" title={tip}>
    {#if href !== undefined}<a class="hit" {href} aria-label={parsed.text}> </a>
    {:else}<button class="hit" aria-label={parsed.text} onclick={onactivate}></button>{/if}
    <span><TitleText {title} {fallback} /></span>{@render children?.()}
  </span>
{/if}
