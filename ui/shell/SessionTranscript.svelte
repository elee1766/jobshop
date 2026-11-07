<script lang="ts">
  import { onDestroy, onMount, tick, untrack } from 'svelte';
  import { SvelteSet } from 'svelte/reactivity';
  import { Virtualizer, elementScroll, observeElementOffset, observeElementRect } from '@tanstack/virtual-core';
  import type { SessionEvent } from '../../packages/jobshop-client/rpc.js';
  import type { PendingMessage } from './lib/session.js';
  import { client } from './lib/api.svelte.js';
  import { md } from '../kit/markdown.js';

  let { frames = $bindable(), viewport, pending = [] }: {
    frames: SessionEvent[];
    viewport: HTMLDivElement | undefined;
    pending?: PendingMessage[];
  } = $props();
  type Message = SessionEvent & { key: string; pending?: PendingMessage };
  let before: number | null = $state(null);
  let loading = $state(false);
  let historyError = $state('');
  let started = $state(false);
  let starting = false;
  let alive = true;
  const abort = new AbortController();
  const expanded = new SvelteSet<string>();
  // Keep streamed rows stable when an older page completes a partial row.
  const rowKeys = new Map<number, string>();
  onDestroy(() => { alive = false; abort.abort(); });

  const messages = $derived.by(() => {
    const out: Message[] = [];
    const tools = new Map<string, Message>();
    const delivered = new Set(frames.flatMap(frame => (frame.messages || []).map(message => message.id)));
    let text: Message | undefined;
    for (const frame of frames) {
      if (frame.type === 'text' && text) {
        text.text = (text.text || '') + (frame.text || '');
        text.key = rowKeys.get(frame.seq!) || text.key;
        rowKeys.set(frame.seq!, text.key);
        continue;
      }
      if (frame.type === 'turn_start' || frame.type === 'turn_end') tools.clear();
      if (['turn_start', 'turn_end', 'tool', 'image'].includes(frame.type) || frame.error) text = undefined;
      if (frame.type === 'tool' && frame.tool) {
        const previous = tools.get(frame.tool.id);
        if (previous) {
          previous.tool = { ...previous.tool!, ...frame.tool };
          previous.key = rowKeys.get(frame.seq!) || previous.key;
          rowKeys.set(frame.seq!, previous.key);
          if (frame.images?.length) previous.images = frame.images;
          continue;
        }
      }
      if (['turn_start', 'text', 'tool', 'image'].includes(frame.type) || frame.error) {
        const message = { ...frame, key: rowKeys.get(frame.seq!) || 'frame-' + frame.seq };
        rowKeys.set(frame.seq!, message.key);
        out.push(message);
        if (frame.tool) tools.set(frame.tool.id, message);
        if (frame.type === 'text') text = message;
      }
    }
    for (const message of pending) {
      if (!delivered.has(message.id)) out.push({ session: '', type: 'pending', key: message.id, pending: message });
    }
    return out;
  });
  let revision = $state(0);
  function refresh() {
    revision++;
    // Anchor writes need the new spacer height committed to the DOM first.
    void tick().then(() => { if (alive) virtualizer._willUpdate(); });
  }
  const virtualizer = new Virtualizer<HTMLDivElement, HTMLDivElement>({
    observeElementRect,
    observeElementOffset,
    scrollToFn: elementScroll,
    onChange: refresh,
    count: 0,
    getScrollElement: () => null,
    estimateSize: () => 90,
    paddingStart: 28,
    anchorTo: 'end',
    followOnAppend: true,
    scrollEndThreshold: 80,
    overscan: 6,
  });
  onMount(() => virtualizer._didMount());
  const view = $derived.by(() => {
    revision;
    return { items: virtualizer.getVirtualItems(), height: virtualizer.getTotalSize(), offset: virtualizer.scrollOffset || 0, direction: virtualizer.scrollDirection };
  });
  $effect(() => {
    const rows = messages;
    const parent = viewport;
    untrack(() => {
      virtualizer.setOptions({
        ...virtualizer.options,
        count: rows.length,
        getScrollElement: () => parent || null,
        getItemKey: index => rows[index].key,
      });
      refresh();
    });
  });
  $effect(() => {
    if (before !== null) return;
    const first = frames.find(frame => frame.first)?.first ?? frames.find(frame => frame.seq)?.seq;
    if (first) before = first;
  });
  $effect(() => {
    if (!viewport || !messages.length || started || starting) return;
    starting = true;
    void tick().then(() => {
      if (!alive) return;
      virtualizer.scrollToEnd();
      started = true;
    });
  });
  $effect(() => {
    if (started && view.direction === 'backward' && before !== null && before > 1 && !loading && !historyError && view.offset < 160) {
      void untrack(older);
    }
  });
  async function older() {
    if (before === null || before <= 1 || loading) return;
    loading = true;
    historyError = '';
    let cursor = before;
    try {
      let entries: SessionEvent[] = [];
      // A page of protocol bookkeeping is not the beginning of the journal.
      do {
        const data = await client.get(`api/session/${encodeURIComponent(frames[0].session)}/events?before=${cursor}&limit=200`, { signal: abort.signal });
        if (!alive) return;
        if (!Number.isInteger(data.first) || data.first < 1 || data.first >= cursor) throw new Error('History cursor did not advance.');
        cursor = data.first;
        entries = [...(data.events || []), ...entries];
      } while (cursor > 1 && !entries.some(frame => ['turn_start', 'text', 'tool', 'image'].includes(frame.type) || frame.error));
      const seen = new Set(frames.map(frame => frame.seq));
      frames = [...entries.filter(frame => !seen.has(frame.seq)), ...frames];
      before = cursor;
    } catch (cause) {
      if (alive) historyError = String(cause);
    } finally {
      if (alive) loading = false;
    }
  }
  function remember(event: Event, key: string) {
    if ((event.currentTarget as HTMLDetailsElement).open) expanded.add(key);
    else expanded.delete(key);
  }
  function measure(node: HTMLDivElement) {
    virtualizer.measureElement(node);
    return { destroy: () => virtualizer.measureElement(null) };
  }
