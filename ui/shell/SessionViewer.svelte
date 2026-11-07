<script lang="ts">
  import { onMount, tick, untrack } from 'svelte';
  import type { SessionEvent } from '../../packages/jobshop-client/rpc.js';
  import { client } from './lib/api.svelte.js';
  import SessionTranscript from './SessionTranscript.svelte';
  let { id, onclose }: { id: string; onclose: () => void } = $props();
  let frames: SessionEvent[] = $state([]);
  let connection = $state('connecting');
  let error = $state('');
  let loading = $state(false);
  let exhausted = $state(false);
  let viewport: HTMLDivElement;
  let closeButton: HTMLButtonElement;
  let epoch = 0;
  $effect(() => {
    const session = id;
    const generation = ++epoch;
    frames = [];
    error = '';
    connection = 'connecting';
    exhausted = false;
    const close = client.session(session, (event, kind) => {
      if (epoch !== generation) return;
      if (kind === 'end') { connection = 'ended'; return; }
      const frame = event as SessionEvent;
      if (frame.seq && frames.some(existing => existing.seq === frame.seq)) return;
      const atBottom = !viewport || viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight < 80;
      frames = [...frames, frame];
      if (atBottom) void tick().then(() => { if (epoch === generation && viewport) viewport.scrollTop = viewport.scrollHeight; });
    }, { onState: (value, detail) => {
      if (epoch !== generation || connection === 'ended') return;
      connection = value;
      if (value === 'error') error = detail?.error || 'Unable to read this session.';
    } });
    return () => { epoch++; close(); };
  });
  async function older() {
    const before = frames.find(frame => (frame.seq || 0) > 0)?.seq;
    if (!before || loading) return;
    const generation = epoch;
    const height = viewport.scrollHeight;
    const top = viewport.scrollTop;
    loading = true;
    try {
      const data = await client.get(`api/session/${encodeURIComponent(id)}/events?before=${before}&limit=200`);
      if (generation !== epoch) return;
      const entries = (data.events || []) as SessionEvent[];
      exhausted = entries.length === 0;
      frames = [...entries.filter(frame => !frames.some(existing => existing.seq === frame.seq)), ...frames];
      await tick();
      if (generation === epoch) viewport.scrollTop = top + viewport.scrollHeight - height;
    } catch (cause) { if (generation === epoch) error = String(cause); }
    finally { loading = false; }
  }
  onMount(() => { untrack(() => closeButton?.focus({ preventScroll: true })); });
</script>

<svelte:window onkeydown={event => { if (event.key === 'Escape') { event.preventDefault(); onclose(); } }} />
<section class="session-viewer" aria-label="Read-only session">
  <div class="tag"><span>session</span><span class="muted">read only · {connection}</span><button bind:this={closeButton} class="push-right" onclick={onclose}>dismiss ×</button></div>
  <div class="session-identity mono">{id}</div>
  <div class="session-output" bind:this={viewport} aria-live="polite" aria-label="Session transcript">
    {#if !exhausted && frames.some(frame => (frame.seq || 0) > 1)}<button disabled={loading} onclick={older}>load older</button>{/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if !frames.length && !error}<p class="muted">{connection === 'ended' ? 'No journal messages recorded.' : 'Waiting for session output…'}</p>{/if}
    <SessionTranscript {frames} />
  </div>
</section>
