<script>
  // The session console (#535): a tabbed strip along the bottom of the
  // page column on desktop, full screen with the tab bar on top on a
  // phone. Each tab is a session; the active one's body is the `session`
  // slot (the built-in chat). Minimizing keeps the tabs; closing a tab
  // drops it.
  import { slot, resource } from '../../kit/plugin.svelte.js';
  import { panel, tabs, setHeight, resetHeight, show, closeTab, track } from './panel.svelte.js';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ChevronUp from '@lucide/svelte/icons/chevron-up';
  import X from '@lucide/svelte/icons/x';

  /** @type {{ api: any, url: Record<string, string> }} */
  let { api, url } = $props();

  const active = $derived(url.session || '');
  const item = $derived(url.item || '');
  const readonly = $derived(url.ro === '1');
  // Opening a session anywhere (a link, api.openSession) adds its tab.
  $effect(() => { track({ session: active, item, ro: url.ro || '' }); });

  // The last session contribution wins, so a user plugin can replace the chat.
  /** @type {{ plugin: string, element: string } | undefined} */
  const body = $derived(api.contributions('session').at(-1));

  // Tab labels: the pool, the id until the list arrives.
  const list = resource(() => api, () => (tabs.list.length ? 'api/sessions?since=720h' : null));
  const byId = $derived(Object.fromEntries((list.data ?? []).map((/** @type {any} */ s) => [s.id, s])));
  /** @param {string} id */
  const label = (id) => { const s = byId[id]; return s?.ticket?.pool || s?.display_name || id; };
  /** @param {string} id */
  const tip = (id) => { const s = byId[id]; return [s?.display_name || s?.ticket?.title, s?.model, id].filter(Boolean).join('\n'); };

  // Phone: the console covers the page, so the page behind must not scroll.
  $effect(() => {
    if (!panel.mobile || !active) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  });

  /** @type {HTMLElement | undefined} */
  let root = $state();
  $effect(() => { if (active) root?.focus({ preventScroll: true }); });

  /** @param {PointerEvent} e */
  function drag(e) {
    const grip = /** @type {HTMLElement} */ (e.currentTarget);
    grip.setPointerCapture(e.pointerId);
    const bottom = /** @type {HTMLElement} */ (root).getBoundingClientRect().bottom;
    /** @param {PointerEvent} m */
    const move = (m) => setHeight(bottom - m.clientY);
    const stop = () => { grip.removeEventListener('pointermove', move); grip.removeEventListener('pointerup', stop); grip.removeEventListener('pointercancel', stop); };
    grip.addEventListener('pointermove', move);
    grip.addEventListener('pointerup', stop);
    grip.addEventListener('pointercancel', stop);
  }

  /** @param {KeyboardEvent} e */
  function nudge(e) {
    if (e.key === 'ArrowUp') { e.preventDefault(); setHeight(panel.h + 24); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); setHeight(panel.h - 24); }
  }
</script>

{#if active || (tabs.list.length && !panel.mobile)}
  <section class="console" class:open={!!active} bind:this={root} tabindex="-1" aria-label="sessions console" data-testid={active ? 'session-panel' : 'console'} style:--ch={panel.h + 'px'}>
    {#if active && !panel.mobile}
      <div
        class="grip"
        role="slider"
        aria-orientation="vertical"
        aria-label="console height"
        aria-valuemin="160"
        aria-valuenow={panel.h}
        tabindex="0"
        onpointerdown={drag}
        ondblclick={resetHeight}
        onkeydown={nudge}
      ></div>
    {/if}
    <div class="bar">
      <div class="tabs" role="tablist">
        {#each tabs.list as t (t.id)}
          <div class="tab" class:on={t.id === active} role="tab" aria-selected={t.id === active} data-testid="console-tab" data-session={t.id} title={tip(t.id)}>
            <button type="button" class="tl" onclick={() => (t.id === active && !panel.mobile ? api.closeSession() : show(api, t))}>{label(t.id)}</button>
            <button type="button" class="tx" onclick={() => closeTab(api, active, t.id)} aria-label={'Close tab ' + label(t.id)} data-testid="console-tab-close"><X size={12} /></button>
          </div>
        {/each}
      </div>
      {#if active}
        <button type="button" class="ic" onclick={() => api.closeSession()} aria-label="Minimize" title="Minimize (Esc)" data-testid="session-panel-close"><ChevronDown size={16} /></button>
      {:else}
        <button type="button" class="ic" onclick={() => show(api, tabs.list.at(-1) ?? tabs.list[0])} aria-label="Restore" title="Restore" data-testid="console-restore"><ChevronUp size={16} /></button>
      {/if}
    </div>
    {#if active}
      <div class="body">
        {#if body}
          {#key active + item + readonly}
            <div class="plugin-host" data-plugin={body.plugin} use:slot={{ api, slot: 'session', target: { plugin: body.plugin, element: body.element }, context: { session: active, item, readonly } }}></div>
          {/key}
        {:else}
          <p class="faint none">No plugin shows sessions (slot <span class="mono">session</span>).</p>
        {/if}
      </div>
    {/if}
  </section>
{/if}

<style>
  .console {
    flex: none; position: relative; display: flex; flex-direction: column; min-width: 0; min-height: 0;
    background: var(--bg); border-top: 1px solid var(--border); outline: none;
  }
  .console.open { height: var(--ch); }
  .grip { position: absolute; left: 0; right: 0; top: -3px; height: 7px; cursor: row-resize; z-index: 3; touch-action: none; }
  .grip:hover, .grip:focus-visible { background: var(--accent-soft); outline: none; }
  .bar { flex: none; display: flex; align-items: stretch; gap: 4px; height: 32px; padding: 0 4px 0 0; background: var(--side); border-bottom: 1px solid var(--border); }
  .console:not(.open) .bar { border-bottom: 0; }
  .tabs { flex: 1; min-width: 0; display: flex; overflow-x: auto; scrollbar-width: none; }
  .tab { flex: 0 1 auto; min-width: 72px; max-width: 200px; display: flex; align-items: center; border-right: 1px solid var(--border); color: var(--muted); }
  .tab.on { background: var(--bg); color: var(--text); box-shadow: inset 0 2px 0 var(--accent); }
  .tl { flex: 1; min-width: 0; height: 100%; padding: 0 4px 0 10px; border: 0; background: none; color: inherit; font-size: 12px; text-align: left; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; cursor: pointer; }
  .tx, .ic { flex: none; display: inline-flex; align-items: center; justify-content: center; border: 0; background: none; color: var(--faint); cursor: pointer; border-radius: var(--radius); }
  .tx { width: 22px; height: 22px; margin-right: 4px; }
  .ic { width: 32px; }
  .tx:hover, .ic:hover { background: var(--hover); color: var(--text); }
  .body { flex: 1; min-height: 0; display: flex; flex-direction: column; }
  .body > :global(.plugin-host) { flex: 1; min-height: 0; display: flex; flex-direction: column; }
  .none { padding: 16px; }

  @media (max-width: 768px) {
    .console.open {
      position: fixed; left: 0; right: 0; top: var(--vvtop, 0px); height: var(--vvh, 100dvh); z-index: 20; border-top: 0;
      padding-top: env(safe-area-inset-top); padding-left: env(safe-area-inset-left); padding-right: env(safe-area-inset-right);
    }
    .bar { height: 44px; }
    .tab { min-width: 96px; }
    .tx { width: 36px; height: 36px; }
    .ic { width: 44px; }
  }
</style>
