<script>
  import { post } from '../lib/api.svelte.js';
  import { safe } from '../lib/format.js';

  /**
   * One session's conversation: turns stream from session.subscribe, and
   * when `key` or `write` is set a composer sends with session.send.
   * @type {{ id: string, key?: string, write?: boolean, reason?: string, tail?: number }}
   */
  let { id, key = '', write = false, reason = '', tail = 200 } = $props();

  /** @typedef {{ kind: string, text: string }} Line */
  /** @type {Line[]} */
  let lines = $state([]);
  let status = $state('connecting');
  let busy = $state(false);
  let text = $state('');
  let sending = $state(false);
  let error = $state('');
  /** @type {HTMLElement | undefined} */
  let box = $state();

  $effect(() => {
    lines = [];
    status = 'connecting';
    const es = new EventSource('/api/session/' + encodeURIComponent(id) + '/stream?tail=' + tail);
    es.addEventListener('open', () => { status = 'live'; });
    es.addEventListener('session', (e) => {
      const ev = JSON.parse(e.data);
      switch (ev.type) {
        case 'turn_start': busy = true; push('turn', ev.text || ''); break;
        case 'text': {
          const last = lines[lines.length - 1];
          if (last && last.kind === 'text') last.text += ev.text;
          else push('text', ev.text);
          break;
        }
        case 'tool': if (ev.tool?.title) push('tool', ev.tool.title + (ev.tool.status ? ' · ' + ev.tool.status : '')); break;
        case 'turn_end': busy = false; push('end', ev.error ? 'failed: ' + ev.error : ev.stop_reason || 'end'); break;
        case 'state': push('end', 'session ' + ev.status + (ev.resolution ? ' ' + ev.resolution : '')); break;
      }
    });
    es.addEventListener('end', () => { status = 'closed'; busy = false; es.close(); });
    es.addEventListener('error', () => { if (status !== 'closed') status = 'reconnecting'; });
    return () => es.close();
  });

  /** @param {string} kind @param {string} t */
  function push(kind, t) {
    lines.push({ kind, text: t });
    if (lines.length > 400) lines.splice(0, lines.length - 400);
  }

  $effect(() => {
    void lines.length;
    if (box) box.scrollTop = box.scrollHeight;
  });

  /** @param {boolean} interrupt */
  async function send(interrupt) {
    if (!text.trim() || sending) return;
    sending = true;
    error = '';
    try {
      await post('/api/chat/send', key ? { key, text, interrupt } : { session: id, text, interrupt });
      text = '';
    } catch (e) {
      error = String(/** @type {Error} */ (e).message || e);
    } finally {
      sending = false;
    }
  }

  /** Splits text into plain parts and http(s) links, rendered as text and anchors. @param {string} t */
  function parts(t) {
    /** @type {{ text: string, href?: string }[]} */
    const out = [];
    let last = 0;
    for (const m of t.matchAll(/https?:\/\/[^\s<>()"']+/g)) {
      const i = m.index ?? 0;
      if (i > last) out.push({ text: t.slice(last, i) });
      out.push(safe(m[0]) ? { text: m[0], href: m[0] } : { text: m[0] });
      last = i + m[0].length;
    }
    if (last < t.length) out.push({ text: t.slice(last) });
    return out;
  }

  const canSend = $derived(write || !!key);
</script>

<div class="chat">
  <div class="head faint">
    {status}{busy ? ' · working' : ''}{key ? ' · ' + key : ''}
  </div>
  <div class="box" bind:this={box} data-testid="journal">
    {#each lines as l, i (i)}
      <div class="line {l.kind}">
        <pre>{#each parts(l.text) as p, j (j)}{#if p.href}<a href={p.href} target="_blank" rel="noopener noreferrer">{p.text}</a>{:else}{p.text}{/if}{/each}</pre>
      </div>
    {:else}
      <div class="faint">No turns yet.</div>
    {/each}
  </div>
  {#if canSend}
    <form class="composer" onsubmit={(e) => { e.preventDefault(); send(false); }}>
      <textarea
        bind:value={text}
        rows="2"
        placeholder="Message the session (Enter sends, Shift+Enter for a new line)"
        data-testid="chat-input"
        onkeydown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send(false); } }}
      ></textarea>
      <div class="actions">
        <button type="submit" disabled={sending || !text.trim()} data-testid="chat-send">Send</button>
        <button type="button" disabled={sending || !text.trim() || !busy} onclick={() => send(true)} title="Cancel the current turn and send">Interrupt</button>
      </div>
    </form>
    {#if error}<div class="err">{error}</div>{/if}
  {:else if reason}
    <div class="faint ro">Read-only: {reason}</div>
  {/if}
</div>

<style>
  .chat { padding: 8px 10px 12px; }
  .head { font-size: 11px; margin-bottom: 6px; }
  .box { max-height: 460px; overflow: auto; border: 1px solid var(--border); border-radius: 6px; padding: 8px 10px; background: var(--bg); }
  .line { padding: 2px 0; }
  .turn { border-top: 1px solid var(--border); margin-top: 6px; padding-top: 6px; }
  .turn pre { color: var(--accent); }
  .tool pre { color: var(--muted); }
  .end pre { color: var(--faint); font-size: 11px; }
  .composer { display: flex; gap: 8px; margin-top: 8px; align-items: flex-end; }
  textarea {
    flex: 1; font: inherit; color: inherit; background: var(--raised); border: 1px solid var(--border);
    border-radius: 6px; padding: 6px 8px; resize: vertical;
  }
  textarea:focus { outline: none; border-color: var(--accent); }
  .actions { display: flex; flex-direction: column; gap: 4px; }
  button:disabled { opacity: 0.5; cursor: default; }
  .err { color: var(--bad); margin-top: 6px; }
  .ro { margin-top: 6px; font-size: 12px; }
</style>
