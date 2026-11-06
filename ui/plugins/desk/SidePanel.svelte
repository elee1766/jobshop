<script>
  // The session panel frame (#535): a resizable column beside the page
  // on desktop, full screen on a phone. Its body is the `session` slot;
  // the built-in chat plugin fills it. The frame owns only width, focus,
  // scroll lock and Escape.
  import { slot } from '../../kit/plugin.svelte.js';
  import { panel, setWidth, resetWidth } from './panel.svelte.js';

  /** @type {{ api: any, id: string, item: string, readonly: boolean }} */
  let { api, id, item, readonly } = $props();

  // The last session contribution wins, so a user plugin can replace the chat.
  /** @type {{ plugin: string, element: string } | undefined} */
  const body = $derived(api.contributions('session').at(-1));

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

<aside class="panel" bind:this={root} tabindex="-1" data-testid="session-panel" aria-label="session">
  {#if !panel.mobile}
    <div
      class="grip"
      role="slider"
      aria-orientation="horizontal"
      aria-label="panel width"
      aria-valuemin="300"
      aria-valuemax="760"
      aria-valuenow={panel.w}
      tabindex="0"
      onpointerdown={drag}
      ondblclick={resetWidth}
      onkeydown={nudge}
    ></div>
  {/if}
  {#if body}
    {#key id + item + readonly}
      <div class="plugin-host" data-plugin={body.plugin} use:slot={{ api, slot: 'session', target: { plugin: body.plugin, element: body.element }, context: { session: id, item, readonly } }}></div>
    {/key}
  {:else}
    <p class="faint none">No plugin shows sessions (slot <span class="mono">session</span>).</p>
  {/if}
</aside>

<style>
  .panel {
    grid-area: panel; position: relative; min-width: 0; min-height: 0; display: flex; flex-direction: column;
    background: var(--bg); border-left: 1px solid var(--border); outline: none;
  }
  .panel > :global(.plugin-host) { flex: 1; min-height: 0; display: flex; flex-direction: column; }
  .grip { position: absolute; left: -3px; top: 0; bottom: 0; width: 7px; cursor: col-resize; z-index: 3; touch-action: none; }
  .grip:hover, .grip:focus-visible { background: var(--accent-soft); outline: none; }
  .none { padding: 16px; }

  @media (max-width: 768px) {
    .panel {
      position: fixed; left: 0; right: 0; top: var(--vvtop, 0px); height: var(--vvh, 100dvh); z-index: 20;
      border-left: 0; padding-top: env(safe-area-inset-top);
      padding-left: env(safe-area-inset-left); padding-right: env(safe-area-inset-right);
    }
  }
</style>
