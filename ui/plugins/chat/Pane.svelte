<script>
  // The session chat (#535), after jcode-web's Pane.svelte: header,
  // scrolling transcript with "Load older", a working line, and a
  // composer fixed at the bottom. Opened for an attention item
  // (context.item, "kind:id") it shows the item above the composer with
  // its resolve actions. Talks to Desk only through api (#517).
  import { tick } from 'svelte';
  import { Transcript } from './transcript.svelte.js';
  import { promptHistory, rememberPrompt, tokens } from './format.js';
  import Block from './Block.svelte';
  import { resolveActions, runAction, markPending } from '../../kit/attention.js';
  import { slots } from '../../kit/plugin.svelte.js';

  /** @type {{ api: any, context?: any, config?: any }} */
  let { api, context = {} } = $props();

  const id = $derived(String(context.session || ''));
  const item = $derived(String(context.item || ''));
  const itemKind = $derived(item.split(':')[0] || '');
  const itemId = $derived(item.slice(itemKind.length + 1));
  // A worker opened from a run card or trace (JOB-54): watch only.
  const readonly = $derived(!!context.readonly);

  /** @type {Transcript | null} */
  let t = $state(/** @type {Transcript | null} */ (null));
  $effect(() => {
    if (!id) return;
    const tr = new Transcript(api, id);
    t = tr;
    tr.open();
    return () => tr.close();
  });

  /** @type {any} */
  let info = $state(null);
  let chatKey = $state('');
  let access = $state({ write: false, reason: '' });
  $effect(() => {
    if (!id) return;
    let stale = false;
    info = null;
    chatKey = '';
    access = { write: false, reason: '' };
    api.get('api/sessions?since=720h').then((/** @type {any[]} */ l) => { if (!stale) info = l.find((s) => s.id === id) || null; }).catch(() => {});
    api.get('api/chat').then((/** @type {any} */ c) => {
      if (stale) return;
      const s = (c.sessions || []).find((/** @type {any} */ x) => x.id === id);
      chatKey = s?.alive ? s.key : '';
    }).catch(() => {});
    api.session.access(id).then((/** @type {any} */ a) => { if (!stale) access = { write: !!a.write, reason: a.reason || '' }; }).catch(() => {});
    return () => { stale = true; };
  });

  /** The attention item: its ticket when it is one, for title and actions. @type {any} */
  let ticket = $state(null);
  $effect(() => {
    ticket = null;
    if (!itemId || ['pool', 'deploy', 'rollback', 'config', 'config-repo', 'ingest', 'resource'].includes(itemKind)) return;
    let stale = false;
    api.get('api/ticket/' + encodeURIComponent(itemId)).then((/** @type {any} */ v) => { if (!stale) ticket = v.ticket; }).catch(() => {});
    return () => { stale = true; };
  });

  // A closed session never shows as working (JOB-52): its end state shows.
  const closed = $derived(!!t?.ended);
  const working = $derived(!!t?.busy && !closed);
  /** The running tool's name for the working line. */
  const runningTool = $derived.by(() => {
    const bs = t?.blocks ?? [];
    for (let i = bs.length - 1; i >= 0; i--) {
      const b = bs[i];
      if (b.kind === 'tool' && (b.state === 'running' || b.state === 'input')) return b.name;
    }
    return '';
  });
  const canSend = $derived(access.write || !!chatKey);
  // chat:composer-actions (JOB-87): buttons other plugins put in the
  // composer's action row.
  const actionsCtx = $derived({ session: id, item, working });
  const title = $derived(chatKey.startsWith('escalation:') ? 'Escalation ' + chatKey.slice(11) : chatKey.startsWith('attention:') ? 'Resolve ' + item : chatKey.startsWith('desk:') ? 'Desk chat' : info?.ticket?.title || 'Session');

  let draft = $state('');
  let sending = $state(false);
  let notice = $state('');
  // Sent messages not yet in the transcript (#603): shown at once with
  // their state, dropped when the turn that delivers them starts.
  /** @type {{ text: string, state: string }[]} */
  let pending = $state([]);
  $effect(() => {
    const users = (t?.blocks ?? []).filter((b) => b.kind === 'user').map((b) => /** @type {any} */ (b).text || '');
    if (pending.length && pending.some((p) => users.some((u) => u.includes(p.text)))) {
      pending = pending.filter((p) => !users.some((u) => u.includes(p.text)));
    }
  });
  /** @type {HTMLElement | undefined} */
  let scroller = $state();
  let pinned = $state(true);
  /** @type {HTMLTextAreaElement | undefined} */
  let composer = $state();
  let historyIndex = $state(-1);
  /** @type {string[]} */
  let recalled = [];
  let savedDraft = '';
  /** @type {{ height: number, top: number } | null} */
  let anchor = null;

  /** @param {boolean} older */
  function recall(older) {
    if (historyIndex < 0) {
      if (!older) return;
      recalled = promptHistory(id, t?.prompts() ?? []);
      savedDraft = draft;
    }
    const next = older ? Math.min(historyIndex + 1, recalled.length - 1) : historyIndex - 1;
    if (next < -1 || !recalled.length) return;
    historyIndex = next;
    draft = next < 0 ? savedDraft : recalled[next];
    tick().then(() => { composer?.focus(); composer?.setSelectionRange(draft.length, draft.length); grow(); });
  }

  async function loadOlder() {
    if (!t || !scroller) return;
    pinned = false;
    anchor = { height: scroller.scrollHeight, top: scroller.scrollTop };
    await t.loadOlder();
    await tick();
    if (anchor && scroller) {
      scroller.scrollTop = anchor.top + scroller.scrollHeight - anchor.height;
      anchor = null;
    }
  }

  function onScroll() {
    if (!scroller) return;
    pinned = scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 80;
  }

  // Follow output while pinned to the bottom.
  $effect(() => {
    void t?.version;
    void working;
    if (pinned && !anchor) tick().then(() => { if (scroller) scroller.scrollTop = scroller.scrollHeight; });
  });

  function grow() {
    if (!composer) return;
    composer.style.height = 'auto';
    composer.style.height = Math.min(composer.scrollHeight, Math.round(window.innerHeight * 0.4)) + 'px';
  }

  /** @param {boolean} interrupt */
  async function submit(interrupt) {
    const text = draft.trim();
    if (!text || sending) return;
    sending = true;
    notice = '';
    try {
      // A question chat talks to the chat agent too, which relays the
      // answer with question.answer and follows the worker (JOB-52).
      await api.session.send(chatKey ? { key: chatKey, text, interrupt } : { session: id, text, interrupt });
      pending = [...pending, { text, state: working && !interrupt ? 'queued for after this turn' : 'sent, waiting for the session' }];
      if (working && !interrupt) notice = 'Queued for after this turn';
      rememberPrompt(id, text);
      historyIndex = -1;
      draft = '';
      pinned = true;
      await tick();
      grow();
    } catch (e) {
      notice = String(/** @type {any} */ (e).message || e);
    } finally {
      sending = false;
    }
  }

  async function stop() {
    notice = '';
    try {
      const r = await api.session.cancel(id);
      if (!r.cancelled) notice = 'No turn running';
    } catch (e) {
      notice = String(/** @type {any} */ (e).message || e);
    }
  }

  /** @param {KeyboardEvent} e */
  function onKey(e) {
    if (e.isComposing) return;
    const input = /** @type {HTMLTextAreaElement} */ (e.currentTarget);
    if (e.key === 'ArrowUp' && !e.shiftKey && input.selectionStart === input.selectionEnd && (historyIndex >= 0 || !draft.slice(0, input.selectionStart).includes('\n'))) {
      e.preventDefault(); recall(true);
    } else if (e.key === 'ArrowDown' && !e.shiftKey && historyIndex >= 0) {
      e.preventDefault(); recall(false);
    } else if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      submit(e.ctrlKey || e.metaKey);
    } else if (e.key === 'Escape' && working) {
      e.preventDefault();
      e.stopPropagation();
      stop();
    }
  }

  // Resolve actions per item kind, shared with the Attention rows (#607).
  const actions = $derived(resolveActions(api, itemKind, itemId));
  let acting = $state(false);
  /** A success marks the item pending in the Attention list (JOB-60),
   * closes the panel and returns to the list (#607); a failure keeps it
   * open with the error toast.
   * @param {import('../../kit/attention.js').Action} a */
  async function act(a) {
    acting = true;
    try {
      if (await runAction(api, a)) {
        markPending(itemKind + ':' + itemId, a.label);
        api.closeSession();
      }
    } finally {
      acting = false;
    }
  }
