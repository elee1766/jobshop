<script>
  /** @type {{ id: string }} */
  let { id } = $props();

  /** @typedef {{ kind: string, text: string, time?: string }} Line */
  /** @type {Line[]} */
  let lines = $state([]);
  let status = $state('connecting');
  /** @type {HTMLElement | undefined} */
  let box = $state();

  $effect(() => {
    lines = [];
    status = 'connecting';
    const es = new EventSource('/api/session/' + encodeURIComponent(id) + '/stream?tail=200');
    es.addEventListener('open', () => { status = 'live'; });
    es.addEventListener('session', (e) => {
      const ev = JSON.parse(e.data);
      switch (ev.type) {
        case 'turn_start': push('turn', ev.text || '', ev.time); break;
        case 'text': {
          const last = lines[lines.length - 1];
          if (last && last.kind === 'text') last.text += ev.text;
          else push('text', ev.text);
          break;
        }
        case 'tool': if (ev.tool?.title) push('tool', ev.tool.title + (ev.tool.status ? ' · ' + ev.tool.status : '')); break;
        case 'turn_end': push('end', ev.error ? 'failed: ' + ev.error : ev.stop_reason || 'end'); break;
        case 'state': push('end', 'session ' + ev.status + (ev.resolution ? ' ' + ev.resolution : '')); break;
      }
    });
    es.addEventListener('end', () => { status = 'closed'; es.close(); });
    es.addEventListener('error', () => { if (status !== 'closed') status = 'reconnecting'; });
    return () => es.close();
  });

  /** @param {string} kind @param {string} text @param {string} [time] */
  function push(kind, text, time) {
    lines.push({ kind, text, time });
    if (lines.length > 400) lines.splice(0, lines.length - 400);
  }

  $effect(() => {
    void lines.length;
    if (box) box.scrollTop = box.scrollHeight;
  });
</script>

<div class="wrap">
  <div class="head faint">journal tail · {status}</div>
  <div class="box" bind:this={box} data-testid="journal">
    {#each lines as l, i (i)}
      <div class="line {l.kind}"><pre>{l.text}</pre></div>
    {:else}
      <div class="faint">No turns yet.</div>
    {/each}
  </div>
</div>

<style>
  .wrap { padding: 8px 10px 12px; }
  .head { font-size: 11px; margin-bottom: 6px; }
  .box { max-height: 420px; overflow: auto; border: 1px solid var(--border); border-radius: 6px; padding: 8px 10px; }
  .line { padding: 2px 0; }
  .turn pre { color: var(--accent); }
  .turn { border-top: 1px solid var(--border); margin-top: 6px; padding-top: 6px; }
  .tool pre { color: var(--muted); }
  .end pre { color: var(--faint); font-size: 11px; }
</style>
