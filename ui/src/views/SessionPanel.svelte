<script>
  import { url, go, closeSession } from '../lib/url.svelte.js';
  import { resource, get } from '../lib/api.svelte.js';
  import { chatLabel } from '../lib/format.js';
  import { panel, setWidth, resetWidth } from '../lib/panel.svelte.js';
  import Chat from './Chat.svelte';
  import X from '@lucide/svelte/icons/x';
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';

  /**
   * Chat for one session, beside the page on desktop and full screen on a
   * phone. Live sessions in a pool that allows it take messages; closed
   * ones show their history read-only.
   * @type {{ id: string }}
   */
  let { id } = $props();

  const sessions = resource(() => 'api/sessions');
  const chats = resource(() => 'api/chat');
  const info = $derived(sessions.data?.find((/** @type {any} */ s) => s.id === id));
  const chat = $derived(chats.data?.sessions?.find((/** @type {any} */ s) => s.id === id));
  const key = $derived(chat?.alive ? chat.key : '');
  const title = $derived(chat?.key ? chatLabel(chat.key) : 'Session');

  let access = $state({ write: false, reason: '' });
  $effect(() => {
    access = { write: false, reason: '' };
    let stale = false;
    get('api/session/' + encodeURIComponent(id) + '/access').then(
      (a) => { if (!stale) access = { write: !!a.write, reason: a.reason || '' }; },
      () => {},
    );
    return () => { stale = true; };
  });

  // Phone: the panel covers the page, so the page behind must not scroll.
  $effect(() => {
    if (!panel.mobile) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  });

  /** @type {HTMLElement | undefined} */
  let root = $state();
  $effect(() => { root?.focus({ preventScroll: true }); });

  /** @param {PointerEvent} e */
  function drag(e) {
    const grip = /** @type {HTMLElement} */ (e.currentTarget);
    grip.setPointerCapture(e.pointerId);
    /** @param {PointerEvent} m */
    const move = (m) => setWidth(window.innerWidth - m.clientX);
    const stop = () => { grip.removeEventListener('pointermove', move); grip.removeEventListener('pointerup', stop); grip.removeEventListener('pointercancel', stop); };
    grip.addEventListener('pointermove', move);
    grip.addEventListener('pointerup', stop);
    grip.addEventListener('pointercancel', stop);
  }

  /** @param {KeyboardEvent} e */
  function nudge(e) {
    if (e.key === 'ArrowLeft') { e.preventDefault(); setWidth(panel.w + 24); }
    else if (e.key === 'ArrowRight') { e.preventDefault(); setWidth(panel.w - 24); }
  }
</script>

<aside class="panel" bind:this={root} tabindex="-1" data-testid="session-panel" aria-label="session chat">
  {#if !panel.mobile}
    <div
      class="grip"
      role="slider"
      aria-orientation="horizontal"
      aria-label="chat panel width"
      aria-valuemin="300"
      aria-valuemax="760"
      aria-valuenow={panel.w}
      tabindex="0"
      onpointerdown={drag}
      ondblclick={resetWidth}
      onkeydown={nudge}
    ></div>
  {/if}
  <header>
    <button class="ghost icon close" onclick={closeSession} aria-label="Close chat" title="Close (Esc)" data-testid="session-panel-close">
      {#if panel.mobile}<ArrowLeft size={18} />{:else}<X size={14} />{/if}
    </button>
    <div class="who">
      <div class="title">{title}</div>
      <div class="meta faint">
        <span class="mono">{id}</span>
        {#if info?.ticket?.pool}<span>· {info.ticket.pool}</span>{/if}
        {#if info?.current}
          · <a href={'?ticket=' + info.current} class="mono" onclick={(e) => { e.preventDefault(); go({ ticket: info.current }); }}>{info.current}</a>
        {/if}
      </div>
    </div>
    {#if info}<span class="pill" class:ok={info.alive} class:bad={!info.alive}>{info.alive ? 'live' : 'dead'}</span>{/if}
  </header>
  <div class="body">
    {#key id}
      <Chat {id} {key} write={access.write} reason={access.reason} />
    {/key}
  </div>
</aside>

<style>
  .panel {
    grid-area: panel; position: relative; min-width: 0; min-height: 0; display: flex; flex-direction: column;
    background: var(--panel); border-left: 1px solid var(--border); outline: none;
  }
  .grip { position: absolute; left: -3px; top: 0; bottom: 0; width: 7px; cursor: col-resize; z-index: 3; touch-action: none; }
  .grip:hover, .grip:focus-visible { background: var(--accent-soft); outline: none; }
  header { flex: none; display: flex; align-items: center; gap: 8px; min-height: 44px; padding: 0 10px 0 8px; border-bottom: 1px solid var(--border); }
  .who { flex: 1; min-width: 0; }
  .title { font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .meta { font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .body { flex: 1; min-height: 0; background: var(--bg); }

  @media (max-width: 768px) {
    .panel {
      position: fixed; left: 0; right: 0; top: var(--vvtop, 0px); height: var(--vvh, 100dvh); z-index: 20;
      border-left: 0; padding-top: env(safe-area-inset-top); padding-bottom: env(safe-area-inset-bottom);
      padding-left: env(safe-area-inset-left); padding-right: env(safe-area-inset-right);
    }
    header { min-height: 52px; padding: 0 8px 0 4px; }
    .title { font-size: 15px; }
    .meta { font-size: 12px; }
  }
</style>