</script>

<section class="pane" data-testid="chat-pane" aria-label={title}>
  <header class="head">
    <div class="who">
      <div class="name">{title}</div>
      <div class="meta">
        <span class="mono">{id}</span>
        {#if info?.ticket?.pool}<span>· {info.ticket.pool}</span>{/if}
        {#if info?.current}· <a href={'?ticket=' + info.current} class="mono" onclick={(e) => { e.preventDefault(); api.openTicket(info.current); }}>{info.current}</a>{/if}
      </div>
    </div>
    <span class="state" data-testid="chat-status">{t?.status ?? 'connecting'}</span>
  </header>

  <div class="scroll" bind:this={scroller} onscroll={onScroll} data-testid="journal">
    {#if t && t.earlier > 0}
      <button class="older" disabled={t.loadingOlder} onclick={loadOlder} data-testid="load-older">{t.loadingOlder ? 'Loading history…' : `Load older messages (${t.earlier} earlier)`}</button>
    {/if}
    {#if t?.error}<p class="err">{t.error}</p>{/if}
    {#each t?.blocks ?? [] as b, i (i)}
      <Block block={b} />
    {:else}
      <p class="none">No messages yet.</p>
    {/each}
    {#each pending as p, i (i)}
      <div class="pending" data-testid="chat-pending"><div class="ptext">{p.text}</div><span class="pstate">{p.state}</span></div>
    {/each}
    {#if working}
      <p class="status" data-testid="chat-working"><span class="spin"></span>{runningTool || 'Working'}</p>
    {/if}
  </div>

  {#if item}
    <div class="item" data-testid="chat-item">
      <div class="itxt"><span class="pill">{itemKind}</span> <span class="ititle">{ticket?.title || itemId}</span></div>
      <div class="iacts">
        {#each actions as a (a.label)}
          <button type="button" disabled={acting} onclick={() => act(a)} data-testid="resolve">{a.label}</button>
        {/each}
      </div>
    </div>
  {/if}

  {#if readonly}
    <p class="ro" data-testid="chat-readonly">Read-only: a worker's turn is not driven from here.{#if info?.current} <a href={'?ticket=' + info.current} class="mono" onclick={(e) => { e.preventDefault(); api.openTicket(info.current); }} data-testid="chat-ro-ticket">{info.current}</a>{/if}</p>
  {:else if canSend && !closed}
    <form class="composer" onsubmit={(e) => { e.preventDefault(); submit(false); }}>
      <textarea
        rows="2"
        bind:value={draft}
        bind:this={composer}
        aria-label="Message"
        oninput={() => { historyIndex = -1; grow(); }}
        onkeydown={onKey}
        placeholder={working ? 'Queued for after this turn' : 'Message'}
        enterkeyhint="send"
        data-testid="chat-input"
      ></textarea>
      <div class="actions">
        <button type="button" class="history" onclick={() => recall(true)} aria-label="Previous sent message" title="Previous sent message (Up arrow)">↑</button>
        <button type="button" class="history" onclick={() => recall(false)} disabled={historyIndex < 0} aria-label="Next sent message" title="Next sent message (Down arrow)">↓</button>
        {#if notice}<span class="notice" data-testid="chat-notice">{notice}</span>{/if}
        <span class="grow"></span>
        <span class="plugged" data-testid="chat-composer-actions" use:slots={{ api, slot: 'chat:composer-actions', context: actionsCtx }}></span>
        {#if working}
          <button type="button" class="stop" onclick={stop} title="Cancel the turn (Esc)" data-testid="chat-stop">Stop</button>
          <button type="button" class="interrupt" disabled={!draft.trim() || sending} onclick={() => submit(true)} title="Cancel the turn and send now (Ctrl+Enter)" data-testid="chat-interrupt">Interrupt</button>
        {/if}
        <button type="submit" class="send" disabled={!draft.trim() || sending} data-testid="chat-send">{working ? 'Queue' : 'Send'}</button>
      </div>
    </form>
  {:else if closed}
    <p class="ro" data-testid="chat-closed">Session closed{t?.ended && t.ended !== 'closed' ? ' · ' + t.ended : ''}</p>
  {:else}
    <p class="ro" data-testid="chat-readonly">Read-only: {access.reason || 'this session does not accept input'}</p>
  {/if}
  <footer class="foot" data-testid="chat-usage">
    {#if info?.model}<span class="model" data-testid="chat-model">{info.model}{info.effort ? ' · ' + info.effort : ''}</span>{/if}
    {#if t?.usage}<span class="usage">{tokens(t.usage.used)} / {tokens(t.usage.size)} tokens</span>{/if}
    <span class="grow"></span>
    <span class="mono fid">{id}</span>
  </footer>
</section>

<style>
  .pane { display: flex; flex-direction: column; height: 100%; min-height: 0; min-width: 0; background: var(--desk-bg); color: var(--desk-text); font: 13px/1.5 var(--desk-font); }
  .head { flex: none; display: flex; align-items: center; gap: 8px; min-height: 40px; padding: 0 12px; border-bottom: 1px solid var(--desk-border); background: var(--desk-panel); }
  .who { flex: 1; min-width: 0; }
  .name { font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .meta { font-size: 11px; color: var(--desk-faint); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .meta a { color: var(--desk-accent); }
  .mono { font-family: var(--desk-mono); }
  .state { font-size: 11px; color: var(--desk-faint); flex: none; }
  .scroll { flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden; overscroll-behavior: contain; padding: 12px 16px 16px; }
  .older { font-size: 12px; margin: 0 0 12px; padding: 4px 8px; border: 1px solid var(--desk-border); border-radius: 4px; background: none; color: var(--desk-faint); cursor: pointer; }
  .none { color: var(--desk-faint); text-align: center; padding: 24px 0; margin: 0; }
  .pending { margin: 8px 0 0 auto; max-width: 85%; padding: 8px 10px; border-radius: var(--desk-radius); border: 1px dashed var(--desk-border); background: var(--desk-raised); }
  .ptext { white-space: pre-wrap; overflow-wrap: anywhere; }
  .pstate { display: block; margin-top: 4px; font-size: 11px; color: var(--desk-faint); }
  .err { color: var(--desk-bad); font-size: 12px; }
  .status { display: flex; align-items: center; gap: 8px; color: var(--desk-accent); font-size: 12px; margin: 8px 0 0; overflow-wrap: anywhere; }
  .spin { flex: none; width: 10px; height: 10px; border: 2px solid var(--desk-accent); border-right-color: transparent; border-radius: 50%; animation: spin 0.9s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  @media (prefers-reduced-motion: reduce) { .spin { animation: none; } }
  .item { flex: none; display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; padding: 8px 12px; border-top: 1px solid var(--desk-border); background: var(--desk-panel); }
  .itxt { flex: 1 1 200px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .pill { display: inline-flex; align-items: center; height: 18px; padding: 0 7px; border-radius: 999px; font-size: 11px; border: 1px solid var(--desk-border); color: var(--desk-warn); }
  .iacts { display: flex; flex-wrap: wrap; gap: 6px; align-items: center; }
  .iacts button, .actions button { height: 28px; padding: 0 10px; font-size: 12px; border-radius: var(--desk-radius); border: 1px solid var(--desk-border); background: var(--desk-raised); color: var(--desk-text); cursor: pointer; white-space: nowrap; }
  .iacts button:disabled, .actions button:disabled { opacity: 0.45; cursor: default; }
  .composer { flex: none; border-top: 1px solid var(--desk-border); padding: 8px 12px; background: var(--desk-panel); }
  textarea { display: block; width: 100%; resize: none; padding: 8px 10px; border: 1px solid var(--desk-border); border-radius: 6px; background: var(--desk-bg); color: var(--desk-text); font: inherit; max-height: 40vh; box-sizing: border-box; }
  textarea:focus { outline: none; border-color: var(--desk-accent); }
  .actions { display: flex; align-items: center; gap: 6px; margin-top: 6px; min-width: 0; }
  .actions > button { flex: none; }
  .actions .history { width: 28px; padding: 0; background: none; border-color: transparent; color: var(--desk-muted); }
  .foot { flex: none; display: flex; gap: 10px; align-items: center; min-width: 0; height: 22px; padding: 0 12px calc(env(safe-area-inset-bottom)); border-top: 1px solid var(--desk-border); background: var(--desk-panel); color: var(--desk-faint); font-size: 11px; white-space: nowrap; overflow: hidden; }
  .fid { min-width: 0; overflow: hidden; text-overflow: ellipsis; }
  .model { min-width: 0; overflow: hidden; text-overflow: ellipsis; }
  .usage { flex: none; }
  .grow { flex: 1; }
  .plugged { display: flex; align-items: center; gap: 6px; flex: none; }
  .plugged:empty { display: none; }
  .notice { font-size: 11px; color: var(--desk-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
  .actions .stop { border-color: var(--desk-bad); color: var(--desk-bad); background: none; }
  .actions .send { background: var(--desk-accent); color: var(--desk-accent-fg); border-color: transparent; }
  .ro { flex: none; margin: 0; padding: 12px; border-top: 1px solid var(--desk-border); color: var(--desk-faint); font-size: 12px; text-align: center; }
  @media (max-width: 768px) {
    .foot { height: calc(24px + env(safe-area-inset-bottom)); }
    .scroll { padding: 10px 12px 12px; }
    textarea { font-size: 16px; }
    .iacts button, .actions button { height: 40px; }
    .actions .history { width: 36px; }
  }
</style>
