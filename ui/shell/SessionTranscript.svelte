<script lang="ts">
  import type { SessionEvent } from '../../packages/jobshop-client/rpc.js';
  import { md } from '../kit/markdown.js';
  let { frames }: { frames: SessionEvent[] } = $props();
  const messages = $derived.by(() => {
    const out: SessionEvent[] = [];
    for (const frame of frames) {
      if (frame.type === 'tool' && frame.tool) {
        const previous = out.find(m => m.tool?.id === frame.tool?.id);
        if (previous) {
          previous.tool = { ...previous.tool!, ...frame.tool };
          if (frame.images?.length) previous.images = frame.images;
          continue;
        }
      }
      if (frame.type === 'text' && out.at(-1)?.type === 'text') out[out.length - 1].text = (out.at(-1)?.text || '') + (frame.text || '');
      else if (['turn_start', 'text', 'tool', 'image'].includes(frame.type) || frame.error) out.push({ ...frame });
    }
    return out;
  });
</script>

{#each messages as message, index (message.seq || 'event-' + index)}
  {#if message.type === 'turn_start'}
    {#if message.context}<details class="controller-context"><summary>controller context</summary><div class="markdown">{@html md(message.context)}</div></details>{/if}
    {#if message.raw_prompt}
      {#each message.messages || [] as original (original.id)}
        <div class="operator" data-message={original.id} title={'from ' + original.from}>{original.text}</div>
      {/each}
      <details class="raw-prompt"><summary>raw model prompt</summary><pre>{message.raw_prompt}</pre></details>
    {:else if message.text}
      <details class="controller-context legacy-prompt"><summary>legacy prompt</summary><pre>{message.text}</pre></details>
    {/if}
    {#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="User attachment" loading="lazy" />{/each}
  {:else if message.tool}
    <details class="tool mono"><summary>{message.tool.intent || message.tool.title || 'tool'} · {message.tool.status}</summary>
      <pre>{message.tool.target || ''}\n{message.tool.input || ''}\n{message.tool.output || ''}</pre>
      {#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="Tool output" loading="lazy" />{/each}
    </details>
  {:else}
    <div class="markdown" class:error={!!message.error}>{@html md(message.text || message.error || '')}{#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="Agent output" loading="lazy" />{/each}</div>
  {/if}
{/each}
