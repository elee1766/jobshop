<script>
  import { tick as afterRender } from 'svelte';
  import { post } from '../lib/api.svelte.js';
  import { safe } from '../lib/format.js';
  import ArrowUp from '@lucide/svelte/icons/arrow-up';
  import ArrowDown from '@lucide/svelte/icons/arrow-down';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import Check from '@lucide/svelte/icons/check';
  import X from '@lucide/svelte/icons/x';
  import Circle from '@lucide/svelte/icons/circle';
  import ChevronRight from '@lucide/svelte/icons/chevron-right';
  import Wrench from '@lucide/svelte/icons/wrench';
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import Square from '@lucide/svelte/icons/square';

  /**
   * One session's conversation, from the session stream: operator prompts,
   * the agent's text per turn, tool calls updated in place, and a composer
   * when `key` or `write` is set. The parent gives it a height.
   * @type {{ id: string, key?: string, write?: boolean, reason?: string, tail?: number }}
   */
  let { id, key = '', write = false, reason = '', tail = 200 } = $props();

  /** @typedef {{ id: string, title: string, status: string }} Call */
  /** @typedef {{ kind: 'text', text: string, time: string }} TextPart */
  /** @typedef {{ kind: 'tools', calls: Call[] }} ToolsPart */
  /** @typedef {{ kind: 'turn', prompt: string, time: string, parts: (TextPart | ToolsPart)[], end: { error: string, stop: string } | null, long: boolean }} Turn */
  /** @typedef {{ kind: 'system', text: string, time: string }} System */

  /** @type {(Turn | System)[]} */
  let items = $state([]);
  let status = $state('connecting');
  let busy = $state(false);
  let lastTextAt = $state(0);
  let now = $state(Date.now());
  let text = $state('');
  let sending = $state(false);
  let error = $state('');
  let unread = $state(false);
  /** @type {HTMLElement | undefined} */
  let box = $state();
  /** @type {HTMLTextAreaElement | undefined} */
  let input = $state();
  let stuck = true;
  let version = $state(0);

  /** @param {string | undefined} t */
  const stamp = (t) => t || new Date().toISOString();

  /** The turn that events belong to; a stream tail can start mid-turn. @param {string} time */
  function turnFor(time) {
    const last = items[items.length - 1];
    if (last && last.kind === 'turn') return last;
    items.push({ kind: 'turn', prompt: '', time, parts: [], end: null, long: false });
    return /** @type {Turn} */ (items[items.length - 1]);
  }

  /** @param {any} ev */
  function apply(ev) {
    const time = stamp(ev.time);
    switch (ev.type) {
      case 'turn_start':
        busy = true;
        lastTextAt = Date.now();
        items.push({ kind: 'turn', prompt: ev.text || '', time, parts: [], end: null, long: false });
        break;
      case 'text': {
        const t = turnFor(time);
        const last = t.parts[t.parts.length - 1];
        if (last && last.kind === 'text') last.text += ev.text;
        else t.parts.push({ kind: 'text', text: ev.text || '', time });
        lastTextAt = Date.now();
        break;
      }
      case 'tool': {
        const tl = ev.tool;
        if (!tl || (!tl.id && !tl.title)) break;
        const t = turnFor(time);
        const cid = tl.id || tl.title;
        let found = null;
        for (const p of t.parts) {
          if (p.kind !== 'tools') continue;
          found = p.calls.find((c) => c.id === cid) ?? found;
        }
        if (found) {
          if (tl.title) found.title = tl.title;
          if (tl.status) found.status = tl.status;
        } else {
          const call = { id: cid, title: tl.title || cid, status: tl.status || 'pending' };
          const last = t.parts[t.parts.length - 1];
          if (last && last.kind === 'tools') last.calls.push(call);
          else t.parts.push({ kind: 'tools', calls: [call] });
        }
        break;
      }
      case 'turn_end': {
        const t = turnFor(time);
        t.end = { error: ev.error || '', stop: ev.stop_reason || '' };
        busy = false;
        break;
      }
      case 'state':
        items.push({ kind: 'system', text: 'Session ' + ev.status + (ev.resolution ? ' · ' + ev.resolution : ''), time });
        break;
      default:
        return;
    }
    if (items.length > 300) items.splice(0, items.length - 300);
    version++;
  }

  $effect(() => {
    items = [];
    status = 'connecting';
    busy = false;
    const es = new EventSource('/api/session/' + encodeURIComponent(id) + '/stream?tail=' + tail);
    es.addEventListener('open', () => { status = 'live'; });
    es.addEventListener('session', (e) => apply(JSON.parse(e.data)));
    es.addEventListener('end', () => { status = 'closed'; busy = false; es.close(); });
    es.addEventListener('error', () => { if (status !== 'closed') status = 'reconnecting'; });
    return () => es.close();
  });

  // The "typing" state needs a clock while a turn is open.
  $effect(() => {
    if (!busy) return;
    now = Date.now();
    const h = window.setInterval(() => { now = Date.now(); }, 500);
    return () => window.clearInterval(h);
  });

  const lastTurn = $derived.by(() => {
    for (let i = items.length - 1; i >= 0; i--) if (items[i].kind === 'turn') return /** @type {Turn} */ (items[i]);
    return null;
  });
  const running = $derived.by(() => {
    if (!busy || !lastTurn) return null;
    let r = null;
    for (const p of lastTurn.parts) if (p.kind === 'tools') for (const c of p.calls) if (c.status === 'in_progress' || c.status === 'pending') r = c;
    return r;
  });
  const typing = $derived(busy && !running && now - lastTextAt > 1500);

  // Follow new content only while the reader is at the bottom.
  $effect(() => {
    void version;
    void typing;
    void running;
    if (!box) return;
    if (stuck) afterRender().then(() => { if (box) box.scrollTop = box.scrollHeight; });
    else unread = true;
  });

  function onScroll() {
    if (!box) return;
    stuck = box.scrollHeight - box.scrollTop - box.clientHeight < 48;
    if (stuck) unread = false;
  }

  function jump() {
    stuck = true;
    unread = false;
    if (box) box.scrollTo({ top: box.scrollHeight, behavior: 'smooth' });
  }

  function grow() {
    if (!input) return;
    input.style.height = 'auto';
    input.style.height = Math.min(input.scrollHeight, 168) + 'px';
  }

  /** @param {boolean} interrupt */
  async function send(interrupt) {
    if (!text.trim() || sending) return;
    sending = true;
    error = '';
    try {
      await post('/api/chat/send', key ? { key, text, interrupt } : { session: id, text, interrupt });
      text = '';
      stuck = true;
      await afterRender();
      grow();
      input?.focus();
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

  /** @param {string} t */
  const hm = (t) => new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

  /** @param {string} t */
  function dayOf(t) {
    const d = new Date(t);
    const today = new Date();
    const y = new Date(today);
    y.setDate(today.getDate() - 1);
    if (d.toDateString() === today.toDateString()) return 'Today';
    if (d.toDateString() === y.toDateString()) return 'Yesterday';
    return d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });
  }

  /** Items with a day label where the day changes (only when the log spans days). */
  const rows = $derived.by(() => {
    const days = new Set(items.map((it) => new Date(it.time).toDateString()));
    let prev = '';
    return items.map((it) => {
      const d = new Date(it.time).toDateString();
      const label = days.size > 1 && d !== prev ? dayOf(it.time) : '';
      prev = d;
      return { it, label };
    });
  });

  /** @param {Call[]} calls */
  function summary(calls) {
    const run = calls.find((c) => c.status === 'in_progress' || c.status === 'pending');
    const bad = calls.filter((c) => c.status === 'failed').length;
    const n = calls.length + (calls.length === 1 ? ' tool call' : ' tool calls');
    return { run, bad, label: run ? run.title : n + (bad ? ' · ' + bad + ' failed' : '') };
  }

  const canSend = $derived(write || !!key);