</script>

<div class="virtual-transcript" style:height={view.height + 'px'}>
  <div class="history-loader">
    {#if historyError}<button onclick={older}>retry history</button><span class="error" role="alert" title={historyError}>{historyError}</span>
    {:else if loading}<span role="status">loading history…</span>
    {:else if before !== null && before > 1}<button onclick={older}>load older</button>
    {:else if before === 1}<span class="muted">start of conversation</span>{/if}
  </div>
  {#each view.items as row (row.key)}
    {@const message = messages[row.index]}
    <div class="transcript-row" data-index={row.index} data-row-key={row.key} use:measure style:transform={'translateY(' + row.start + 'px)'}>
        {#if message.pending}
          <div class="operator" data-message={message.pending.id}>{message.pending.text}{#if message.pending.attachments.length}<small>{message.pending.attachments.join(', ')}</small>{/if}<span class="message-delivery muted" role="status">{message.pending.state === 'sending' ? 'sending…' : 'queued'}</span></div>
        {:else if message.type === 'turn_start'}
          {#if message.context}<details class="controller-context" open={expanded.has(message.key + ':context')} ontoggle={event => remember(event, message.key + ':context')}><summary>controller context</summary><div class="markdown">{@html md(message.context)}</div></details>{/if}
          {#if message.raw_prompt}
            {#each message.messages || [] as original (original.id)}
              <div class="operator" data-message={original.id} title={'from ' + original.from}>{original.text}</div>
            {/each}
            <details class="raw-prompt" open={expanded.has(message.key + ':raw')} ontoggle={event => remember(event, message.key + ':raw')}><summary>raw model prompt</summary><pre>{message.raw_prompt}</pre></details>
          {:else if message.text}
            <details class="controller-context legacy-prompt" open={expanded.has(message.key + ':legacy')} ontoggle={event => remember(event, message.key + ':legacy')}><summary>legacy prompt</summary><pre>{message.text}</pre></details>
          {/if}
          {#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="User attachment" loading="lazy" />{/each}
        {:else if message.tool}
          <details class="tool mono" open={expanded.has(message.key + ':tool')} ontoggle={event => remember(event, message.key + ':tool')}><summary>{message.tool.intent || message.tool.title || 'tool'} · {message.tool.status}</summary>
            <pre>{[message.tool.target, message.tool.input, message.tool.output].filter(Boolean).join('\n')}</pre>
            {#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="Tool output" loading="lazy" />{/each}
          </details>
        {:else}
          <div class="markdown" class:error={!!message.error}>{@html md(message.text || message.error || '')}{#each message.images || [] as image}<img class="transcript-image" src={'data:' + image.mime_type + ';base64,' + image.data} alt="Agent output" loading="lazy" />{/each}</div>
        {/if}
    </div>
  {/each}
</div>
