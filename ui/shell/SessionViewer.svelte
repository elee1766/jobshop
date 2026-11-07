<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import type { SessionEvent } from '../../packages/jobshop-client/rpc.js';
  import { client } from './lib/api.svelte.js';
  import { sessionActivity } from './lib/session.js';
  import SessionTranscript from './SessionTranscript.svelte';
  let { id, onclose }: { id: string; onclose: () => void } = $props();
  let frames: SessionEvent[] = $state([]);
  let connection = $state('connecting');
  let error = $state('');
  const activity = $derived(sessionActivity(frames));
  let viewport = $state<HTMLDivElement>();
  let closeButton: HTMLButtonElement;
  let epoch = 0;
  $effect(() => {
    const session = id;
    const generation = ++epoch;
    frames = [];
    error = '';
    connection = 'connecting';
    const close = client.session(session, (event, kind) => {
      if (epoch !== generation) return;
      if (kind === 'end') { connection = 'ended'; return; }
      const frame = event as SessionEvent;
      if (frame.seq && frames.some(existing => existing.seq === frame.seq)) return;
      frames = [...frames, frame];
    }, { onState: (value, detail) => {
      if (epoch !== generation || connection === 'ended') return;
      connection = value;
      if (value === 'error') error = detail?.error || 'Unable to read this session.';
    } });
    return () => { epoch++; close(); };
  });
  onMount(() => { untrack(() => closeButton?.focus({ preventScroll: true })); });
</script>

<svelte:window onkeydown={event => { if (event.key === 'Escape') { event.preventDefault(); onclose(); } }} />
<section class="session-viewer" aria-label="Read-only session">
  <div class="tag"><span>session</span><span class="chat-state" class:working={connection === 'connected' && activity === 'working'} role="status">read only · {connection === 'connected' ? activity : connection}</span><button bind:this={closeButton} class="push-right" onclick={onclose}>dismiss ×</button></div>
  <div class="session-identity mono">{id}</div>
  <div class="session-output" bind:this={viewport} aria-live="polite" aria-label="Session transcript">
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if !frames.length && !error}<p class="muted">{connection === 'ended' ? 'No journal messages recorded.' : 'Waiting for session output…'}</p>{/if}
    {#key id}<SessionTranscript bind:frames {viewport} />{/key}
  </div>
</section>