</script>

<div class="chat">
  <div class="head faint">
    <span class="dot" class:ok={status === 'live'} class:warn={status === 'reconnecting'}></span>
    <span>{status}{busy ? ' · working' : ''}{key ? ' · ' + key : ''}</span>
  </div>
  <div class="wrap">
    <div class="box" bind:this={box} onscroll={onScroll} data-testid="journal">
      {#each rows as { it, label } (it)}
        {#if label}<div class="day"><span>{label}</span></div>{/if}
        {#if it.kind === 'system'}
          <div class="system faint" title={it.time}>{it.text}</div>
        {:else}
          <section class="turn-group">
            {#if it.prompt}
              <div class="line turn me">
                <div class="bubble" class:clamp={it.prompt.length > 600 && !it.long}>
                  <pre class="prose">{#each parts(it.prompt) as p, j (j)}{#if p.href}<a href={p.href} target="_blank" rel="noopener noreferrer">{p.text}</a>{:else}{p.text}{/if}{/each}</pre>
                </div>
                {#if it.prompt.length > 600}
                  <button class="ghost more" onclick={() => (it.long = !it.long)}>{it.long ? 'Show less' : 'Show more'}</button>
                {/if}
                <time class="when faint">{hm(it.time)}</time>
              </div>
            {/if}
            {#each it.parts as p, k (k)}
              {#if p.kind === 'text'}
                <div class="line text">
                  <div class="msg"><pre class="prose">{#each parts(p.text) as q, j (j)}{#if q.href}<a href={q.href} target="_blank" rel="noopener noreferrer">{q.text}</a>{:else}{q.text}{/if}{/each}</pre></div>
                  <time class="when faint">{hm(p.time)}</time>
                </div>
              {:else}
                {@const s = summary(p.calls)}
                <details class="line tools">
                  <summary>
                    <ChevronRight size={12} class="chev" />
                    {#if s.run}<LoaderCircle size={12} class="spin st-run" />
                    {:else if s.bad}<X size={12} class="st-bad" />
                    {:else}<Check size={12} class="st-ok" />{/if}
                    <Wrench size={12} class="faint" />
                    <span class="sum">{s.label}</span>
                  </summary>
                  <ul>
                    {#each p.calls as c (c.id)}
                      <li>
                        {#if c.status === 'completed'}<Check size={12} class="st-ok" />
                        {:else if c.status === 'failed'}<X size={12} class="st-bad" />
                        {:else if c.status === 'in_progress' || c.status === 'pending'}<LoaderCircle size={12} class="spin st-run" />
                        {:else}<Circle size={12} class="faint" />{/if}
                        <span class="t">{c.title}</span>
                        <span class="faint">{c.status}</span>
                      </li>
                    {/each}
                  </ul>
                </details>
              {/if}
            {/each}
            {#if it.end?.error}
              <div class="notice bad"><CircleAlert size={13} /> <span>Turn failed: {it.end.error}</span></div>
            {:else if it.end?.stop && it.end.stop !== 'end_turn'}
              <div class="system faint">stopped: {it.end.stop}</div>
            {/if}
          </section>
        {/if}
      {:else}
        <div class="faint none">No turns yet.</div>
      {/each}
      {#if running}
        <div class="typing" data-testid="chat-working"><LoaderCircle size={12} class="spin st-run" /> working: {running.title}</div>
      {:else if typing}
        <div class="typing" data-testid="chat-typing" aria-label="agent is typing"><span class="dots"><i></i><i></i><i></i></span></div>
      {/if}
    </div>
    {#if unread}
      <button class="jump" onclick={jump}><ArrowDown size={12} /> New messages</button>
    {/if}
  </div>
  {#if canSend}
    <form class="composer" onsubmit={(e) => { e.preventDefault(); send(false); }}>
      <textarea
        bind:this={input}
        bind:value={text}
        rows="1"
        enterkeyhint="send"
        placeholder="Message the session"
        data-testid="chat-input"
        oninput={grow}
        onkeydown={(e) => { if (e.key === 'Enter' && !e.shiftKey && !e.isComposing) { e.preventDefault(); send(false); } }}
      ></textarea>
      {#if busy}
        <button type="button" class="stop" disabled={sending || !text.trim()} onclick={() => send(true)} title="Cancel the current turn and send" aria-label="Interrupt and send"><Square size={12} /> <span class="lbl">Interrupt</span></button>
      {/if}
      <button type="submit" class="primary icon send" disabled={sending || !text.trim()} data-testid="chat-send" aria-label="Send"><ArrowUp size={14} /></button>
    </form>
    {#if error}<div class="err">{error}</div>{/if}
  {:else}
    <div class="composer ro faint">Read-only: {reason || 'this session does not accept input'}</div>
  {/if}
</div>

<style>
  .chat { display: flex; flex-direction: column; height: 100%; min-height: 0; }
  .head { display: flex; align-items: center; gap: 6px; font-size: 11px; padding: 6px 12px; border-bottom: 1px solid var(--border); flex: none; }
  .wrap { position: relative; flex: 1; min-height: 0; display: flex; }
  .box { flex: 1; min-width: 0; overflow-y: auto; overscroll-behavior: contain; padding: 12px 14px 8px; display: flex; flex-direction: column; gap: 10px; }
  .none { text-align: center; padding: 24px 0; }
  .turn-group { display: flex; flex-direction: column; gap: 6px; }
  .day { display: flex; align-items: center; gap: 10px; color: var(--faint); font-size: 11px; }
  .day::before, .day::after { content: ''; flex: 1; height: 1px; background: var(--border); }
  .system { text-align: center; font-size: 11px; }
  .line { position: relative; min-width: 0; }
  .when { display: block; font-size: 10.5px; margin-top: 2px; opacity: 0; transition: opacity 0.12s; }
  .line:hover .when { opacity: 1; }
  @media (hover: none) { .when { opacity: 1; } }

  .turn.me { align-self: flex-end; max-width: min(86%, 560px); display: flex; flex-direction: column; align-items: flex-end; }
  .bubble { background: var(--accent-soft); border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent); border-radius: 12px 12px 3px 12px; padding: 7px 11px; max-width: 100%; }
  .bubble.clamp { max-height: 9.5em; overflow: hidden; -webkit-mask-image: linear-gradient(#000 70%, transparent); mask-image: linear-gradient(#000 70%, transparent); }
  .more { height: 22px; font-size: 11px; padding: 0 6px; margin-top: 2px; }
  .prose { font-family: var(--font); font-size: 13px; line-height: 1.5; white-space: pre-wrap; overflow-wrap: anywhere; }
  .text { max-width: min(100%, 720px); }
  .msg { padding: 2px 2px; }

  .tools { font-size: 12px; max-width: min(100%, 720px); }
  .tools summary {
    list-style: none; display: inline-flex; align-items: center; gap: 6px; cursor: pointer; max-width: 100%;
    padding: 3px 8px; border: 1px solid var(--border); border-radius: 999px; color: var(--muted); background: var(--panel);
    min-height: 24px;
  }
  .tools summary::-webkit-details-marker { display: none; }
  .tools summary:hover { background: var(--hover); }
  .sum { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
  :global(.chev) { transition: transform 0.12s; flex: none; }
  .tools[open] > summary :global(.chev) { transform: rotate(90deg); }
  .tools ul { list-style: none; margin: 4px 0 0; padding: 4px 0 4px 10px; border-left: 1px solid var(--border); margin-left: 10px; }
  .tools li { display: flex; align-items: center; gap: 6px; padding: 2px 0; min-width: 0; }
  .tools .t { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; font-family: var(--mono); font-size: 11.5px; }
  .tools li .faint:last-child { margin-left: auto; flex: none; font-size: 11px; padding-left: 6px; }
  :global(.st-ok) { color: var(--ok); flex: none; }
  :global(.st-bad) { color: var(--bad); flex: none; }
  :global(.st-run) { color: var(--accent); flex: none; }

  .notice { display: flex; gap: 6px; align-items: flex-start; padding: 6px 10px; border-radius: var(--radius); font-size: 12px; overflow-wrap: anywhere; }
  .notice.bad { color: var(--bad); background: color-mix(in srgb, var(--bad) 10%, transparent); border: 1px solid color-mix(in srgb, var(--bad) 25%, transparent); }

  .typing { display: flex; align-items: center; gap: 6px; color: var(--muted); font-size: 12px; min-height: 22px; }
  .dots { display: inline-flex; gap: 3px; padding: 6px 10px; background: var(--raised); border-radius: 999px; }
  .dots i { width: 5px; height: 5px; border-radius: 50%; background: var(--muted); animation: blink 1.2s infinite ease-in-out; }
  .dots i:nth-child(2) { animation-delay: 0.15s; }
  .dots i:nth-child(3) { animation-delay: 0.3s; }
  @keyframes blink { 0%, 80%, 100% { opacity: 0.25; transform: translateY(0); } 40% { opacity: 1; transform: translateY(-2px); } }
  @media (prefers-reduced-motion: reduce) { .dots i { animation: none; opacity: 0.6; } }

  .jump {
    position: absolute; left: 50%; bottom: 10px; transform: translateX(-50%); border-radius: 999px;
    background: var(--accent); color: var(--accent-fg); border-color: transparent; box-shadow: var(--shadow); height: 28px;
  }
  .composer { flex: none; display: flex; gap: 6px; align-items: flex-end; padding: 8px 10px; border-top: 1px solid var(--border); background: var(--bg); }
  .composer textarea { flex: 1; min-width: 0; resize: none; max-height: 168px; line-height: 1.45; padding: 7px 10px; border-radius: 10px; }
  .composer button { height: 32px; flex: none; }
  .composer .send { width: 32px; border-radius: 50%; }
  .ro { justify-content: center; font-size: 12px; padding: 12px; align-items: center; }
  .err { color: var(--bad); padding: 0 12px 8px; font-size: 12px; }

  @media (max-width: 768px) {
    .box { padding: 10px 10px 6px; }
    .composer { padding: 8px 8px 8px; }
    .composer textarea { min-height: 40px; }
    .composer button { height: 40px; min-height: 40px; }
    .composer .send { width: 40px; }
    .tools summary { min-height: 32px; }
    .turn.me { max-width: 92%; }
    .when { opacity: 1; }
    .stop .lbl { display: none; }
  }
</style>
